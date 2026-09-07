package com.zjnu.academic.department.dto;

import com.zjnu.academic.student.dto.GpaTrendItemDTO;

import java.math.BigDecimal;
import java.util.List;

public record DepartmentOverviewDTO(
        int totalStudents,
        BigDecimal avgGpa,
        BigDecimal alertRate,
        BigDecimal passRate,
        List<GradeGpaTrendItemDTO> gradeGpaTrend,
        AlertDistributionDTO alertDistribution,
        List<FocusStudentDTO> focusStudents
) {
    public record GradeGpaTrendItemDTO(String grade, List<GpaTrendItemDTO> trend) {
    }

    public record AlertDistributionDTO(int yellow, int orange, int red) {
    }

    public record FocusStudentDTO(
            String id,
            String name,
            String studentId,
            String grade,
            String major,
            BigDecimal gpa,
            String alertLevel,
            String className
    ) {
    }
}
