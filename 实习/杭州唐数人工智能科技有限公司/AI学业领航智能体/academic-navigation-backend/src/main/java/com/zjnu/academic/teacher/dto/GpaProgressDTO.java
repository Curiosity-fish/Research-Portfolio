package com.zjnu.academic.teacher.dto;

import java.util.List;

public record GpaProgressDTO(
        List<String> terms,
        List<GpaProgressStudentDTO> students
) {
}
