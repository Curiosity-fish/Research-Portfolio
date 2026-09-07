package com.zjnu.academic.teacher.dto;

import java.math.BigDecimal;
import java.util.Map;

public record GpaProgressStudentDTO(
        String id,
        String name,
        String studentId,
        Map<String, BigDecimal> gpaByTerm
) {
}
