package com.zjnu.academic.department.dto;

import java.math.BigDecimal;
import java.util.List;

public record DepartmentAlertDTO(
        String id,
        String studentId,
        String studentName,
        String grade,
        String className,
        String major,
        BigDecimal gpa,
        String trend,
        String level,
        int absences,
        String reason,
        String type,
        String title,
        String description,
        String course,
        List<String> failedCourses,
        String date,
        String status,
        String suggestion,
        String triggerEvent,
        String pushedAt
) {
}
