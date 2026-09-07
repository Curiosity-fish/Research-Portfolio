package com.zjnu.academic.student.dto;

import java.math.BigDecimal;

public record StudentCourseGradeDTO(
        String courseCode,
        String courseName,
        BigDecimal credits,
        BigDecimal score,
        BigDecimal gradePoint,
        String status
) {
}
