package com.zjnu.academic.department.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zjnu.academic.alert.entity.Alert;
import com.zjnu.academic.alert.mapper.AlertMapper;
import com.zjnu.academic.alert.service.AlertStudentAggregator;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.common.term.AcademicTerm;
import com.zjnu.academic.course.entity.Course;
import com.zjnu.academic.course.entity.CourseClass;
import com.zjnu.academic.course.entity.Grade;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.department.dto.CourseHeatmapDTO;
import com.zjnu.academic.department.dto.DepartmentAlertDTO;
import com.zjnu.academic.department.dto.DepartmentCourseDTO;
import com.zjnu.academic.department.dto.DepartmentOverviewDTO;
import com.zjnu.academic.department.dto.PlanAnalysisDTO;
import com.zjnu.academic.department.entity.Department;
import com.zjnu.academic.department.entity.MajorPlan;
import com.zjnu.academic.department.mapper.DepartmentMapper;
import com.zjnu.academic.department.mapper.MajorPlanMapper;
import com.zjnu.academic.major.entity.Major;
import com.zjnu.academic.major.mapper.MajorMapper;
import com.zjnu.academic.student.dto.GpaTrendItemDTO;
import com.zjnu.academic.student.entity.GpaHistory;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.LinkedHashSet;
import java.util.Map;
import java.util.Objects;
import java.util.TreeMap;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

@Service
public class DepartmentService {

    private static final DateTimeFormatter DATETIME = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm");
    private static final Pattern ABSENCE_PATTERN = Pattern.compile("(\\d+)\\s*次");

    private final UserMapper userMapper;
    private final MajorMapper majorMapper;
    private final StudentMapper studentMapper;
    private final GpaHistoryMapper gpaHistoryMapper;
    private final GradeMapper gradeMapper;
    private final CourseClassMapper courseClassMapper;
    private final CourseMapper courseMapper;
    private final AlertMapper alertMapper;
    private final MajorPlanMapper majorPlanMapper;
    private final DepartmentMapper departmentMapper;

    @Autowired
    public DepartmentService(UserMapper userMapper,
                             MajorMapper majorMapper,
                             StudentMapper studentMapper,
                             GpaHistoryMapper gpaHistoryMapper,
                             GradeMapper gradeMapper,
                             CourseClassMapper courseClassMapper,
                             CourseMapper courseMapper,
                             AlertMapper alertMapper,
                             MajorPlanMapper majorPlanMapper,
                             DepartmentMapper departmentMapper) {
        this.userMapper = userMapper;
        this.majorMapper = majorMapper;
        this.studentMapper = studentMapper;
        this.gpaHistoryMapper = gpaHistoryMapper;
        this.gradeMapper = gradeMapper;
        this.courseClassMapper = courseClassMapper;
        this.courseMapper = courseMapper;
        this.alertMapper = alertMapper;
        this.majorPlanMapper = majorPlanMapper;
        this.departmentMapper = departmentMapper;
    }

    public DepartmentOverviewDTO overview(Long userId, String term) {
        Major major = requireMajor(userId);
        List<Student> students = scopeStudents(major.getId());
        List<Long> studentIds = students.stream().map(Student::getId).toList();
        String resolvedTerm = resolveTerm(studentIds, term);
        List<GpaHistory> termHistories = loadHistoriesForTerm(studentIds, resolvedTerm);
        Map<Long, GpaHistory> historyByStudent = termHistories.stream()
                .collect(Collectors.toMap(GpaHistory::getStudentId, history -> history, (a, b) -> a));
        List<Alert> alerts = AlertStudentAggregator.perStudent(
                loadAlerts(studentIds, resolvedTerm, null, null));
        Map<Long, User> users = loadUsers(students.stream().map(Student::getUserId).toList());

        int warningStudents = alerts.size();

        return new DepartmentOverviewDTO(
                students.size(),
                average(termHistories.stream().map(GpaHistory::getGpa).filter(Objects::nonNull).toList(), 2),
                percent(warningStudents, students.size(), 1),
                passRate(studentIds, resolvedTerm),
                gradeGpaTrend(students, users, resolvedTerm),
                alertDistribution(alerts),
                focusStudents(students, users, major.getName(), historyByStudent, alerts));
    }

