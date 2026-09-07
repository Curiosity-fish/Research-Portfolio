package com.zjnu.academic.dean.service;

import com.zjnu.academic.alert.entity.Alert;
import com.zjnu.academic.alert.mapper.AlertMapper;
import com.zjnu.academic.course.entity.Grade;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.dean.dto.DeanDashboardDTO;
import com.zjnu.academic.department.entity.Department;
import com.zjnu.academic.department.mapper.DepartmentMapper;
import com.zjnu.academic.major.entity.Major;
import com.zjnu.academic.major.mapper.MajorMapper;
import com.zjnu.academic.student.entity.GpaHistory;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.user.mapper.UserMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class DeanServiceTest {

    @Mock private MajorMapper majorMapper;
    @Mock private StudentMapper studentMapper;
    @Mock private UserMapper userMapper;
    @Mock private GpaHistoryMapper gpaHistoryMapper;
    @Mock private GradeMapper gradeMapper;
    @Mock private AlertMapper alertMapper;
    @Mock private DepartmentMapper departmentMapper;

    private DeanService service;

    @BeforeEach
    void setUp() {
        service = new DeanService(majorMapper, studentMapper, userMapper,
                gpaHistoryMapper, gradeMapper, alertMapper, departmentMapper);
    }

    @Test
    void availableTermsComeFromCollegeStudentHistory() {
        stubScope();
        when(gpaHistoryMapper.selectList(any())).thenReturn(List.of(
                history(10L, "2023-2024-2", "3.20"),
                history(11L, "2024-2025-1", "2.80")));
        when(gradeMapper.selectList(any())).thenReturn(List.of());

        assertEquals(List.of("2024-2025-1", "2023-2024-2"), service.availableTerms(99L));
    }

    @Test
    void dashboardUsesSelectedTermAcrossCollegeAggregates() {
        stubScope();
        when(gpaHistoryMapper.selectList(any())).thenReturn(List.of(
                history(10L, "2023-2024-2", "3.20"),
                history(11L, "2023-2024-2", "2.80")));

        Grade passed = new Grade();
        passed.setStudentId(10L);
        passed.setTerm("2023-2024-2");
        passed.setStatus("passed");
        when(gradeMapper.selectList(any())).thenReturn(List.of(passed));

        Alert resolved = new Alert();
        resolved.setId(1L);
        resolved.setStudentId(10L);
        resolved.setLevel("red");
        resolved.setStatus("resolved");
        resolved.setTriggerDate(LocalDate.of(2024, 7, 10));
        when(alertMapper.selectList(any())).thenReturn(List.of(resolved));

        DeanDashboardDTO dto = service.dashboard(99L, "2023-2024-2");

        assertEquals(new BigDecimal("3.00"), dto.avgGpa());
        assertEquals(new BigDecimal("50.0"), dto.alertRate());
        assertEquals(new BigDecimal("3.00"), dto.departmentRanking().get(0).avgGpa());
        assertEquals(1, dto.alertHeatmap().get(0).red());
    }

    private void stubScope() {
        Department department = new Department();
        department.setId(1L);
        department.setDeanId(99L);
        when(departmentMapper.selectOne(any())).thenReturn(department);

        Major major = new Major();
        major.setId(2L);
        major.setDepartmentId(1L);
        major.setName("应用统计");
        when(majorMapper.selectList(any())).thenReturn(List.of(major));

        Student first = new Student();
        first.setId(10L);
        first.setUserId(20L);
        first.setMajorId(2L);
        Student second = new Student();
        second.setId(11L);
        second.setUserId(21L);
        second.setMajorId(2L);
        when(studentMapper.selectList(any())).thenReturn(List.of(first, second));
    }

    private GpaHistory history(Long studentId, String term, String gpa) {
        GpaHistory history = new GpaHistory();
        history.setStudentId(studentId);
        history.setTerm(term);
        history.setGpa(new BigDecimal(gpa));
        return history;
    }
}
