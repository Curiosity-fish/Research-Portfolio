package com.zjnu.academic.department.service;

import com.zjnu.academic.alert.entity.Alert;
import com.zjnu.academic.alert.mapper.AlertMapper;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.department.dto.DepartmentOverviewDTO;
import com.zjnu.academic.department.entity.Department;
import com.zjnu.academic.department.mapper.DepartmentMapper;
import com.zjnu.academic.department.mapper.MajorPlanMapper;
import com.zjnu.academic.major.entity.Major;
import com.zjnu.academic.major.mapper.MajorMapper;
import com.zjnu.academic.student.entity.GpaHistory;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.user.entity.User;
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
class DepartmentServiceTest {

    @Mock private UserMapper userMapper;
    @Mock private MajorMapper majorMapper;
    @Mock private StudentMapper studentMapper;
    @Mock private GpaHistoryMapper gpaHistoryMapper;
    @Mock private GradeMapper gradeMapper;
    @Mock private CourseClassMapper courseClassMapper;
    @Mock private CourseMapper courseMapper;
    @Mock private AlertMapper alertMapper;
    @Mock private MajorPlanMapper majorPlanMapper;
    @Mock private DepartmentMapper departmentMapper;

    private DepartmentService service;

    @BeforeEach
    void setUp() {
        service = new DepartmentService(userMapper, majorMapper, studentMapper, gpaHistoryMapper,
                gradeMapper, courseClassMapper, courseMapper, alertMapper,
                majorPlanMapper, departmentMapper);
    }

    @Test
    void availableTermsComeFromScopedStudentHistory() {
        stubScope();
        when(gpaHistoryMapper.selectList(any())).thenReturn(List.of(
                history(10L, "2023-2024-2", "3.20"),
                history(10L, "2024-2025-1", "3.30"),
                history(11L, "2023-2024-2", "2.80")));
        when(gradeMapper.selectList(any())).thenReturn(List.of());

        assertEquals(List.of("2024-2025-1", "2023-2024-2"), service.availableTerms(7L));
    }

    @Test
    void overviewUsesSelectedTermGpaAndHistoricalAlerts() {
        List<Student> students = stubScope();
        List<GpaHistory> histories = List.of(
                history(10L, "2023-2024-2", "3.20"),
                history(11L, "2023-2024-2", "2.80"));
        when(gpaHistoryMapper.selectList(any())).thenReturn(histories);
        when(gradeMapper.selectList(any())).thenReturn(List.of());
        when(userMapper.selectBatchIds(any())).thenReturn(List.of(user(20L, "学生甲"), user(21L, "学生乙")));

        Alert alert = new Alert();
        alert.setId(1L);
        alert.setStudentId(students.get(0).getId());
        alert.setLevel("red");
        alert.setStatus("resolved");
        alert.setTriggerDate(LocalDate.of(2024, 7, 10));
        when(alertMapper.selectList(any())).thenReturn(List.of(alert));

        DepartmentOverviewDTO dto = service.overview(7L, "2023-2024-2");

        assertEquals(new BigDecimal("3.00"), dto.avgGpa());
        assertEquals(new BigDecimal("50.0"), dto.alertRate());
        assertEquals(1, dto.alertDistribution().red());
        assertEquals(new BigDecimal("3.20"), dto.focusStudents().get(0).gpa());
        assertEquals("red", dto.focusStudents().get(0).alertLevel());
    }

    private List<Student> stubScope() {
        User owner = new User();
        owner.setId(7L);
        owner.setCollege("数学与计算机科学学院");
        owner.setMajor("应用统计");
        when(userMapper.selectById(7L)).thenReturn(owner);

        Department department = new Department();
        department.setId(1L);
        when(departmentMapper.selectOne(any())).thenReturn(department);

        Major major = new Major();
        major.setId(2L);
        major.setName("应用统计");
        when(majorMapper.selectOne(any())).thenReturn(major);

        Student first = student(10L, 20L, "20231001");
        Student second = student(11L, 21L, "20231002");
        List<Student> students = List.of(first, second);
        when(studentMapper.selectList(any())).thenReturn(students);
        return students;
    }

    private Student student(Long id, Long userId, String studentId) {
        Student student = new Student();
        student.setId(id);
        student.setUserId(userId);
        student.setStudentId(studentId);
        return student;
    }

    private User user(Long id, String name) {
        User user = new User();
        user.setId(id);
        user.setName(name);
        user.setGrade("2023");
        user.setClassName("应统2301");
        user.setMajor("应用统计");
        return user;
    }

    private GpaHistory history(Long studentId, String term, String gpa) {
        GpaHistory history = new GpaHistory();
        history.setStudentId(studentId);
        history.setTerm(term);
        history.setGpa(new BigDecimal(gpa));
        return history;
    }
}