    public List<String> availableTerms(Long userId) {
        Major major = requireMajor(userId);
        List<Long> studentIds = scopeStudents(major.getId()).stream().map(Student::getId).toList();
        return availableTerms(studentIds);
    }

    public CourseHeatmapDTO courseHeatmap(Long userId, String term) {
        Major major = requireMajor(userId);
        List<Student> students = scopeStudents(major.getId());
        List<Long> studentIds = students.stream().map(Student::getId).toList();
        String resolvedTerm = resolveTerm(studentIds, term);
        List<Grade> grades = loadGrades(studentIds, resolvedTerm);
        if (grades.isEmpty()) {
            return new CourseHeatmapDTO(List.of(), List.of(), List.of());
        }

        List<Long> classIds = grades.stream().map(Grade::getCourseClassId).distinct().toList();
        Map<Long, Long> classToCourse = courseClassMapper.selectBatchIds(classIds).stream()
                .collect(Collectors.toMap(CourseClass::getId, CourseClass::getCourseId, (a, b) -> a));
        List<Long> courseIds = classToCourse.values().stream().distinct().toList();
        LinkedHashSet<String> seenCourses = new LinkedHashSet<>();
        List<Course> courses = courseMapper.selectBatchIds(courseIds).stream()
                .sorted(Comparator.comparing(Course::getId))
                .filter(c -> seenCourses.add(c.getName()))
                .toList();

        Map<Long, User> users = loadUsers(students.stream().map(Student::getUserId).toList());
        Map<Long, String> studentGrade = students.stream()
                .collect(Collectors.toMap(Student::getId,
                        s -> {
                            User user = users.get(s.getUserId());
                            return user == null || user.getGrade() == null ? "" : user.getGrade() + "级";
                        }, (a, b) -> a, TreeMap::new));
        List<String> gradeLabels = studentGrade.values().stream().distinct().sorted().toList();
        List<String> courseNames = courses.stream().map(Course::getName).toList();

        List<int[]> heatData = new ArrayList<>();
        for (int ci = 0; ci < courses.size(); ci++) {
            Course course = courses.get(ci);
            for (int gi = 0; gi < gradeLabels.size(); gi++) {
                String gradeLabel = gradeLabels.get(gi);
                List<Grade> matched = grades.stream()
                        .filter(g -> Objects.equals(classToCourse.get(g.getCourseClassId()), course.getId()))
                        .filter(g -> gradeLabel.equals(studentGrade.get(g.getStudentId())))
                        .toList();
                if (!matched.isEmpty()) {
                    int passed = (int) matched.stream().filter(g -> "passed".equals(g.getStatus())).count();
                    int value = (int) Math.round(passed * 100.0 / matched.size());
                    heatData.add(new int[]{gi, ci, value});
                }
            }
        }
        return new CourseHeatmapDTO(courseNames, gradeLabels, heatData);
    }

