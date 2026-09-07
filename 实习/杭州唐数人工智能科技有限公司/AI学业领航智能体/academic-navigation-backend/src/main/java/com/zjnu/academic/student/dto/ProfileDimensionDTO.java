package com.zjnu.academic.student.dto;

import java.util.List;

public record ProfileDimensionDTO(
        String key,
        String label,
        Integer score,
        Integer avgScore,
        Integer maxScore,
        String description,
        List<String> details
) {
}
