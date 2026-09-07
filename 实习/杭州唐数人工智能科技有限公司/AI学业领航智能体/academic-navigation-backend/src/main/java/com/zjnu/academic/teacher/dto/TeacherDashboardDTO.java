package com.zjnu.academic.teacher.dto;

import com.zjnu.academic.alert.dto.AlertItemDTO;

import java.math.BigDecimal;
import java.util.List;

public record TeacherDashboardDTO(
        int totalStudents,
        BigDecimal avgGpa,
        int alertCount,
        BigDecimal passRate,
        AlertDistribution alertDistribution,
        List<DimensionAvgItem> dimensionAvg,
        List<AlertItemDTO> recentAlerts,
        List<GradeDistributionItem> gradeDistribution
) {
    public record AlertDistribution(int yellow, int orange, int red) {
    }

    public record DimensionAvgItem(String label, int value) {
    }

    public record GradeDistributionItem(String name, int value) {
    }
}
