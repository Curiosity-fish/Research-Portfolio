package com.zjnu.academic.dean.dto;

import java.math.BigDecimal;
import java.util.List;

public record DeanAlertStatsDTO(
        String term,
        TotalDTO total,
        List<DeptRowDTO> byDept,
        List<GradeRowDTO> byGrade
) {
    public record TotalDTO(int red, int orange, int yellow, int total) {
    }

    public record DeptRowDTO(String dept, int totalStudents, int red, int orange, int yellow, BigDecimal rate) {
    }

    public record GradeRowDTO(String grade, int red, int orange, int yellow, int total, String note) {
    }
}
