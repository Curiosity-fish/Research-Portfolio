package com.zjnu.academic.llm;

import java.util.Map;

public record AiInterpretRequest(String instruction, Map<String, Object> data) {
}
