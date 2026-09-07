package com.zjnu.academic.alert.service;

import com.zjnu.academic.alert.entity.Alert;
import org.junit.jupiter.api.Test;

import java.time.LocalDate;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

class AlertStudentAggregatorTest {

    @Test
    void termAggregationKeepsResolvedHistoricalAlert() {
        Alert resolved = alert(1L, 10L, "red", "resolved", LocalDate.of(2024, 7, 10));

        assertEquals(List.of(resolved), AlertStudentAggregator.perStudent(List.of(resolved)));
        assertTrue(AlertStudentAggregator.currentPerStudent(List.of(resolved)).isEmpty());
    }

    @Test
    void termAggregationKeepsMostSevereAlertPerStudent() {
        Alert yellow = alert(1L, 10L, "yellow", "resolved", LocalDate.of(2024, 7, 10));
        Alert orange = alert(2L, 10L, "orange", "resolved", LocalDate.of(2024, 7, 9));

        assertEquals(List.of(orange), AlertStudentAggregator.perStudent(List.of(yellow, orange)));
    }

    private Alert alert(Long id, Long studentId, String level, String status, LocalDate date) {
        Alert alert = new Alert();
        alert.setId(id);
        alert.setStudentId(studentId);
        alert.setLevel(level);
        alert.setStatus(status);
        alert.setTriggerDate(date);
        return alert;
    }
}
