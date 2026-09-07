package com.zjnu.academic.development.dto;

import java.util.List;

public record SchoolItemDTO(
        String id,
        String type,
        String name,
        String nameZh,
        String country,
        String countryEmoji,
        String location,
        String rank,
        String tier,
        List<String> programs,
        String admissionGpa,
        List<String> examRequirements,
        String languageRequirement,
        List<String> highlights,
        int matchScore,
        List<String> matchReasons,
        String officialUrl,
        String deadline
) {
}
