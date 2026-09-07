package com.zjnu.academic.course.dto;

import com.zjnu.academic.teacher.dto.FailStudentDTO;

import java.util.List;

public record CourseClassStatDTO(
        String className,
        double avg,
        double pass,
        double high,
        double low,
        int count,
        int risk,
        List<FailStudentDTO> failStudents,
        List<TermFailStudentDTO> termFailStudents
) {
}
