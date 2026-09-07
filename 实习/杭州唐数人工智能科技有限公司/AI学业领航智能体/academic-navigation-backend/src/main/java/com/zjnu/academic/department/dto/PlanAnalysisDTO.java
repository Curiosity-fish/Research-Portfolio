package com.zjnu.academic.department.dto;

import java.util.List;

public record PlanAnalysisDTO(
        String grade,
        List<String> dimensions,
        List<Integer> standard,
        List<Integer> actual,
        List<String> tips
) {
}
