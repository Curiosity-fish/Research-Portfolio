package com.zjnu.academic.alert.dto;

import java.time.LocalDate;
import java.util.List;

public record InterventionRequest(
        LocalDate interventionDate,
        List<String> methods,
        String content,
        String studentResponse,
        String followUpPlan
) {
}
