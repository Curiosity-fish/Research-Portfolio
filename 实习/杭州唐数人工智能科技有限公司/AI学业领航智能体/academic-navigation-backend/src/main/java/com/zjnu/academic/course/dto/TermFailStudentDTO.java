package com.zjnu.academic.course.dto;

import java.math.BigDecimal;
import java.util.List;

public record TermFailStudentDTO(
        String name,
        String studentId,
        List<String> failedCourses,
        BigDecimal lowestScore
) {
}
