package com.zjnu.academic.student.dto;

import java.math.BigDecimal;

public record StudentDTO(
        String id,
        String name,
        String studentId,
        String college,
        String major,
        String className,
        String grade,
        BigDecimal gpa,
        Integer rank,
        Integer totalStudents,
        String alertLevel,
        String avatar
) {
}
