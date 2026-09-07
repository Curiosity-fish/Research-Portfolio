package com.zjnu.academic.student.dto;

import java.util.List;
import java.util.Map;

public record HealthReportDTO(
        String id,
        String studentId,
        String term,
        Integer totalScore,
        List<Map<String, Object>> items,
        String reportUrl
) {
}
