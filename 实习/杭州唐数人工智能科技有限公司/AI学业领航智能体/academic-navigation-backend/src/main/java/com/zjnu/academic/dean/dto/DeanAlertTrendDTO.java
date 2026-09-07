package com.zjnu.academic.dean.dto;

import java.util.List;

public record DeanAlertTrendDTO(
        List<String> terms,
        List<Integer> red,
        List<Integer> orange,
        List<Integer> yellow
) {
}
