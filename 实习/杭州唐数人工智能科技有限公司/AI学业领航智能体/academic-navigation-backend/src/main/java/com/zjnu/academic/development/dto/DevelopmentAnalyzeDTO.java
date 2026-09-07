package com.zjnu.academic.development.dto;

import java.util.List;

public record DevelopmentAnalyzeDTO(
        String targetKey,
        String targetLabel,
        int matchScore,
        List<DevelopmentPathDTO.GapItemDTO> gaps,
        List<DevelopmentPathDTO.ActionItemDTO> actions,
        List<DevelopmentPathDTO.MilestoneItemDTO> milestones,
        List<String> suggestions
) {
}
