package com.zjnu.academic.department.dto;

import java.math.BigDecimal;
import java.util.List;

public record DepartmentCourseDTO(
        String id,
        String name,
        String code,
        BigDecimal credits,
        BigDecimal avgScore,
        BigDecimal passRate,
        BigDecimal highRate,
        BigDecimal lowRate,
        List<GradeStatDTO> gradeStats
) {
    public record GradeStatDTO(
            String grade,
            BigDecimal avgScore,
            BigDecimal passRate,
            BigDecimal highRate,
            BigDecimal lowRate
    ) {
    }
}
