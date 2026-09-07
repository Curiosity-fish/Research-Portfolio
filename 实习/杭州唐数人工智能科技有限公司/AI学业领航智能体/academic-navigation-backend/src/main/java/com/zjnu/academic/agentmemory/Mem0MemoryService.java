package com.zjnu.academic.agentmemory;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.zjnu.academic.common.exception.BusinessException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.net.URI;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.time.Instant;
import java.time.format.DateTimeParseException;
import java.util.ArrayList;
import java.util.Base64;
import java.util.List;
import java.util.Map;

@Service
public class Mem0MemoryService {

    private static final Logger log = LoggerFactory.getLogger(Mem0MemoryService.class);
    static final String LIFE_PLANNING_AGENT = "life-planning";

    private final AgentMemoryProperties properties;
    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;

    public Mem0MemoryService(AgentMemoryProperties properties, ObjectMapper objectMapper) {
        this.properties = properties;
        this.objectMapper = objectMapper;
        this.httpClient = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(5)).build();
    }

    public boolean isAvailable() {
        AgentMemoryProperties.Mem0 config = properties.getMem0();
        return config.isEnabled()
                && notBlank(config.getBaseUrl())
                && notBlank(config.getApiKey())
                && notBlank(config.getIdentitySecret());
    }

    public MemoryLookupResult search(Long userId, String query) {
        if (!isAvailable() || query == null || query.isBlank()) {
            return new MemoryLookupResult(List.of(), false);
        }
        try {
            JsonNode body = call("POST", "/search", Map.of(
                    "query", query,
                    "filters", Map.of("user_id", scopedUserId(userId), "agent_id", LIFE_PLANNING_AGENT),
                    "top_k", 5,
                    "threshold", 0.65));
            return new MemoryLookupResult(readItems(body.path("results")), true);
        } catch (Exception e) {
            log.warn("Mem0 search unavailable: {}", e.getMessage());
            return new MemoryLookupResult(List.of(), false);
        }
    }

    public boolean remember(Long userId, String text) {
        if (!isAvailable() || !MemorySafetyPolicy.canStore(text)) {
            return false;
        }
        try {
            call("POST", "/memories", Map.of(
                    "messages", List.of(Map.of("role", "user", "content", text)),
                    "user_id", scopedUserId(userId),
                    "agent_id", LIFE_PLANNING_AGENT,
                    "metadata", Map.of(
                            "tenant_id", "zjnu",
                            "sensitivity", "low",
                            "source", "student_confirmed"),
                    "infer", true,
                    "prompt", "Store only explicitly stated low-sensitivity long-term career goals, preferences, constraints, or agreed actions. Never store academic scores, rankings, alerts, health, mental health, identities, credentials, or inferred traits."));
            return true;
        } catch (Exception e) {
            log.warn("Mem0 write unavailable: {}", e.getMessage());
            return false;
        }
    }

    public List<MemoryItemDTO> list(Long userId) {
        if (!isAvailable()) {
            return List.of();
        }
        try {
            String path = "/memories?user_id=" + encode(scopedUserId(userId))
                    + "&agent_id=" + LIFE_PLANNING_AGENT + "&top_k=100";
            return readItems(call("GET", path, null).path("results"));
        } catch (Exception e) {
            log.warn("Mem0 list unavailable: {}", e.getMessage());
            return List.of();
        }
    }

    public void delete(Long userId, String memoryId) {
        if (!isAvailable()) {
            throw new BusinessException(503, "长期记忆服务暂不可用");
        }
        boolean owned = list(userId).stream().anyMatch(item -> item.id().equals(memoryId));
        if (!owned) {
            throw new BusinessException(404, "记忆不存在");
        }
        try {
            call("DELETE", "/memories/" + encode(memoryId), null);
        } catch (Exception e) {
            log.warn("Mem0 delete unavailable: {}", e.getMessage());
            throw new BusinessException(503, "长期记忆服务暂不可用");
        }
    }

    private JsonNode call(String method, String path, Object payload) throws Exception {
        AgentMemoryProperties.Mem0 config = properties.getMem0();
        HttpRequest.Builder request = HttpRequest.newBuilder()
                .uri(URI.create(trimTrailingSlash(config.getBaseUrl()) + path))
                .timeout(Duration.ofSeconds(Math.max(1, config.getTimeoutSeconds())))
                .header("X-API-Key", config.getApiKey())
                .header("Accept", "application/json");
        if (payload != null) {
            request.header("Content-Type", "application/json")
                    .method(method, HttpRequest.BodyPublishers.ofString(objectMapper.writeValueAsString(payload)));
        } else {
            request.method(method, HttpRequest.BodyPublishers.noBody());
        }
        HttpResponse<String> response = httpClient.send(request.build(), HttpResponse.BodyHandlers.ofString());
        if (response.statusCode() < 200 || response.statusCode() >= 300) {
            throw new IllegalStateException("Mem0 HTTP " + response.statusCode());
        }
        return response.body().isBlank() ? objectMapper.createObjectNode() : objectMapper.readTree(response.body());
    }

    private List<MemoryItemDTO> readItems(JsonNode results) {
        List<MemoryItemDTO> items = new ArrayList<>();
        if (!results.isArray()) {
            return items;
        }
        for (JsonNode row : results) {
            String id = row.path("id").asText();
            String memory = row.path("memory").asText();
            if (id.isBlank() || memory.isBlank()) {
                continue;
            }
            Double score = row.hasNonNull("score") ? row.get("score").asDouble() : null;
            items.add(new MemoryItemDTO(id, memory, score, parseInstant(row.path("created_at").asText())));
        }
        return items;
    }

    private String scopedUserId(Long userId) {
        try {
            Mac mac = Mac.getInstance("HmacSHA256");
            mac.init(new SecretKeySpec(properties.getMem0().getIdentitySecret().getBytes(StandardCharsets.UTF_8), "HmacSHA256"));
            byte[] digest = mac.doFinal(("zjnu:student:" + userId).getBytes(StandardCharsets.UTF_8));
            return "student-" + Base64.getUrlEncoder().withoutPadding().encodeToString(digest);
        } catch (Exception e) {
            throw new IllegalStateException("Cannot derive scoped memory identity", e);
        }
    }

    private Instant parseInstant(String value) {
        try {
            return value == null || value.isBlank() ? null : Instant.parse(value);
        } catch (DateTimeParseException ignored) {
            return null;
        }
    }

    private static String trimTrailingSlash(String value) {
        return value.endsWith("/") ? value.substring(0, value.length() - 1) : value;
    }

    private static boolean notBlank(String value) {
        return value != null && !value.isBlank();
    }

    private static String encode(String value) {
        return URLEncoder.encode(value, StandardCharsets.UTF_8);
    }
}
