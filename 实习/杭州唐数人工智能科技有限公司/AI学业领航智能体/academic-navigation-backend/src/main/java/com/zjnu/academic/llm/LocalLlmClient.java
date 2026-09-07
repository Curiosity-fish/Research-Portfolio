package com.zjnu.academic.llm;

import com.zjnu.academic.config.LlmProperties;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.http.client.JdkClientHttpRequestFactory;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

import java.time.Duration;
import java.util.List;
import java.util.Map;

/**
 * Ollama OpenAI 兼容客户端：POST {baseUrl}/v1/chat/completions
 */
@Service
public class LocalLlmClient {

    private final LlmProperties properties;
    private final RestClient restClient;

    @Autowired
    public LocalLlmClient(LlmProperties properties, RestClient.Builder builder) {
        this.properties = properties;
        JdkClientHttpRequestFactory requestFactory = new JdkClientHttpRequestFactory();
        requestFactory.setReadTimeout(Duration.ofSeconds(properties.getTimeoutSeconds()));
        this.restClient = builder.baseUrl(properties.getBaseUrl())
                .requestFactory(requestFactory)
                .build();
    }

    public String chat(String systemPrompt, String userPrompt) {
        Map<String, Object> body = Map.of(
                "model", properties.getModel(),
                "stream", false,
                "temperature", 0.3,
                "messages", List.of(
                        Map.of("role", "system", "content", systemPrompt),
                        Map.of("role", "user", "content", userPrompt)));
        Map<?, ?> response = restClient.post()
                .uri("/v1/chat/completions")
                .contentType(MediaType.APPLICATION_JSON)
                .body(body)
                .retrieve()
                .body(Map.class);
        List<?> choices = (List<?>) response.get("choices");
        if (choices == null || choices.isEmpty()) {
            return "";
        }
        Map<?, ?> first = (Map<?, ?>) choices.get(0);
        Map<?, ?> message = (Map<?, ?>) first.get("message");
        return message == null ? "" : String.valueOf(message.get("content"));
    }
}
