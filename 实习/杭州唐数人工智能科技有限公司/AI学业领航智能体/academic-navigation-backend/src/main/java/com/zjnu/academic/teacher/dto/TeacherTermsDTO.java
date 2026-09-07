package com.zjnu.academic.teacher.dto;

import java.util.List;

public record TeacherTermsDTO(
        String className,
        String latestTerm,
        List<String> terms
) {
}
