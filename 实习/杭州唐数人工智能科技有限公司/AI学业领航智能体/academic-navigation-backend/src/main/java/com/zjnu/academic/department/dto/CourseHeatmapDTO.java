package com.zjnu.academic.department.dto;

import java.util.List;

public record CourseHeatmapDTO(
        List<String> courses,
        List<String> grades,
        List<int[]> heatData
) {
}
