package com.zjnu.academic.student.service;

import com.zjnu.academic.alert.service.AlertService;
import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.health.entity.HealthReport;
import com.zjnu.academic.health.mapper.HealthReportMapper;
import com.zjnu.academic.student.dto.AcademicProfileDTO;
import com.zjnu.academic.student.dto.ProfileDimensionDTO;
import com.zjnu.academic.student.dto.StudentDashboardDTO;
import com.zjnu.academic.student.entity.GpaHistory;
import com.zjnu.academic.student.entity.ProfileScore;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.ProfileScoreMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class StudentServiceTest {

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
    void dashboardAggregatesStudentData() {
        Student student = new Student();
        student.setId(10L);
        student.setUserId(5L);
        student.setStudentId("2022001");
        student.setGpa(new BigDecimal("3.62"));
        student.setRank(15);
        student.setTotalStudents(120);
        student.setTotalCredits(new BigDecimal("89.0"));
        student.setRequiredCredits(new BigDecimal("176.0"));

        User user = new User();
        user.setId(5L);
        user.setName("李明");
        user.setGrade("2022");

        GpaHistory history = new GpaHistory();
        history.setTerm("2024-2025-1");
        history.setGpa(new BigDecimal("3.62"));
        history.setAvgGpa(new BigDecimal("3.20"));
        history.setRank(15);
        history.setTotalStudents(120);

        HealthReport health = new HealthReport();
        health.setTerm("2024-2025-1");
        health.setTotalScore(85);

        when(studentMapper.selectOne(any())).thenReturn(student);
        when(userMapper.selectById(5L)).thenReturn(user);
        when(gpaHistoryMapper.selectList(any())).thenReturn(List.of(history));
        when(gpaHistoryMapper.selectOne(any())).thenReturn(history);
        when(alertService.list(anyLong(), anyString(), any(), any(), any(), any(), anyInt(), anyInt()))
                .thenReturn(PageResult.of(List.of(), 0, 1, 1000));
        when(healthReportMapper.selectList(any())).thenReturn(List.of(health));

        StudentDashboardDTO dto = studentService.dashboard(5L, null);

        assertEquals(new BigDecimal("3.62"), dto.gpa());
        assertEquals(120, dto.totalStudents());
        assertEquals(85, dto.healthScore());
        assertEquals(1, dto.gpaTrend().size());
        assertEquals("大三上", dto.gpaTrend().get(0).semester());
    }

    @Test
    void dashboardUsesRequestedTermForAlerts() {
        Student student = new Student();
        student.setId(10L);
        student.setUserId(5L);
        student.setStudentId("2022001");
        student.setRequiredCredits(new BigDecimal("176.0"));

        User user = new User();
        user.setId(5L);
        user.setGrade("2022");

        GpaHistory history = new GpaHistory();
        history.setStudentId(10L);
        history.setTerm("2023-2024-2");
        history.setGpa(new BigDecimal("3.20"));

        AlertItemDTO historical = new AlertItemDTO(
                "9", "2022001", "李明", "课程预警", "orange", "橙色预警",
                "当学期挂科 1 门", null, List.of("数学分析"), "2024-07-10",
                "resolved", "已完成帮扶", "挂科=1", "2024-07-10 10:00");

        when(studentMapper.selectOne(any())).thenReturn(student);
        when(userMapper.selectById(5L)).thenReturn(user);
        when(gpaHistoryMapper.selectList(any())).thenReturn(List.of(history));
        when(gpaHistoryMapper.selectOne(any())).thenReturn(history);
        when(alertService.list(anyLong(), anyString(), any(), any(), any(), any(), anyInt(), anyInt()))
                .thenReturn(PageResult.of(List.of(historical), 1, 1, 1000));
        when(healthReportMapper.selectList(any())).thenReturn(List.of());

        StudentDashboardDTO dto = studentService.dashboard(5L, "2023-2024-2");

        verify(alertService).list(eq(5L), eq("student"), eq("2023-2024-2"),
                any(), any(), any(), eq(1), eq(1000));
        assertEquals(1, dto.alertCount());
        assertEquals(List.of(historical), dto.recentAlerts());
    }

    @Test
    void alertsUseRequestedTerm() {
        when(alertService.list(anyLong(), anyString(), any(), any(), any(), any(), anyInt(), anyInt()))
                .thenReturn(PageResult.of(List.of(), 0, 1, 20));

        studentService.alerts(5L, "2023-2024-2", 1, 20);

        verify(alertService).list(5L, "student", "2023-2024-2",
                null, null, null, 1, 20);
    }

    @Test
    void profileDimensionsFollowCanonicalOrder() {
        Student student = new Student();
        student.setId(10L);
        student.setUserId(5L);

        User user = new User();
        user.setId(5L);
        user.setGrade("2022");

        List<ProfileScore> scores = List.of(
                score("health", "身心健康", 88),
                score("quality", "综合素质", 68),
                score("academic", "学业成绩", 85),
                score("culture", "人文素养", 62),
                score("practice", "实践能力", 74));

        when(studentMapper.selectOne(any())).thenReturn(student);
        when(userMapper.selectById(5L)).thenReturn(user);
        when(profileScoreMapper.selectList(any())).thenReturn(scores);
        when(gpaHistoryMapper.selectList(any())).thenReturn(List.of());

        AcademicProfileDTO dto = studentService.profile(5L, "2024-2025-1");

        assertEquals(
                List.of("学业成绩", "实践能力", "综合素质", "人文素养", "身心健康"),
                dto.dimensions().stream().map(ProfileDimensionDTO::label).toList());
    }

    private static ProfileScore score(String key, String label, int score) {
        ProfileScore ps = new ProfileScore();
        ps.setDimensionKey(key);
        ps.setLabel(label);
        ps.setScore(score);
        return ps;
    }
}
