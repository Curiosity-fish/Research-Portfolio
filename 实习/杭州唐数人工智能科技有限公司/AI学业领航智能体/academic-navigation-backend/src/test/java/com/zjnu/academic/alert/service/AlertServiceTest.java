package com.zjnu.academic.alert.service;

import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.alert.dto.InterventionRequest;
import com.zjnu.academic.alert.entity.Alert;
import com.zjnu.academic.alert.mapper.AlertMapper;
import com.zjnu.academic.alert.mapper.InterventionRecordMapper;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.teacher.entity.Teacher;
import com.zjnu.academic.teacher.mapper.TeacherMapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.LocalDate;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class AlertServiceTest {

    @Mock
    private AlertMapper alertMapper;

    @Mock
    private StudentMapper studentMapper;

    @Mock
    private TeacherMapper teacherMapper;

    @Mock
    private CourseClassMapper courseClassMapper;

    @Mock
    private GradeMapper gradeMapper;

    @Mock
    private InterventionRecordMapper interventionRecordMapper;

    @Mock
    private UserMapper userMapper;

    private AlertService alertService;

    @BeforeEach
    void setUp() {
        alertService = new AlertService(
                alertMapper, studentMapper, teacherMapper, courseClassMapper,
                gradeMapper, interventionRecordMapper, userMapper);
    }

    @Test
    void studentOnlySeesOwnAlerts() {
        Student current = new Student();
        current.setId(10L);
        current.setUserId(5L);
        current.setStudentId("2022001");

        Student other = new Student();
        other.setId(11L);
        other.setUserId(6L);
        other.setStudentId("2022002");

        User currentUser = new User();
        currentUser.setId(5L);
        currentUser.setName("李明");
        User otherUser = new User();
        otherUser.setId(6L);
        otherUser.setName("王芳");

        when(studentMapper.selectOne(any())).thenReturn(current);
        when(alertMapper.selectList(any())).thenReturn(List.of(
                alert(1L, 10L, "yellow", "成绩预警"),
                alert(2L, 11L, "red", "学业危机")));
        when(studentMapper.selectBatchIds(any())).thenReturn(List.of(current, other));
        when(userMapper.selectBatchIds(any())).thenReturn(List.of(currentUser, otherUser));

        PageResult<AlertItemDTO> page = alertService.list(5L, "student", null, null, null, 1, 10);

        assertEquals(1, page.total());
        assertEquals("2022001", page.list().get(0).studentId());
        assertEquals("yellow", page.list().get(0).level());
    }

    @Test
    void emptyScopeReturnsEmptyPage() {
        when(studentMapper.selectOne(any())).thenReturn(null);

        PageResult<AlertItemDTO> page = alertService.list(5L, "student", null, null, null, 1, 10);

        assertEquals(0, page.total());
        assertEquals(0, page.list().size());
        verify(alertMapper, never()).selectList(any());
    }

    @Test
    void teacherCannotInterveneOutsideScope() {
        Teacher teacher = new Teacher();
        teacher.setId(1L);
        teacher.setUserId(5L);
        teacher.setIsClassAdvisor(1);
        teacher.setClassId(1L);
        Student inClass = new Student();
        inClass.setId(10L);
        when(teacherMapper.selectOne(any())).thenReturn(teacher);
        when(studentMapper.selectList(any())).thenReturn(List.of(inClass));
        when(alertMapper.selectById(1L)).thenReturn(alert(1L, 11L, "red", "academic"));

        assertThrows(BusinessException.class,
                () -> alertService.intervention(5L, "teacher", 1L,
                        new InterventionRequest(LocalDate.now(), List.of("talk"), "c", "ok", "f")));
    }

    private Alert alert(long id, long studentId, String level, String type) {
        Alert alert = new Alert();
        alert.setId(id);
        alert.setStudentId(studentId);
        alert.setLevel(level);
        alert.setType(type);
        alert.setTitle("预警");
        alert.setStatus("pending");
        alert.setTriggerDate(LocalDate.of(2025, 4, 1));
        return alert;
    }
}
