package com.zjnu.academic.common.term;

import org.junit.jupiter.api.Test;

import java.time.LocalDate;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

class AcademicTermTest {

    @Test
    void firstSemesterIncludesLeapDay() {
        AcademicTerm.DateRange range = AcademicTerm.range("2023-2024-1").orElseThrow();

        assertEquals(LocalDate.of(2023, 9, 1), range.start());
        assertEquals(LocalDate.of(2024, 2, 29), range.end());
    }

    @Test
    void secondSemesterUsesMarchThroughAugust() {
        AcademicTerm.DateRange range = AcademicTerm.range("2024-2025-2").orElseThrow();

        assertEquals(LocalDate.of(2025, 3, 1), range.start());
        assertEquals(LocalDate.of(2025, 8, 31), range.end());
    }

    @Test
    void invalidTermHasNoRange() {
        assertTrue(AcademicTerm.range("2024-autumn").isEmpty());
        assertTrue(AcademicTerm.range(null).isEmpty());
    }

    @Test
    void dateMapsBackToCanonicalTerm() {
        assertEquals("2024-2025-1", AcademicTerm.fromDate(LocalDate.of(2025, 1, 15)));
        assertEquals("2024-2025-2", AcademicTerm.fromDate(LocalDate.of(2025, 7, 10)));
    }
}
