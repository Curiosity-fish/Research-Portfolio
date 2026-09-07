package com.zjnu.academic.alert.service;

import com.zjnu.academic.alert.entity.Alert;

import java.time.LocalDate;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;

/**
 * Aggregates warning records into the current warning student set used by
 * department and dean views.
 */
public final class AlertStudentAggregator {

    private AlertStudentAggregator() {
    }

    public static List<Alert> currentPerStudent(List<Alert> alerts) {
        return aggregate(alerts, true);
    }

    public static List<Alert> perStudent(List<Alert> alerts) {
        return aggregate(alerts, false);
    }

    private static List<Alert> aggregate(List<Alert> alerts, boolean activeOnly) {
        Map<Long, Alert> byStudent = new LinkedHashMap<>();
        for (Alert alert : alerts) {
            if (alert == null || alert.getStudentId() == null
                    || (activeOnly && "resolved".equals(alert.getStatus()))) {
                continue;
            }
            byStudent.merge(alert.getStudentId(), alert, AlertStudentAggregator::moreSevere);
        }
        return new ArrayList<>(byStudent.values());
    }

    private static Alert moreSevere(Alert first, Alert second) {
        int firstSeverity = severity(first.getLevel());
        int secondSeverity = severity(second.getLevel());
        if (secondSeverity != firstSeverity) {
            return secondSeverity > firstSeverity ? second : first;
        }
        LocalDate firstDate = dateOf(first);
        LocalDate secondDate = dateOf(second);
        if (!Objects.equals(firstDate, secondDate)) {
            return secondDate.isAfter(firstDate) ? second : first;
        }
        if (first.getId() == null) {
            return second;
        }
        if (second.getId() == null) {
            return first;
        }
        return second.getId() > first.getId() ? second : first;
    }

    private static int severity(String level) {
        return switch (level == null ? "" : level) {
            case "red" -> 3;
            case "orange" -> 2;
            case "yellow" -> 1;
            default -> 0;
        };
    }

    private static LocalDate dateOf(Alert alert) {
        if (alert.getTriggerDate() != null) {
            return alert.getTriggerDate();
        }
        return alert.getCreatedAt() == null ? LocalDate.MIN : alert.getCreatedAt().toLocalDate();
    }
}
