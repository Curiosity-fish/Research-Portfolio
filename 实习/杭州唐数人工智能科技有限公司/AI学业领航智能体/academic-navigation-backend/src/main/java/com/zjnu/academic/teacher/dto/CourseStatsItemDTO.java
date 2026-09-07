package com.zjnu.academic.teacher.dto;

import java.math.BigDecimal;
import java.util.List;

public record CourseStatsItemDTO(
        String id,
        String name,
        BigDecimal credits,
        BigDecimal avgScore,
        BigDecimal passRate,
        BigDecimal failRate,
        List<FailStudentDTO> failStudents
) {
}
