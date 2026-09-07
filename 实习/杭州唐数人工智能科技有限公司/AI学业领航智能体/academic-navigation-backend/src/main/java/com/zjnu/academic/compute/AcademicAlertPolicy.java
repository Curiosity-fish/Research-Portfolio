package com.zjnu.academic.compute;

/**
 * Academic early-warning policy shared by recomputation and tests.
 * Severity is based on failed courses and absences only. GPA and psychological
 * assessment data remain available for reference but are not grading signals.
 */
public final class AcademicAlertPolicy {

    private AcademicAlertPolicy() {
    }

    public static Assessment assess(int failedCourses, int absences) {
        String level;
        if (failedCourses >= 2 || absences >= 4) {
            level = "red";
        } else if (failedCourses >= 1 || absences == 3) {
            level = "orange";
        } else if (failedCourses == 0 && absences >= 1 && absences <= 2) {
            level = "yellow";
        } else {
            level = "none";
        }

        return new Assessment(level, primaryRisk(failedCourses));
    }

    private static String primaryRisk(int failedCourses) {
        return failedCourses > 0 ? "课程预警" : "出勤预警";
    }

    public record Assessment(String level, String primaryRisk) {
    }
}
