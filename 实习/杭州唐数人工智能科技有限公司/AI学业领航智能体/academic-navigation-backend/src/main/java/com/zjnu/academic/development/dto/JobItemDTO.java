package com.zjnu.academic.development.dto;

import java.util.List;

public record JobItemDTO(
        String id,
        String platform,
        String platformLabel,
        String companyName,
        String companySize,
        String companyStage,
        String jobTitle,
        String salaryRange,
        String city,
        String district,
        String education,
        String experience,
        List<String> tags,
        List<String> highlights,
        int matchScore,
        List<String> matchReasons,
        String publishDate,
        String sourceUrl,
        String category
) {
}
