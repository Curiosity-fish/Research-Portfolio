package com.zjnu.academic.alert.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.alert.dto.AlertStatsDTO;
import com.zjnu.academic.alert.dto.InterventionRequest;
import com.zjnu.academic.alert.entity.Alert;
import com.zjnu.academic.alert.entity.InterventionRecord;
import com.zjnu.academic.alert.mapper.AlertMapper;
import com.zjnu.academic.alert.mapper.InterventionRecordMapper;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.common.term.AcademicTerm;
import com.zjnu.academic.course.entity.CourseClass;
import com.zjnu.academic.course.entity.Grade;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.teacher.entity.Teacher;
import com.zjnu.academic.teacher.mapper.TeacherMapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;

@Service
public class AlertService {

    private static final DateTimeFormatter DATETIME = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm");

    private final AlertMapper alertMapper;
    private final StudentMapper studentMapper;
    private final TeacherMapper teacherMapper;
    private final CourseClassMapper courseClassMapper;
    private final GradeMapper gradeMapper;
    private final InterventionRecordMapper interventionRecordMapper;
    private final UserMapper userMapper;

    @Autowired
    public AlertService(AlertMapper alertMapper,
                        StudentMapper studentMapper,
                        TeacherMapper teacherMapper,
                        CourseClassMapper courseClassMapper,
                        GradeMapper gradeMapper,
                        InterventionRecordMapper interventionRecordMapper,
                        UserMapper userMapper) {
        this.alertMapper = alertMapper;
        this.studentMapper = studentMapper;
        this.teacherMapper = teacherMapper;
        this.courseClassMapper = courseClassMapper;
        this.gradeMapper = gradeMapper;
        this.interventionRecordMapper = interventionRecordMapper;
        this.userMapper = userMapper;
    }

    public PageResult<AlertItemDTO> list(Long userId, String role, String level, String status,
                                         String keyword, int page, int size) {
        return list(userId, role, null, level, status, keyword, page, size);
    }

    public PageResult<AlertItemDTO> list(Long userId, String role, String term,
                                         String level, String status, String keyword,
                                         int page, int size) {
        List<Long> scope = resolveScope(userId, role);
        return listForStudents(scope, term, level, status, keyword, page, size);
    }

    public PageResult<AlertItemDTO> listForStudents(List<Long> scope, String term,
                                                    String level, String status, String keyword,
                                                    int page, int size) {
        List<Alert> alerts = loadAlerts(scope, term, level, status);
        Map<Long, Student> students = loadStudents(alerts);
        Map<Long, User> users = loadUsers(students);

        List<AlertItemDTO> dtos = alerts.stream()
                .filter(alert -> scope == null || scope.contains(alert.getStudentId()))
                .map(alert -> toDto(alert, students, users))
                .filter(dto -> matchesKeyword(dto, keyword))
                .sorted(Comparator.comparing(AlertItemDTO::date).reversed())
                .toList();

        int from = Math.min((page - 1) * size, dtos.size());
        int to = Math.min(from + size, dtos.size());
        return PageResult.of(dtos.subList(from, to), dtos.size(), page, size);
    }

    public AlertItemDTO detail(Long userId, String role, Long id) {
        Alert alert = requireAlert(id);
        ensureAccess(alert, resolveScope(userId, role));
        Map<Long, Student> students = loadStudents(List.of(alert));
        Map<Long, User> users = loadUsers(students);
        return toDto(alert, students, users);
    }

    public int unreadCount(Long userId, String role) {
        return loadAlerts(resolveScope(userId, role), null, null, "pending").size();
    }

    public void confirm(Long userId, Long id) {
        Alert alert = requireAlert(id);
        Student student = studentMapper.selectOne(
                new LambdaQueryWrapper<Student>().eq(Student::getUserId, userId));
        if (student == null || !student.getId().equals(alert.getStudentId())) {
            throw new BusinessException(403, "无权限操作该预警");
        }
        alert.setStudentConfirmedAt(LocalDateTime.now());
        if ("pending".equals(alert.getStatus())) {
            alert.setStatus("processing");
        }
        alertMapper.updateById(alert);
    }

    public void intervention(Long userId, String role, Long id, InterventionRequest request) {
        if (!Set.of("teacher", "course_teacher", "department", "dean").contains(role)) {
            throw new BusinessException(403, "无权限提交干预");
        }
        Alert alert = requireAlert(id);
        ensureAccess(alert, resolveScope(userId, role));

        Teacher teacher = teacherMapper.selectOne(
                new LambdaQueryWrapper<Teacher>().eq(Teacher::getUserId, userId));
        InterventionRecord record = new InterventionRecord();
        record.setAlertId(alert.getId());
        record.setTeacherId(teacher == null ? null : teacher.getId());
        record.setInterventionDate(request.interventionDate());
        record.setMethods(request.methods());
        record.setContent(request.content());
        record.setStudentResponse(request.studentResponse());
        record.setFollowUpPlan(request.followUpPlan());
        interventionRecordMapper.insert(record);

        alert.setStatus("processing");
        alert.setHandlerId(userId);
        alertMapper.updateById(alert);
    }

