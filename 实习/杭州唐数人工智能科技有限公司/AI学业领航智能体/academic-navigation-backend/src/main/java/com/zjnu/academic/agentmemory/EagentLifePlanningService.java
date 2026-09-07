package com.zjnu.academic.agentmemory;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.zjnu.academic.common.exception.BusinessException;
import org.springframework.stereotype.Service;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class EagentLifePlanningService {
    private final AgentMemoryProperties properties;
    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;
    private final Map<Long, WorkflowSession> sessions = new ConcurrentHashMap<>();

    public EagentLifePlanningService(AgentMemoryProperties properties, ObjectMapper objectMapper) {
        this.properties = properties;
        this.objectMapper = objectMapper;
        this.httpClient = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(5)).build();
    }

    public String chat(Long userId, String question, List<LifePlanningChatRequest.ChatHistoryItem> history,
                       List<MemoryItemDTO> memories, String currentStudentContext) {
        AgentMemoryProperties.Eagent config = properties.getEagent();
        if (blank(config.getBaseUrl()) || blank(config.getWorkflowId())) {
            throw new BusinessException(503, "人生规划智能体尚未完成后端配置");
        }
        String userInput = buildUserInput(question, history, memories, currentStudentContext);
        WorkflowSession session = sessions.get(userId);
        try {
            InvocationResult result = invoke(config, session, userInput);
            if (result.sessionId() != null) {
                session = new WorkflowSession(result.sessionId(), result.inputNodeId(), result.messageId());
                sessions.put(userId, session);
            }
            if (result.answer().isBlank() && session != null && !blank(session.sessionId())) {
                result = invoke(config, session, userInput);
                if (result.sessionId() != null) {
                    session = new WorkflowSession(result.sessionId(), result.inputNodeId(), result.messageId());
                    sessions.put(userId, session);
                }
            }
            if (result.closed() && blank(result.inputNodeId())) sessions.remove(userId);
            if (result.answer().isBlank()) throw new BusinessException(502, "人生规划智能体未返回有效答复");
            return result.answer();
        } catch (BusinessException e) {
            throw e;
        } catch (Exception e) {
            throw new BusinessException(502, "人生规划智能体暂时不可用");
        }
    }

    private InvocationResult invoke(AgentMemoryProperties.Eagent config, WorkflowSession session, String userInput) throws Exception {
        Map<String, Object> inputValue = Map.of("user_input", userInput);
        Map<String, Object> input = session != null && !blank(session.inputNodeId()) ? Map.of(session.inputNodeId(), inputValue) : inputValue;
        Map<String, Object> body = new java.util.LinkedHashMap<>();
        body.put("workflow_id", config.getWorkflowId()); body.put("stream", true); body.put("input", input);
        if (session != null && !blank(session.sessionId())) body.put("session_id", session.sessionId());
        if (session != null && session.messageId() != null) body.put("message_id", session.messageId());
        HttpRequest.Builder requestBuilder = HttpRequest.newBuilder()
                .uri(URI.create(trimTrailingSlash(config.getBaseUrl()) + "/api/v2/workflow/invoke"))
                .timeout(Duration.ofSeconds(Math.max(1, config.getTimeoutSeconds())))
                .header("Content-Type", "application/json")
                .header("Accept", "text/event-stream");
        if (!blank(config.getApiKey())) requestBuilder.header("X-API-Key", config.getApiKey());
        HttpRequest request = requestBuilder
                .POST(HttpRequest.BodyPublishers.ofString(objectMapper.writeValueAsString(body), StandardCharsets.UTF_8)).build();
        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        if (response.statusCode() == 401 || response.statusCode() == 403) {
            throw new BusinessException(503, "E-Agent 工作流要求服务端 API Key；请由 E-Agent 管理员配置 EAGENT_API_KEY");
        }
        if (response.statusCode() < 200 || response.statusCode() >= 300) throw new BusinessException(502, "人生规划智能体暂时不可用");
        return parseInvocation(response.body());
    }

    private InvocationResult parseInvocation(String body) throws Exception {
        List<JsonNode> events = new ArrayList<>(); String sessionId = null, inputNodeId = null; Long messageId = null; boolean closed = false; StringBuilder answer = new StringBuilder();
        for (String raw : body.split("\\r?\\n")) {
            String line = raw.trim(); if (!line.startsWith("data:")) continue;
            String payload = line.substring(5).trim(); if (payload.isBlank() || "[DONE]".equals(payload)) continue;
            JsonNode sse; try { sse = objectMapper.readTree(payload); } catch (Exception ignored) { continue; }
            JsonNode event = unwrapEvent(sse); events.add(event); sessionId = firstText(sessionId, text(sse, "session_id"), text(event, "session_id"));
            String eventName = firstText(null, text(event, "event"), text(event, "type"), text(sse, "event"));
            if ("input".equals(eventName)) { inputNodeId = firstText(inputNodeId, text(event, "node_id")); if (event.path("message_id").canConvertToLong()) messageId = event.path("message_id").asLong(); }
            if ("close".equals(eventName)) closed = true;
            String chunk = extractText(sse, event, eventName); if (!chunk.isBlank()) appendDistinct(answer, chunk);
        }
        if (events.isEmpty() && body.trim().startsWith("{")) { JsonNode json = objectMapper.readTree(body); sessionId = firstText(sessionId, text(json, "session_id")); String extracted = extractText(json, json, text(json, "event")); if (!extracted.isBlank()) appendDistinct(answer, extracted); }
        return new InvocationResult(sessionId, inputNodeId, messageId, answer.toString().trim(), closed);
    }

    private static void appendDistinct(StringBuilder answer, String chunk) {
        String current = answer.toString();
        if (current.isBlank()) { answer.append(chunk); return; }
        if (chunk.startsWith(current)) { answer.setLength(0); answer.append(chunk); return; }
        if (!current.endsWith(chunk) && !current.contains(chunk)) answer.append(chunk);
    }

    private JsonNode unwrapEvent(JsonNode sse) {
        JsonNode data = sse.path("data");
        if (data.isTextual()) try { data = objectMapper.readTree(data.asText()); } catch (Exception ignored) { }
        if (data.isObject() && data.has("data")) data = data.path("data");
        return data.isObject() ? data : sse;
    }

    private String extractText(JsonNode sse, JsonNode event, String eventName) {
        if ("guide_word".equals(eventName) || "guide_question".equals(eventName) || "input".equals(eventName)) return "";
        List<JsonNode> candidates = List.of(event.path("choices").path(0).path("delta").path("content"), sse.path("choices").path(0).path("delta").path("content"), event.path("delta").path("content"), event.path("output_schema").path("message"), event.path("output").path("message"), event.path("answer"), event.path("content"));
        for (JsonNode candidate : candidates) { String value = nodeText(candidate); if (!value.isBlank()) return value; }
        return "";
    }

    private String nodeText(JsonNode node) {
        if (node == null || node.isMissingNode() || node.isNull()) return "";
        if (node.isTextual()) return node.asText();
        if (node.isArray()) { StringBuilder b = new StringBuilder(); node.forEach(n -> b.append(nodeText(n))); return b.toString(); }
        if (node.isObject()) { String value = nodeText(node.get("content")); if (!value.isBlank()) return value; value = nodeText(node.get("text")); return value.isBlank() ? nodeText(node.get("message")) : value; }
        return "";
    }

    private String buildUserInput(String question, List<LifePlanningChatRequest.ChatHistoryItem> history, List<MemoryItemDTO> memories, String currentStudentContext) {
        StringBuilder b = new StringBuilder("你是高校学生的人生规划助手。给出可执行、不过度承诺的建议，并区分事实与建议。不能做医疗、心理诊断；遇到危机表达时建议联系学校心理中心、辅导员或当地紧急支持资源。\n");
        if (memories != null && !memories.isEmpty()) { b.append("学生确认允许使用的低敏长期记忆（仅作背景）：\n"); memories.forEach(item -> b.append("- ").append(item.memory()).append('\n')); }
        if (currentStudentContext != null && !currentStudentContext.isBlank()) b.append("当前学业摘要（来自学校数据库，缺失时不要编造）：\n").append(currentStudentContext).append('\n');
        if (history != null && !history.isEmpty()) { b.append("最近对话：\n"); history.stream().filter(item -> item != null && item.content() != null && !item.content().isBlank()).limit(16).forEach(item -> b.append(item.role()).append(": ").append(item.content()).append('\n')); }
        b.append("学生本次输入：\n").append(question); return b.toString();
    }

    private static String text(JsonNode node, String field) { JsonNode value = node == null ? null : node.get(field); return value != null && value.isValueNode() ? value.asText(null) : null; }
    private static String firstText(String current, String... values) { if (!blank(current)) return current; for (String value : values) if (!blank(value)) return value; return current; }
    private static boolean blank(String value) { return value == null || value.isBlank(); }
    private static String trimTrailingSlash(String value) { return value.endsWith("/") ? value.substring(0, value.length() - 1) : value; }
    private record WorkflowSession(String sessionId, String inputNodeId, Long messageId) { }
    private record InvocationResult(String sessionId, String inputNodeId, Long messageId, String answer, boolean closed) { }
}
