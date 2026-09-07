package com.zjnu.academic.development.dto;

public record DevelopmentAnalyzeRequest(
        String targetKey,
        String targetLabel,
        Integer targetScore
) {
}