    public AlertStatsDTO stats(Long userId, String role, String term) {
        List<Alert> alerts = loadAlerts(resolveScope(userId, role), term, null, null);
        int yellow = 0;
        int orange = 0;
        int red = 0;
        for (Alert alert : alerts) {
            switch (alert.getLevel()) {
                case "yellow" -> yellow++;
                case "orange" -> orange++;
                case "red" -> red++;
                default -> {
                }
            }
        }
        return new AlertStatsDTO(alerts.size(), yellow, orange, red);
    }

    private List<Long> resolveScope(Long userId, String role) {
        return switch (role) {
            case "student" -> {
                Student student = studentMapper.selectOne(
                        new LambdaQueryWrapper<Student>().eq(Student::getUserId, userId));
                yield student == null ? List.of() : List.of(student.getId());
            }
            case "teacher" -> {
                Teacher teacher = teacherMapper.selectOne(new LambdaQueryWrapper<Teacher>()
                        .eq(Teacher::getUserId, userId)
                        .eq(Teacher::getIsClassAdvisor, 1));
                if (teacher == null || teacher.getClassId() == null) {
                    yield List.of();
                }
                List<Student> students = studentMapper.selectList(
                        new LambdaQueryWrapper<Student>().eq(Student::getClassId, teacher.getClassId()));
                yield students.stream().map(Student::getId).toList();
            }
            case "course_teacher" -> {
                Teacher teacher = teacherMapper.selectOne(
                        new LambdaQueryWrapper<Teacher>().eq(Teacher::getUserId, userId));
                if (teacher == null) {
                    yield List.of();
                }
                List<CourseClass> courseClasses = courseClassMapper.selectList(
                        new LambdaQueryWrapper<CourseClass>().eq(CourseClass::getTeacherId, teacher.getId()));
                if (courseClasses.isEmpty()) {
                    yield List.of();
                }
                List<Long> courseClassIds = courseClasses.stream().map(CourseClass::getId).toList();
                List<Grade> grades = gradeMapper.selectList(
                        new LambdaQueryWrapper<Grade>().in(Grade::getCourseClassId, courseClassIds));
                yield grades.stream().map(Grade::getStudentId).distinct().toList();
            }
            default -> null;
        };
    }

    private List<Alert> loadAlerts(List<Long> scope, String term, String level, String status) {
        if (scope != null && scope.isEmpty()) {
            return List.of();
        }
        LambdaQueryWrapper<Alert> wrapper = new LambdaQueryWrapper<>();
        if (scope != null && !scope.isEmpty()) {
            wrapper.in(Alert::getStudentId, scope);
        }
        if (level != null && !level.isBlank()) {
            wrapper.eq(Alert::getLevel, level);
        }
        if (status != null && !status.isBlank()) {
            wrapper.eq(Alert::getStatus, status);
        }
        AcademicTerm.range(term).ifPresent(range ->
                wrapper.between(Alert::getTriggerDate, range.start(), range.end()));
        wrapper.orderByDesc(Alert::getTriggerDate);
        return alertMapper.selectList(wrapper);
    }

    private Map<Long, Student> loadStudents(List<Alert> alerts) {
        List<Long> ids = alerts.stream().map(Alert::getStudentId).distinct().toList();
        if (ids.isEmpty()) {
            return Map.of();
        }
        return studentMapper.selectBatchIds(ids).stream()
                .collect(Collectors.toMap(Student::getId, student -> student));
    }

    private Map<Long, User> loadUsers(Map<Long, Student> students) {
        List<Long> ids = students.values().stream().map(Student::getUserId).distinct().toList();
        if (ids.isEmpty()) {
            return Map.of();
        }
        return userMapper.selectBatchIds(ids).stream()
                .collect(Collectors.toMap(User::getId, user -> user));
    }

    private AlertItemDTO toDto(Alert alert, Map<Long, Student> students, Map<Long, User> users) {
        Student student = students.get(alert.getStudentId());
        User user = student == null ? null : users.get(student.getUserId());
        return new AlertItemDTO(
                String.valueOf(alert.getId()),
                student == null ? "" : student.getStudentId(),
                user == null ? "" : user.getName(),
                alert.getType(),
                alert.getLevel(),
                alert.getTitle(),
                alert.getDescription(),
                alert.getCourse(),
                alert.getFailedCourses() == null ? List.of() : alert.getFailedCourses(),
                alert.getTriggerDate() == null ? "" : alert.getTriggerDate().toString(),
                alert.getStatus(),
                alert.getSuggestion(),
                alert.getTriggerEvent(),
                alert.getPushedAt() == null ? null : DATETIME.format(alert.getPushedAt())
        );
    }

    private boolean matchesKeyword(AlertItemDTO dto, String keyword) {
        if (keyword == null || keyword.isBlank()) {
            return true;
        }
        return dto.studentName().contains(keyword) || dto.studentId().contains(keyword);
    }

    private Alert requireAlert(Long id) {
        Alert alert = alertMapper.selectById(id);
        if (alert == null) {
            throw new BusinessException(404, "预警不存在");
        }
        return alert;
    }

    private void ensureAccess(Alert alert, List<Long> scope) {
        if (scope != null && !scope.contains(alert.getStudentId())) {
            throw new BusinessException(403, "无权限查看该预警");
        }
    }
}
