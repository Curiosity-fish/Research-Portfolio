package com.zjnu.academic.teacher.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.alert.service.AlertService;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.compute.AcademicAlertPolicy;
import com.zjnu.academic.course.entity.Course;
import com.zjnu.academic.course.entity.CourseClass;
import com.zjnu.academic.course.entity.Grade;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.student.dto.AcademicProfileDTO;
import com.zjnu.academic.student.dto.GpaTrendItemDTO;
import com.zjnu.academic.student.dto.ProfileDimensionDTO;
import com.zjnu.academic.student.dto.RankTrendItemDTO;
import com.zjnu.academic.student.dto.StudentDTO;
import com.zjnu.academic.student.entity.GpaHistory;
import com.zjnu.academic.student.entity.ProfileScore;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.entity.StudentClass;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.ProfileScoreMapper;
import com.zjnu.academic.student.mapper.StudentClassMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.student.service.StudentService;
import com.zjnu.academic.rawdata.entity.Attendance;
import com.zjnu.academic.rawdata.mapper.AttendanceMapper;
import com.zjnu.academic.teacher.dto.CourseStatsItemDTO;
import com.zjnu.academic.teacher.dto.FailStudentDTO;
import com.zjnu.academic.teacher.dto.GpaProgressDTO;
import com.zjnu.academic.teacher.dto.GpaProgressStudentDTO;
import com.zjnu.academic.teacher.dto.StudentDetailDTO;
import com.zjnu.academic.teacher.dto.TeacherDashboardDTO;
import com.zjnu.academic.teacher.dto.TeacherTermsDTO;
import com.zjnu.academic.teacher.entity.Teacher;
import com.zjnu.academic.teacher.mapper.TeacherMapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.stream.Collectors;

@Service
public class TeacherService {

    private final StudentMapper studentMapper;
    private final StudentClassMapper studentClassMapper;
    private final TeacherMapper teacherMapper;
    private final GpaHistoryMapper gpaHistoryMapper;
    private final ProfileScoreMapper profileScoreMapper;
    private final GradeMapper gradeMapper;
    private final CourseClassMapper courseClassMapper;
    private final CourseMapper courseMapper;
    private final UserMapper userMapper;
    private final AttendanceMapper attendanceMapper;
    private final AlertService alertService;
    private final StudentService studentService;

    @Autowired
    public TeacherService(StudentMapper studentMapper,
                          StudentClassMapper studentClassMapper,
                          TeacherMapper teacherMapper,
                          GpaHistoryMapper gpaHistoryMapper,
                          ProfileScoreMapper profileScoreMapper,
                          GradeMapper gradeMapper,
                          CourseClassMapper courseClassMapper,
                          CourseMapper courseMapper,
                          UserMapper userMapper,
                          AttendanceMapper attendanceMapper,
                          AlertService alertService,
                          StudentService studentService) {
        this.studentMapper = studentMapper;
        this.studentClassMapper = studentClassMapper;
        this.teacherMapper = teacherMapper;
        this.gpaHistoryMapper = gpaHistoryMapper;
        this.profileScoreMapper = profileScoreMapper;
        this.gradeMapper = gradeMapper;
        this.courseClassMapper = courseClassMapper;
        this.courseMapper = courseMapper;
        this.userMapper = userMapper;
        this.attendanceMapper = attendanceMapper;
        this.alertService = alertService;
        this.studentService = studentService;
    }

    public TeacherTermsDTO terms(Long userId) {
        Teacher teacher = requireAdvisor(userId);
        StudentClass studentClass = teacher.getClassId() == null ? null
                : studentClassMapper.selectById(teacher.getClassId());
        List<Long> studentIds = classStudents(userId).stream().map(Student::getId).toList();
        List<String> terms = availableTerms(studentIds);
        return new TeacherTermsDTO(
                studentClass == null ? "" : studentClass.getName(),
                terms.isEmpty() ? "" : terms.get(0),
                terms);
    }

    public TeacherDashboardDTO dashboard(Long userId, String term) {
        List<Student> students = classStudents(userId);
        List<Long> studentIds = students.stream().map(Student::getId).toList();
        String resolvedTerm = resolveTerm(studentIds, term);
        List<GpaHistory> histories = historiesForTerm(studentIds, resolvedTerm);
        List<AlertItemDTO> alerts = termAlerts(students, resolvedTerm);

        return new TeacherDashboardDTO(
                students.size(),
                average(histories.stream().map(GpaHistory::getGpa).filter(Objects::nonNull).toList()),
                countActive(alerts),
                passRate(studentIds, resolvedTerm),
                distribution(alerts),
                dimensionAvg(studentIds, resolvedTerm),
                alerts.stream().limit(3).toList(),
                gradeDistribution(histories));
    }

