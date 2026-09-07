package com.zjnu.academic.student.service;

import com.zjnu.academic.alert.service.AlertService;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.health.entity.HealthReport;
import com.zjnu.academic.health.mapper.HealthReportMapper;
import com.zjnu.academic.student.dto.HealthReportDTO;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.ProfileScoreMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.user.mapper.UserMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class StudentServiceHealthDetailTest {

    @Mock
    private StudentMapper studentMapper;

    @Mock
    private GpaHistoryMapper gpaHistoryMapper;

    @Mock
    private ProfileScoreMapper profileScoreMapper;

    @Mock
    private HealthReportMapper healthReportMapper;

    @Mock
    private UserMapper userMapper;

    @Mock
    private AlertService alertService;

    @Mock
    private GradeMapper gradeMapper;

    @Mock
    private CourseClassMapper courseClassMapper;

    @Mock
    private CourseMapper courseMapper;

    private StudentService studentService;

    @BeforeEach
    void setUp() {
        studentService = new StudentService(
                studentMapper, gpaHistoryMapper, profileScoreMapper,
                healthReportMapper, userMapper, alertService,
                gradeMapper, courseClassMapper, courseMapper);
    }

    @Test
    void ownerCanReadHealthReportDetail() {
        Student student = new Student();
        student.setId(10L);
        student.setUserId(5L);
        student.setStudentId("2022001");
        HealthReport report = new HealthReport();
        report.setId(3L);
        report.setStudentId(10L);
        report.setTerm("2024-2025-1");
        report.setTotalScore(85);

        when(studentMapper.selectOne(any())).thenReturn(student);
        when(healthReportMapper.selectById(3L)).thenReturn(report);

        HealthReportDTO dto = studentService.healthReportDetail(5L, 3L);

        assertEquals("2022001", dto.studentId());
        assertEquals(85, dto.totalScore());
    }

    @Test
    void foreignReportIsRejected() {
        Student student = new Student();
        student.setId(10L);
        student.setUserId(5L);
        student.setStudentId("2022001");
        HealthReport report = new HealthReport();
        report.setId(3L);
        report.setStudentId(99L);

        when(studentMapper.selectOne(any())).thenReturn(student);
        when(healthReportMapper.selectById(3L)).thenReturn(report);

        assertThrows(BusinessException.class, () -> studentService.healthReportDetail(5L, 3L));
    }
}
