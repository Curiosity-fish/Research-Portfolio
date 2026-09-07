package com.zjnu.academic.course.dto;

import java.util.List;

public record CourseTeacherCourseDTO(
        String id,
        String name,
        String code,
        String term,
        int students,
        double avgScore,
        double passRate,
        int highRiskCount,
        List<Integer> scoreDistribution,
        List<CourseClassStatDTO> classes
) {
}
