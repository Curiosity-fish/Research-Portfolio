package com.zjnu.academic.student.dto;

import com.zjnu.academic.alert.dto.AlertItemDTO;

import java.math.BigDecimal;
import java.util.List;

public record StudentDashboardDTO(
        BigDecimal gpa,
        Integer rank,
        Integer totalStudents,
        Integer classRank,
        Integer classTotalStudents,
        BigDecimal credits,
        BigDecimal totalCredits,
        int courseCount,
        int alertCount,
        Integer healthScore,
        List<GpaTrendItemDTO> gpaTrend,
        List<AlertItemDTO> recentAlerts
) {
}
