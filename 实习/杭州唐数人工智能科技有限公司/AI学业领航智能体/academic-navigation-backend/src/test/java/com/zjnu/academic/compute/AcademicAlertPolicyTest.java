package com.zjnu.academic.compute;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;

class AcademicAlertPolicyTest {

    @Test
    void classifiesNormalStudent() {
        assertEquals("none", assess(0, 0));
    }

    @Test
    void classifiesOneOrTwoAbsencesWithoutFailureAsYellow() {
        assertEquals("yellow", assess(0, 1));
        assertEquals("yellow", assess(0, 2));
    }

    @Test
    void classifiesOneFailureOrThreeAbsencesAsOrange() {
        assertEquals("orange", assess(1, 0));
        assertEquals("orange", assess(1, 2));
        assertEquals("orange", assess(0, 3));
    }

    @Test
    void classifiesTwoFailuresOrFourAbsencesAsRed() {
        assertEquals("red", assess(2, 0));
        assertEquals("red", assess(2, 1));
        assertEquals("red", assess(0, 4));
    }

    private String assess(int failedCourses, int absences) {
        return AcademicAlertPolicy.assess(failedCourses, absences).level();
    }
}
