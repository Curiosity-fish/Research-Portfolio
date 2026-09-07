package com.zjnu.academic.alert.dto;

import java.util.List;

public record AlertItemDTO(
        String id,
        String studentId,
        String studentName,
        String type,
        String level,
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
