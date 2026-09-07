package com.zjnu.academic.development.dto;

import java.util.List;

public record DevelopmentPathDTO(
        String key,
        String label,
        int matchScore,
        List<GapItemDTO> gapItems,
        List<ActionItemDTO> actionItems,
        List<MilestoneItemDTO> milestones
) {
    public record GapItemDTO(String dimension, Object current, Object target, String urgent) {
    }

    public record ActionItemDTO(String title, String description, String urgency, String deadline) {
    }

    public record MilestoneItemDTO(String title, String date, String status, String description) {
    }
}
