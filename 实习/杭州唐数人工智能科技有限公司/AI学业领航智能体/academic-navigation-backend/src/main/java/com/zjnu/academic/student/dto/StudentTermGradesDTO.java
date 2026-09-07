package com.zjnu.academic.student.dto;

import java.math.BigDecimal;
import java.util.List;

public record StudentTermGradesDTO(
        String term,
        BigDecimal gpa,
        BigDecimal avgGpa,
        Integer rank,
        Integer totalStudents,
        BigDecimal earnedCredits,
        int courseCount,
        int passedCount,
        int failedCount,
        List<StudentCourseGradeDTO> courses
) {
}