    public List<DepartmentCourseDTO> courses(Long userId, String term) {
        Major major = requireMajor(userId);
        List<Student> students = scopeStudents(major.getId());
        List<Long> studentIds = students.stream().map(Student::getId).toList();
        String resolvedTerm = resolveTerm(studentIds, term);
        List<Grade> grades = loadGrades(studentIds, resolvedTerm);
        if (grades.isEmpty()) {
            return List.of();
        }

        List<Long> classIds = grades.stream().map(Grade::getCourseClassId).distinct().toList();
        Map<Long, Long> classToCourse = courseClassMapper.selectBatchIds(classIds).stream()
                .collect(Collectors.toMap(CourseClass::getId, CourseClass::getCourseId, (a, b) -> a));
        Map<Long, User> users = loadUsers(students.stream().map(Student::getUserId).toList());
        Map<Long, String> studentGrade = students.stream()
                .collect(Collectors.toMap(Student::getId,
                        s -> {
                            User user = users.get(s.getUserId());
                            return user == null || user.getGrade() == null ? "" : user.getGrade() + "级";
                        }, (a, b) -> a));
        Map<Long, Course> courseById = courseMapper.selectBatchIds(classToCourse.values()).stream()
                .collect(Collectors.toMap(Course::getId, course -> course));

        Map<Long, List<Grade>> byCourse = grades.stream()
                .collect(Collectors.groupingBy(g -> classToCourse.getOrDefault(g.getCourseClassId(), -1L)));
        return byCourse.entrySet().stream()
                .sorted(Map.Entry.comparingByKey())
                .map(entry -> {
                    Course course = courseById.get(entry.getKey());
                    List<Grade> courseGrades = entry.getValue();
                    return new DepartmentCourseDTO(
                            course == null ? String.valueOf(entry.getKey()) : String.valueOf(course.getId()),
                            course == null ? "未知课程" : course.getName(),
                            course == null ? "" : course.getCode(),
                            course == null ? BigDecimal.ZERO : course.getCredits(),
                            avgScore(courseGrades),
                            passRate(courseGrades),
                            highRate(courseGrades),
                            lowRate(courseGrades),
                            gradeStats(courseGrades, studentGrade));
                })
                .toList();
    }

    public PlanAnalysisDTO planAnalysis(Long userId, String grade) {
        Major major = requireMajor(userId);
        String resolvedGrade = grade == null || grade.isBlank() ? "2022" : grade.replace("级", "");
        MajorPlan plan = majorPlanMapper.selectOne(new LambdaQueryWrapper<MajorPlan>()
                .eq(MajorPlan::getMajorId, major.getId())
                .eq(MajorPlan::getGrade, resolvedGrade));
        if (plan == null) {
            return new PlanAnalysisDTO(
                    resolvedGrade,
                    List.of("公共基础", "专业必修", "专业选修", "实践环节", "通识教育"),
                    List.of(48, 62, 20, 18, 12),
                    List.of(0, 0, 0, 0, 0),
                    List.of("暂无该年级培养方案数据"));
        }
        return new PlanAnalysisDTO(
                resolvedGrade,
                plan.getDimensions() == null ? List.of("公共基础", "专业必修", "专业选修", "实践环节", "通识教育") : plan.getDimensions(),
                plan.getStandard() == null ? List.of() : plan.getStandard(),
                plan.getActual() == null ? List.of() : plan.getActual(),
                plan.getTips() == null ? List.of() : plan.getTips());
    }

    public PageResult<DepartmentAlertDTO> alerts(Long userId, String term, String grade, String level,
                                                 String status, int page, int size) {
        Major major = requireMajor(userId);
        List<Student> students = scopeStudents(major.getId());
        Map<Long, User> users = loadUsers(students.stream().map(Student::getUserId).toList());

        if (grade != null && !grade.isBlank()) {
            String expected = grade.replace("级", "");
            students = students.stream()
                    .filter(s -> expected.equals(users.get(s.getUserId()) == null ? null : users.get(s.getUserId()).getGrade()))
                    .toList();
        }
        List<Long> studentIds = students.stream().map(Student::getId).toList();
        String resolvedTerm = resolveTerm(studentIds, term);
        List<Alert> alerts = AlertStudentAggregator.perStudent(
                loadAlerts(studentIds, resolvedTerm, null, status));
        if (level != null && !level.isBlank()) {
            alerts = alerts.stream()
                    .filter(alert -> level.equals(alert.getLevel()))
                    .toList();
        }
        Map<Long, Student> studentById = students.stream()
                .collect(Collectors.toMap(Student::getId, s -> s));
        Map<Long, List<GpaHistory>> histories = loadHistoriesMap(studentIds);

        List<DepartmentAlertDTO> dtos = alerts.stream()
                .map(alert -> toAlertDto(alert, studentById, users, histories, resolvedTerm))
                .sorted(Comparator.comparing(DepartmentAlertDTO::date).reversed())
                .toList();
        int from = Math.min((page - 1) * size, dtos.size());
        int to = Math.min(from + size, dtos.size());
        return PageResult.of(dtos.subList(from, to), dtos.size(), page, size);
    }

