package com.zjnu.academic.llm;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.LinkedHashMap;
import java.util.Map;

@Service
public class AiInterpretService {

    private static final Logger log = LoggerFactory.getLogger(AiInterpretService.class);

    private final LocalLlmClient localLlmClient;
    private final ObjectMapper objectMapper;

    @Autowired
    public AiInterpretService(LocalLlmClient localLlmClient, ObjectMapper objectMapper) {
        this.localLlmClient = localLlmClient;
        this.objectMapper = objectMapper;
    }

    public Map<String, Object> interpret(String instruction, Map<String, Object> metrics) {
        String system = "你是高校学业数据分析助手。只依据提供的指标做中文解读、归因和建议，"
                + "不要自行计算或编造数值，不要输出学号全量明细。"
                + "输出要求：报告简短，不超过200字；不要使用任何标题符号（不要出现#、##、###）；"
                + "不要使用**加粗**或反引号；第一行输出一句总体评价，后续每行用“- ”开头输出一条建议，最多5条。";
        String user = (instruction == null || instruction.isBlank()
                ? "请基于以下指标给出简明分析、问题定位和行动建议：" : instruction)
                + "\n指标JSON:\n" + toJson(metrics);
        try {
            String content = localLlmClient.chat(system, user);
            if (content == null || content.isBlank()) {
                return fallback(metrics);
            }
            return Map.of("content", content, "source", "local");
        } catch (Exception e) {
            log.warn("本地模型不可用，使用规则摘要降级: {}", e.getMessage());
            return fallback(metrics);
        }
    }

    private Map<String, Object> fallback(Map<String, Object> metrics) {
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("content", "本地模型暂未启动，以下为基于指标生成的规则摘要："
                + (metrics == null || metrics.isEmpty() ? "暂无指标数据" : "共 " + metrics.size() + " 项指标，请查看页面数值。"));
        result.put("source", "fallback");
        return result;
    }

    private String toJson(Map<String, Object> metrics) {
        try {
            return objectMapper.writeValueAsString(metrics == null ? Map.of() : metrics);
        } catch (JsonProcessingException e) {
            return String.valueOf(metrics);
        }
    }
}