    public GpaProgressDTO gpaProgress(Long userId, String term) {
        List<Student> students = classStudents(userId);
        List<Long> studentIds = students.stream().map(Student::getId).toList();
        List<GpaHistory> allHistories = gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .in(GpaHistory::getStudentId, studentIds)
                .orderByAsc(GpaHistory::getTerm));
        List<GpaHistory> histories = term == null || term.isBlank()
                ? allHistories
                : allHistories.stream().filter(h -> h.getTerm().compareTo(term) <= 0).toList();
        List<String> terms = histories.stream().map(GpaHistory::getTerm).distinct().toList();
        Map<Long, User> users = loadUsers(students.stream().map(Student::getUserId).toList());

        List<GpaProgressStudentDTO> rows = students.stream().map(student -> {
            Map<String, BigDecimal> byTerm = histories.stream()
                    .filter(h -> h.getStudentId().equals(student.getId()))
                    .collect(Collectors.toMap(GpaHistory::getTerm, GpaHistory::getGpa, (a, b) -> a, LinkedHashMap::new));
            User user = users.get(student.getUserId());
            return new GpaProgressStudentDTO(
                    String.valueOf(student.getId()),
                    user == null ? "" : user.getName(),
                    student.getStudentId(),
                    byTerm);
        }).toList();
        return new GpaProgressDTO(terms, rows);
    }

    public List<CourseStatsItemDTO> courseStats(Long userId, String term) {
        Teacher teacher = requireAdvisor(userId);
        StudentClass studentClass = teacher.getClassId() == null ? null
                : studentClassMapper.selectById(teacher.getClassId());
        if (studentClass == null) {
            return List.of();
        }
        List<Long> studentIds = classStudents(userId).stream().map(Student::getId).toList();
        String resolvedTerm = resolveTerm(studentIds, term);
        LambdaQueryWrapper<CourseClass> wrapper = new LambdaQueryWrapper<CourseClass>()
                .eq(CourseClass::getClassName, studentClass.getName())
                .eq(CourseClass::getTerm, resolvedTerm);
        List<CourseClass> courseClasses = courseClassMapper.selectList(wrapper);
        if (courseClasses.isEmpty()) {
            return List.of();
        }
        Map<Long, Long> classCourse = courseClasses.stream()
                .collect(Collectors.toMap(CourseClass::getId, CourseClass::getCourseId));
        List<Long> courseClassIds = courseClasses.stream().map(CourseClass::getId).toList();
        List<Grade> grades = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                .in(Grade::getCourseClassId, courseClassIds));
        Map<Long, List<Grade>> byCourse = grades.stream().collect(Collectors.groupingBy(
                grade -> classCourse.getOrDefault(grade.getCourseClassId(), -1L)));
        Map<Long, Course> courses = courseMapper.selectBatchIds(byCourse.keySet()).stream()
                .collect(Collectors.toMap(Course::getId, course -> course));

        Map<Long, Student> studentById = studentMapper.selectBatchIds(
                        grades.stream().map(Grade::getStudentId).distinct().toList()).stream()
                .collect(Collectors.toMap(Student::getId, student -> student));
        Map<Long, User> users = loadUsers(studentById.values().stream()
                .map(Student::getUserId).distinct().toList());

        return byCourse.entrySet().stream().sorted(Map.Entry.comparingByKey()).map(entry -> {
            Course course = courses.get(entry.getKey());
            List<Grade> courseGrades = entry.getValue();
            int total = courseGrades.size();
            int passed = (int) courseGrades.stream().filter(g -> "passed".equals(g.getStatus())).count();
            int failed = total - passed;
            List<FailStudentDTO> failStudents = courseGrades.stream()
                    .filter(g -> "failed".equals(g.getStatus()))
                    .map(g -> {
                        Student student = studentById.get(g.getStudentId());
                        User user = student == null ? null : users.get(student.getUserId());
                        return new FailStudentDTO(
                                user == null ? "" : user.getName(),
                                student == null ? "" : student.getStudentId(),
                                g.getScore());
                    })
                    .sorted(Comparator.comparing(FailStudentDTO::score))
                    .toList();
            return new CourseStatsItemDTO(
                    course == null ? String.valueOf(entry.getKey()) : String.valueOf(course.getId()),
                    course == null ? "未知课程" : course.getName(),
                    course == null ? BigDecimal.ZERO : course.getCredits(),
                    avgScore(courseGrades),
                    percent(passed, total, 1),
                    percent(failed, total, 1),
                    failStudents);
        }).toList();
    }

    public PageResult<StudentDTO> students(Long userId, String keyword, String term, int page, int size) {
        List<Student> students = classStudents(userId);
        String resolvedTerm = resolveTerm(students.stream().map(Student::getId).toList(), term);
        List<StudentDTO> dtos = students.stream()
                .map(student -> toTermStudentDTO(student, resolvedTerm))
                .filter(dto -> keyword == null || keyword.isBlank()
                        || dto.name().contains(keyword) || dto.studentId().contains(keyword))
                .toList();
        int from = Math.min((page - 1) * size, dtos.size());
        int to = Math.min(from + size, dtos.size());
        return PageResult.of(dtos.subList(from, to), dtos.size(), page, size);
    }

    public StudentDetailDTO studentDetail(Long userId, String studentId, String term) {
        List<Student> students = classStudents(userId);
        Student student = students.stream()
                .filter(s -> s.getStudentId().equals(studentId))
                .findFirst()
                .orElseThrow(() -> new BusinessException(403, "无权限查看该学生"));
        User user = userMapper.selectById(student.getUserId());
        String grade = user == null ? null : user.getGrade();
        String resolvedTerm = resolveTerm(List.of(student.getId()), term);

        List<AlertItemDTO> alerts = termAlerts(List.of(student), resolvedTerm);
        List<GpaHistory> histories = gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .eq(GpaHistory::getStudentId, student.getId())
                .le(GpaHistory::getTerm, resolvedTerm)
                .orderByAsc(GpaHistory::getTerm));

        return new StudentDetailDTO(
                toTermStudentDTO(student, resolvedTerm),
                buildProfile(student.getId(), resolvedTerm, grade),
                alerts,
                toGpaTrend(histories, grade));
    }

    private AcademicProfileDTO buildProfile(Long studentId, String term, String grade) {
        List<ProfileScore> scores = profileScoreMapper.selectList(new LambdaQueryWrapper<ProfileScore>()
                .eq(ProfileScore::getStudentId, studentId)
                .eq(ProfileScore::getTerm, term))
                .stream()
                .sorted(ProfileScore.byDisplayOrder())
                .toList();
        List<ProfileDimensionDTO> dimensions = scores.stream().map(score -> new ProfileDimensionDTO(
                score.getDimensionKey(),
                score.getLabel(),
                score.getScore(),
                score.getAvgScore(),
                score.getMaxScore(),
                score.getDescription(),
                score.getDetails() == null ? List.of() : score.getDetails()
        )).toList();
        List<GpaHistory> histories = gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .eq(GpaHistory::getStudentId, studentId)
                .le(GpaHistory::getTerm, term)
                .orderByAsc(GpaHistory::getTerm));
        return new AcademicProfileDTO(
                dimensions,
                toGpaTrend(histories, grade),
                histories.stream().map(h -> new RankTrendItemDTO(
                        termLabel(h.getTerm(), grade), h.getRank(), h.getTotalStudents())).toList());
    }

    private List<GpaTrendItemDTO> toGpaTrend(List<GpaHistory> histories, String grade) {
        return histories.stream()
                .map(h -> new GpaTrendItemDTO(termLabel(h.getTerm(), grade), h.getGpa(), h.getAvgGpa()))
                .toList();
    }

    private List<Student> classStudents(Long userId) {
        Teacher teacher = requireAdvisor(userId);
        if (teacher.getClassId() == null) {
            return List.of();
        }
        return studentMapper.selectList(new LambdaQueryWrapper<Student>()
                .eq(Student::getClassId, teacher.getClassId()));
    }

    private Teacher requireAdvisor(Long userId) {
        Teacher teacher = teacherMapper.selectOne(new LambdaQueryWrapper<Teacher>()
                .eq(Teacher::getUserId, userId)
                .eq(Teacher::getIsClassAdvisor, 1));
        if (teacher == null) {
            throw new BusinessException(403, "当前账号不是班主任");
        }
        return teacher;
    }

    private int countActive(List<AlertItemDTO> alerts) {
        return (int) alerts.stream()
                .filter(a -> "pending".equals(a.status()) || "processing".equals(a.status()))
                .map(AlertItemDTO::studentId)
                .distinct()
                .count();
    }

    private TeacherDashboardDTO.AlertDistribution distribution(List<AlertItemDTO> alerts) {
        int yellow = 0;
        int orange = 0;
        int red = 0;
        for (AlertItemDTO alert : alerts) {
            switch (alert.level()) {
                case "yellow" -> yellow++;
                case "orange" -> orange++;
                case "red" -> red++;
                default -> {
                }
            }
        }
        return new TeacherDashboardDTO.AlertDistribution(yellow, orange, red);
    }

    private List<TeacherDashboardDTO.DimensionAvgItem> dimensionAvg(List<Long> studentIds, String term) {
        if (studentIds.isEmpty()) {
            return List.of();
        }
        List<ProfileScore> scores = profileScoreMapper.selectList(new LambdaQueryWrapper<ProfileScore>()
                .in(ProfileScore::getStudentId, studentIds)
                .eq(ProfileScore::getTerm, term))
                .stream()
                .sorted(ProfileScore.byDisplayOrder())
                .toList();
        Map<String, List<Integer>> byLabel = new LinkedHashMap<>();
        for (ProfileScore score : scores) {
            byLabel.computeIfAbsent(score.getLabel(), k -> new ArrayList<>()).add(score.getScore());
        }
        return byLabel.entrySet().stream()
                .map(e -> new TeacherDashboardDTO.DimensionAvgItem(
                        e.getKey(),
                        (int) Math.round(e.getValue().stream().mapToInt(Integer::intValue).average().orElse(0))))
                .toList();
    }

    private List<TeacherDashboardDTO.GradeDistributionItem> gradeDistribution(List<GpaHistory> histories) {
        int excellent = 0;
        int good = 0;
        int medium = 0;
        int warning = 0;
        for (GpaHistory history : histories) {
            double gpa = history.getGpa() == null ? 0 : history.getGpa().doubleValue();
            if (gpa >= 3.5) {
                excellent++;
            } else if (gpa >= 3.0) {
                good++;
            } else if (gpa >= 2.5) {
                medium++;
            } else {
                warning++;
            }
        }
        return List.of(
                new TeacherDashboardDTO.GradeDistributionItem("优秀(>=3.5)", excellent),
                new TeacherDashboardDTO.GradeDistributionItem("良好(3.0-3.5)", good),
                new TeacherDashboardDTO.GradeDistributionItem("中等(2.5-3.0)", medium),
                new TeacherDashboardDTO.GradeDistributionItem("预警(<2.5)", warning));
    }

    private BigDecimal passRate(List<Long> studentIds, String term) {
        if (studentIds.isEmpty()) {
            return BigDecimal.ZERO;
        }
        List<Grade> grades = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                .in(Grade::getStudentId, studentIds)
                .eq(Grade::getTerm, term));
        int passed = (int) grades.stream().filter(g -> "passed".equals(g.getStatus())).count();
        return percent(passed, grades.size(), 1);
    }

    private BigDecimal avgScore(List<Grade> grades) {
        return average(grades.stream().map(Grade::getScore).filter(Objects::nonNull).toList(), 1);
    }

    private BigDecimal average(List<BigDecimal> values) {
        return average(values, 2);
    }

    private BigDecimal average(List<BigDecimal> values, int scale) {
        if (values.isEmpty()) {
            return BigDecimal.ZERO.setScale(scale);
        }
        double sum = values.stream().mapToDouble(BigDecimal::doubleValue).sum();
        return BigDecimal.valueOf(sum / values.size()).setScale(scale, RoundingMode.HALF_UP);
    }

    private BigDecimal percent(int part, int total, int scale) {
        if (total == 0) {
            return BigDecimal.ZERO.setScale(scale);
        }
        return BigDecimal.valueOf(part * 100.0 / total).setScale(scale, RoundingMode.HALF_UP);
    }

    private String resolveTerm(List<Long> studentIds, String term) {
        if (term != null && !term.isBlank()) {
            return term;
        }
        List<String> terms = availableTerms(studentIds);
        return terms.isEmpty() ? "" : terms.get(0);
    }

    private List<String> availableTerms(List<Long> studentIds) {
        if (studentIds.isEmpty()) {
            return List.of();
        }
        List<String> gradeTerms = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                        .in(Grade::getStudentId, studentIds))
                .stream().map(Grade::getTerm).toList();
        List<String> historyTerms = gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                        .in(GpaHistory::getStudentId, studentIds))
                .stream().map(GpaHistory::getTerm).toList();
        return java.util.stream.Stream.concat(gradeTerms.stream(), historyTerms.stream())
                .filter(value -> value != null && !value.isBlank())
                .distinct().sorted(Comparator.reverseOrder()).toList();
    }

    private List<GpaHistory> historiesForTerm(List<Long> studentIds, String term) {
        if (studentIds.isEmpty() || term == null || term.isBlank()) {
            return List.of();
        }
        return gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .in(GpaHistory::getStudentId, studentIds)
                .eq(GpaHistory::getTerm, term));
    }

    private StudentDTO toTermStudentDTO(Student student, String term) {
        StudentDTO current = studentService.toStudentDTO(student);
        GpaHistory history = gpaHistoryMapper.selectOne(new LambdaQueryWrapper<GpaHistory>()
                .eq(GpaHistory::getStudentId, student.getId())
                .eq(GpaHistory::getTerm, term));
        List<AlertItemDTO> alerts = termAlerts(List.of(student), term);
        return new StudentDTO(
                current.id(), current.name(), current.studentId(), current.college(), current.major(),
                current.className(), current.grade(), history == null ? BigDecimal.ZERO : history.getGpa(),
                history == null ? null : history.getRank(),
                history == null ? null : history.getTotalStudents(),
                alerts.isEmpty() ? "none" : alerts.get(0).level(), current.avatar());
    }

    private List<AlertItemDTO> termAlerts(List<Student> students, String term) {
        if (students.isEmpty() || term == null || term.isBlank()) {
            return List.of();
        }
        List<Long> studentIds = students.stream().map(Student::getId).toList();
        Map<Long, Integer> failedByStudent = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                        .in(Grade::getStudentId, studentIds)
                        .eq(Grade::getTerm, term))
                .stream().filter(grade -> "failed".equals(grade.getStatus())
                        || (grade.getScore() != null && grade.getScore().compareTo(BigDecimal.valueOf(60)) < 0))
                .collect(Collectors.groupingBy(Grade::getStudentId, Collectors.summingInt(grade -> 1)));
        Map<Long, Integer> absentByStudent = attendanceMapper.selectList(new LambdaQueryWrapper<Attendance>()
                        .in(Attendance::getStudentId, studentIds)
                        .eq(Attendance::getTerm, term))
                .stream().collect(Collectors.groupingBy(Attendance::getStudentId,
                        Collectors.summingInt(attendance -> attendance.getAbsentCount() == null ? 0 : attendance.getAbsentCount())));
        Map<Long, User> users = loadUsers(students.stream().map(Student::getUserId).toList());

        return students.stream().map(student -> {
                    int failed = failedByStudent.getOrDefault(student.getId(), 0);
                    int absences = absentByStudent.getOrDefault(student.getId(), 0);
                    AcademicAlertPolicy.Assessment assessment = AcademicAlertPolicy.assess(failed, absences);
                    if ("none".equals(assessment.level())) {
                        return null;
                    }
                    User user = users.get(student.getUserId());
                    String levelName = switch (assessment.level()) {
                        case "red" -> "红色";
                        case "orange" -> "橙色";
                        default -> "黄色";
                    };
                    return new AlertItemDTO(
                            term + "-" + student.getId(), student.getStudentId(),
                            user == null ? "" : user.getName(), assessment.primaryRisk(), assessment.level(),
                            levelName + "预警", "当学期挂科 " + failed + " 门，缺勤 " + absences + " 次",
                            null, List.of(), term, "pending",
                            "请结合课程表现和考勤情况及时跟进", "挂科 " + failed + " 门；缺勤 " + absences + " 次", term);
                })
                .filter(Objects::nonNull)
                .sorted(Comparator.comparingInt((AlertItemDTO alert) -> alertSeverity(alert.level())).reversed())
                .toList();
    }

    private int alertSeverity(String level) {
        return switch (level) {
            case "red" -> 3;
            case "orange" -> 2;
            case "yellow" -> 1;
            default -> 0;
        };
    }

    private Map<Long, User> loadUsers(List<Long> userIds) {
        if (userIds.isEmpty()) {
            return Map.of();
        }
        return userMapper.selectBatchIds(userIds).stream()
                .collect(Collectors.toMap(User::getId, user -> user));
    }

    private String termLabel(String term, String grade) {
        if (term == null || term.length() < 9) {
            return term;
        }
        int year;
        int gradeYear;
        try {
            year = Integer.parseInt(term.substring(0, 4));
            gradeYear = Integer.parseInt(grade);
        } catch (Exception e) {
            return term;
        }
        int index = year - gradeYear + 1;
        String[] names = {"大一", "大二", "大三", "大四"};
        String suffix = term.endsWith("2") ? "下" : "上";
        if (index >= 1 && index <= names.length) {
            return names[index - 1] + suffix;
        }
        return term;
    }
}
