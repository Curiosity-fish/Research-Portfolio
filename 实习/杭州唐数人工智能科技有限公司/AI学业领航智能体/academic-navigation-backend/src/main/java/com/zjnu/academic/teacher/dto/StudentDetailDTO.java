package com.zjnu.academic.teacher.dto;

import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.student.dto.AcademicProfileDTO;
import com.zjnu.academic.student.dto.GpaTrendItemDTO;
import com.zjnu.academic.student.dto.StudentDTO;

import java.util.List;

public record StudentDetailDTO(
        StudentDTO student,
        AcademicProfileDTO profile,
        List<AlertItemDTO> alerts,
        List<GpaTrendItemDTO> gpaTrend
) {
}
