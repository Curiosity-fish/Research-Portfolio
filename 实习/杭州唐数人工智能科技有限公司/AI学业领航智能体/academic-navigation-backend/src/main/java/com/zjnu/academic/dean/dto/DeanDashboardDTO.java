package com.zjnu.academic.dean.dto;

import java.math.BigDecimal;
import java.util.List;

public record DeanDashboardDTO(
        int totalStudents,
        BigDecimal avgGpa,
        BigDecimal alertRate,
        BigDecimal coursePassRate,
        BigDecimal interventionResponseRate,
        List<DepartmentRankingDTO> departmentRanking,
        List<AlertHeatmapDTO> alertHeatmap
) {
    public record DepartmentRankingDTO(String dept, int healthScore, BigDecimal passRate, BigDecimal avgGpa) {
    }

    public record AlertHeatmapDTO(String dept, int yellow, int orange, int red) {
    }
}
