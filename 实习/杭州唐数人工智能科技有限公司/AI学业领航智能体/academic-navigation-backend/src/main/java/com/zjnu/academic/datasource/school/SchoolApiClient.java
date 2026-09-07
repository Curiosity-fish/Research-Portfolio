package com.zjnu.academic.datasource.school;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.config.SchoolApiProperties;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

@Component
public class SchoolApiClient {

    private final SchoolApiProperties properties;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    public SchoolApiClient(SchoolApiProperties properties) {
        this.properties = properties;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofMillis(properties.getConnectTimeout()))
                .build();
        this.objectMapper = new ObjectMapper()
                .setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
    }

    public JsonNode getData(String path) {
        return getData(path, Map.of());
    }

    public JsonNode getData(String path, Map<String, String> params) {
        String url = buildUrl(path, params);
        HttpRequest.Builder builder = HttpRequest.newBuilder(URI.create(url))
                .timeout(Duration.ofMillis(properties.getReadTimeout()))
                .GET();
        if (StringUtils.hasText(properties.getToken())) {
            builder.header("Authorization", "Bearer " + properties.getToken());
        }
        try {
            HttpResponse<String> response = httpClient.send(builder.build(), HttpResponse.BodyHandlers.ofString());
            if (response.statusCode() != 200) {
                throw new BusinessException(502, "教务接口调用失败: HTTP " + response.statusCode());
            }
            JsonNode root = objectMapper.readTree(response.body());
            int code = root.path("code").asInt(-1);
            if (code != 0) {
                throw new BusinessException(502, "教务接口返回错误: " + root.path("message").asText());
            }
            return root.path("data");
        } catch (BusinessException e) {
            throw e;
        } catch (IOException | InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new BusinessException(502, "教务接口请求失败: " + e.getMessage());
        }
    }

    public List<JsonNode> list(String path) {
        JsonNode data = getData(path);
        List<JsonNode> result = new ArrayList<>();
        if (data != null && data.isArray()) {
            data.forEach(result::add);
        }
        return result;
    }

    public JsonNode fetchStudents(int page, int size) {
        Map<String, String> params = new LinkedHashMap<>();
        params.put("page", String.valueOf(page));
        params.put("size", String.valueOf(size));
        return getData("/api/v1/students", params);
    }

    public List<JsonNode> fetchStudentGrades(String studentId, String term) {
        String path = "/api/v1/students/" + studentId + "/grades";
        return term == null || term.isBlank()
                ? list(path)
                : listWithParam(path, "term", term);
    }

    public List<JsonNode> fetchStudentGpaHistory(String studentId) {
        return list("/api/v1/students/" + studentId + "/gpa-history");
    }

    public List<JsonNode> fetchStudentProfileScores(String studentId, String term) {
        String path = "/api/v1/students/" + studentId + "/profile-scores";
        return term == null || term.isBlank()
                ? list(path)
                : listWithParam(path, "term", term);
    }

    public List<JsonNode> fetchStudentHealthReports(String studentId) {
        return list("/api/v1/students/" + studentId + "/health-reports");
    }

    public List<JsonNode> fetchStudentAlerts(String studentId) {
        return list("/api/v1/students/" + studentId + "/alerts");
    }

    public JsonNode fetchStudent(String studentId) {
        return getData("/api/v1/students/" + studentId);
    }

    private List<JsonNode> listWithParam(String path, String key, String value) {
        Map<String, String> params = new LinkedHashMap<>();
        params.put(key, value);
        JsonNode data = getData(path, params);
        List<JsonNode> result = new ArrayList<>();
        if (data != null && data.isArray()) {
            data.forEach(result::add);
        }
        return result;
    }

    private String buildUrl(String path, Map<String, String> params) {
        StringBuilder url = new StringBuilder(properties.getBaseUrl());
        if (!path.startsWith("/")) {
            url.append('/');
        }
        url.append(path);
        if (!params.isEmpty()) {
            url.append('?');
            params.forEach((k, v) -> url.append(k).append('=').append(v).append('&'));
            url.setLength(url.length() - 1);
        }
        return url.toString();
    }
}
