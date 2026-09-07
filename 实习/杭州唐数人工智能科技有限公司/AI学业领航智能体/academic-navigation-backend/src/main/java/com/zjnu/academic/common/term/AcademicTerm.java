package com.zjnu.academic.common.term;

import java.time.LocalDate;
import java.time.YearMonth;
import java.util.Optional;

public final class AcademicTerm {

    private AcademicTerm() {
    }

    public static Optional<DateRange> range(String term) {
        if (term == null || term.isBlank()) {
            return Optional.empty();
        }
        try {
            String[] parts = term.split("-");
            if (parts.length != 3 || (!"1".equals(parts[2]) && !"2".equals(parts[2]))) {
                return Optional.empty();
            }
            int startYear = Integer.parseInt(parts[0]);
            int endYear = Integer.parseInt(parts[1]);
            if (endYear != startYear + 1) {
                return Optional.empty();
            }
            if ("1".equals(parts[2])) {
                return Optional.of(new DateRange(
                        LocalDate.of(startYear, 9, 1),
                        YearMonth.of(endYear, 2).atEndOfMonth()));
            }
            return Optional.of(new DateRange(
                    LocalDate.of(endYear, 3, 1),
                    LocalDate.of(endYear, 8, 31)));
        } catch (RuntimeException ignored) {
            return Optional.empty();
        }
    }

    public static String fromDate(LocalDate date) {
        if (date == null) {
            return "";
        }
        int year = date.getYear();
        int month = date.getMonthValue();
        if (month >= 9) {
            return year + "-" + (year + 1) + "-1";
        }
        if (month <= 2) {
            return (year - 1) + "-" + year + "-1";
        }
        return (year - 1) + "-" + year + "-2";
    }

    public record DateRange(LocalDate start, LocalDate end) {
    }
}
