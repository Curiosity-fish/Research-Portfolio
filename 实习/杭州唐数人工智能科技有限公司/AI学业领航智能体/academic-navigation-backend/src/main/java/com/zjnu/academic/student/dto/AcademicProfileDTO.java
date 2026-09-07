package com.zjnu.academic.student.dto;

import java.util.List;

public record AcademicProfileDTO(
        List<ProfileDimensionDTO> dimensions,
        List<GpaTrendItemDTO> gpaTrend,
        List<RankTrendItemDTO> rankTrend
) {
}