    private List<DepartmentOverviewDTO.GradeGpaTrendItemDTO> gradeGpaTrend(
            List<Student> students, Map<Long, User> users, String term) {
        Map<String, List<Student>> byGrade = new TreeMap<>();
        for (Student student : students) {
            User user = users.get(student.getUserId());
            if (user == null || user.getGrade() == null) {
                continue;
            }
            byGrade.computeIfAbsent(user.getGrade(), k -> new ArrayList<>()).add(student);
        }
        List<DepartmentOverviewDTO.GradeGpaTrendItemDTO> result = new ArrayList<>();
        byGrade.forEach((grade, gradeStudents) -> {
            List<Long> ids = gradeStudents.stream().map(Student::getId).toList();
            List<GpaHistory> histories = gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                    .in(GpaHistory::getStudentId, ids)
                    .eq(GpaHistory::getTerm, term));
            List<GpaTrendItemDTO> trend = histories.stream()
                    .collect(Collectors.groupingBy(GpaHistory::getTerm, TreeMap::new, Collectors.toList()))
                    .entrySet().stream()
                    .map(entry -> new GpaTrendItemDTO(
                            entry.getKey(),
                            average(entry.getValue().stream().map(GpaHistory::getGpa)
                                    .filter(Objects::nonNull).toList(), 2),
                            average(entry.getValue().stream().map(GpaHistory::getAvgGpa)
                                    .filter(Objects::nonNull).toList(), 2)))
                    .toList();
            result.add(new DepartmentOverviewDTO.GradeGpaTrendItemDTO(grade + "级", trend));
        });
        return result;
    }

    private DepartmentOverviewDTO.AlertDistributionDTO alertDistribution(List<Alert> alerts) {
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
        return new DepartmentOverviewDTO.AlertDistributionDTO(yellow, orange, red);
    }

    private List<DepartmentOverviewDTO.FocusStudentDTO> focusStudents(
            List<Student> students, Map<Long, User> users, String majorName,
            Map<Long, GpaHistory> historyByStudent, List<Alert> alerts) {
        Map<String, Integer> severity = Map.of("red", 0, "orange", 1, "yellow", 2);
        Map<Long, Alert> alertByStudent = alerts.stream()
                .collect(Collectors.toMap(Alert::getStudentId, alert -> alert, (a, b) -> a));
        return students.stream()
                .filter(s -> alertByStudent.containsKey(s.getId()))
                .sorted(Comparator
                        .comparing((Student s) -> severity.getOrDefault(alertByStudent.get(s.getId()).getLevel(), 9))
                        .thenComparing(s -> {
                            GpaHistory history = historyByStudent.get(s.getId());
                            return history == null || history.getGpa() == null ? BigDecimal.ZERO : history.getGpa();
                        }))
                .limit(8)
                .map(s -> {
                    User user = users.get(s.getUserId());
                    GpaHistory history = historyByStudent.get(s.getId());
                    Alert alert = alertByStudent.get(s.getId());
                    return new DepartmentOverviewDTO.FocusStudentDTO(
                            String.valueOf(s.getId()),
                            user == null ? "" : user.getName(),
                            s.getStudentId(),
                            user == null ? "" : (user.getGrade() == null ? "" : user.getGrade() + "级"),
                            majorName,
                            history == null ? BigDecimal.ZERO : history.getGpa(),
                            alert.getLevel(),
                            user == null ? "" : user.getClassName());
                })
                .toList();
    }

    private List<DepartmentCourseDTO.GradeStatDTO> gradeStats(
            List<Grade> courseGrades, Map<Long, String> studentGrade) {
        Map<String, List<Grade>> byGrade = new TreeMap<>();
        for (Grade grade : courseGrades) {
            String label = studentGrade.getOrDefault(grade.getStudentId(), "");
            byGrade.computeIfAbsent(label, k -> new ArrayList<>()).add(grade);
        }
        return byGrade.entrySet().stream()
                .map(entry -> new DepartmentCourseDTO.GradeStatDTO(
                        entry.getKey(),
                        avgScore(entry.getValue()),
                        passRate(entry.getValue()),
                        highRate(entry.getValue()),
                        lowRate(entry.getValue())))
                .toList();
    }

    private DepartmentAlertDTO toAlertDto(Alert alert, Map<Long, Student> students,
                                          Map<Long, User> users, Map<Long, List<GpaHistory>> histories,
                                          String term) {
        Student student = students.get(alert.getStudentId());
        User user = student == null ? null : users.get(student.getUserId());
        String description = alert.getDescription() == null ? "" : alert.getDescription();
        List<GpaHistory> studentHistories = student == null ? List.of()
                : histories.getOrDefault(student.getId(), List.of()).stream()
                .filter(history -> history.getTerm().compareTo(term) <= 0)
                .toList();
        GpaHistory selectedHistory = studentHistories.stream()
                .filter(history -> term.equals(history.getTerm()))
                .findFirst()
                .orElse(null);
        return new DepartmentAlertDTO(
                String.valueOf(alert.getId()),
                student == null ? "" : student.getStudentId(),
                user == null ? "" : user.getName(),
                user == null || user.getGrade() == null ? "" : user.getGrade() + "级",
                user == null ? "" : user.getClassName(),
                user == null ? "" : user.getMajor(),
                selectedHistory == null ? BigDecimal.ZERO : selectedHistory.getGpa(),
                trendOf(studentHistories),
                alert.getLevel(),
                absencesOf(description),
                description,
                alert.getType(),
                alert.getTitle(),
                description,
                alert.getCourse(),
                alert.getFailedCourses() == null ? List.of() : alert.getFailedCourses(),
                alert.getTriggerDate() == null ? "" : alert.getTriggerDate().toString(),
                alert.getStatus(),
                alert.getSuggestion(),
                alert.getTriggerEvent(),
                alert.getPushedAt() == null ? null : DATETIME.format(alert.getPushedAt()));
    }

    private List<GpaHistory> loadHistories(List<Long> studentIds) {
        if (studentIds.isEmpty()) {
            return List.of();
        }
        return gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .in(GpaHistory::getStudentId, studentIds)
                .orderByAsc(GpaHistory::getTerm));
    }

    private Map<Long, List<GpaHistory>> loadHistoriesMap(List<Long> studentIds) {
        return loadHistories(studentIds).stream()
                .collect(Collectors.groupingBy(GpaHistory::getStudentId));
    }

    private List<GpaHistory> loadHistoriesForTerm(List<Long> studentIds, String term) {
        if (studentIds.isEmpty() || term == null || term.isBlank()) {
            return List.of();
        }
        return gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .in(GpaHistory::getStudentId, studentIds)
                .eq(GpaHistory::getTerm, term));
    }

    private Major requireMajor(Long userId) {
        User user = userMapper.selectById(userId);
        if (user == null) {
            throw new BusinessException(403, "当前账号未配置专业数据权限");
        }
        Department department = null;
        if (user.getCollege() != null && !user.getCollege().isBlank()) {
            department = departmentMapper.selectOne(new LambdaQueryWrapper<Department>()
                    .eq(Department::getName, user.getCollege()));
        }
        if (department == null) {
            department = departmentMapper.selectById(1L);
        }
        if (department == null) {
            throw new BusinessException(403, "未找到所属学院数据");
        }
        Major major = null;
        if (user.getMajor() != null && !user.getMajor().isBlank()) {
            major = majorMapper.selectOne(new LambdaQueryWrapper<Major>()
                    .eq(Major::getName, user.getMajor())
                    .eq(Major::getDepartmentId, department.getId()));
        }
        if (major == null) {
            // 未绑定专业或专业未归属本学院时，回退到账号所属学院第一个专业
            major = majorMapper.selectList(new LambdaQueryWrapper<Major>()
                            .eq(Major::getDepartmentId, department.getId())
                            .orderByAsc(Major::getId))
                    .stream()
                    .findFirst()
                    .orElse(null);
        }
        if (major == null) {
            throw new BusinessException(403, "未找到所属专业数据");
        }
        return major;
    }

    private List<Student> scopeStudents(Long majorId) {
        return studentMapper.selectList(new LambdaQueryWrapper<Student>()
                .eq(Student::getMajorId, majorId));
    }

    private List<Alert> loadAlerts(List<Long> studentIds, String term, String level, String status) {
        if (studentIds.isEmpty()) {
            return List.of();
        }
        LambdaQueryWrapper<Alert> wrapper = new LambdaQueryWrapper<Alert>()
                .in(Alert::getStudentId, studentIds);
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

    private List<Grade> loadGrades(List<Long> studentIds, String term) {
        if (studentIds.isEmpty()) {
            return List.of();
        }
        return gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                .in(Grade::getStudentId, studentIds)
                .eq(Grade::getTerm, term));
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
        List<String> historyTerms = gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                        .in(GpaHistory::getStudentId, studentIds))
                .stream().map(GpaHistory::getTerm).toList();
        List<String> gradeTerms = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                        .in(Grade::getStudentId, studentIds))
                .stream().map(Grade::getTerm).toList();
        return java.util.stream.Stream.concat(historyTerms.stream(), gradeTerms.stream())
                .filter(value -> value != null && !value.isBlank())
                .distinct()
                .sorted(Comparator.reverseOrder())
                .toList();
    }

    private Map<Long, User> loadUsers(List<Long> userIds) {
        if (userIds.isEmpty()) {
            return Map.of();
        }
        return userMapper.selectBatchIds(userIds).stream()
                .collect(Collectors.toMap(User::getId, user -> user));
    }

    private BigDecimal passRate(List<Long> studentIds, String term) {
        List<Grade> grades = loadGrades(studentIds, term);
        return passRate(grades);
    }

    private BigDecimal passRate(List<Grade> grades) {
        int passed = (int) grades.stream().filter(g -> "passed".equals(g.getStatus())).count();
        return percent(passed, grades.size(), 1);
    }

    private BigDecimal highRate(List<Grade> grades) {
        int high = (int) grades.stream()
                .filter(g -> g.getScore() != null && g.getScore().doubleValue() >= 90)
                .count();
        return percent(high, grades.size(), 1);
    }

    private BigDecimal lowRate(List<Grade> grades) {
        int low = (int) grades.stream()
                .filter(g -> g.getScore() != null && g.getScore().doubleValue() < 60)
                .count();
        return percent(low, grades.size(), 1);
    }

    private BigDecimal avgScore(List<Grade> grades) {
        return average(grades.stream().map(Grade::getScore).filter(Objects::nonNull).toList(), 1);
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

    private String trendOf(List<GpaHistory> histories) {
        if (histories.size() < 2) {
            return "flat";
        }
        GpaHistory last = histories.get(histories.size() - 1);
        GpaHistory prev = histories.get(histories.size() - 2);
        double delta = last.getGpa().doubleValue() - prev.getGpa().doubleValue();
        if (delta > 0.05) {
            return "up";
        }
        if (delta < -0.05) {
            return "down";
        }
        return "flat";
    }

    private int absencesOf(String text) {
        if (text == null || text.isBlank()) {
            return 0;
        }
        Matcher matcher = ABSENCE_PATTERN.matcher(text);
        return matcher.find() ? Integer.parseInt(matcher.group(1)) : 0;
    }

}
