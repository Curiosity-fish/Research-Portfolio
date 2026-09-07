package com.zjnu.academic.dean.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zjnu.academic.alert.entity.Alert;
import com.zjnu.academic.alert.mapper.AlertMapper;
import com.zjnu.academic.alert.service.AlertStudentAggregator;
import com.zjnu.academic.course.entity.Grade;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.term.AcademicTerm;
import com.zjnu.academic.dean.dto.DeanAlertStatsDTO;
import com.zjnu.academic.dean.dto.DeanAlertTrendDTO;
import com.zjnu.academic.dean.dto.DeanDashboardDTO;
import com.zjnu.academic.department.entity.Department;
import com.zjnu.academic.department.mapper.DepartmentMapper;
import com.zjnu.academic.major.entity.Major;
import com.zjnu.academic.major.mapper.MajorMapper;
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
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.TreeMap;
import java.util.stream.Collectors;

@Service
public class DeanService {

    private final MajorMapper majorMapper;
    private final StudentMapper studentMapper;
    private final UserMapper userMapper;
    private final GpaHistoryMapper gpaHistoryMapper;
    private final GradeMapper gradeMapper;
    private final AlertMapper alertMapper;
    private final DepartmentMapper departmentMapper;

    @Autowired
    public DeanService(MajorMapper majorMapper,
                       StudentMapper studentMapper,
                       UserMapper userMapper,
                       GpaHistoryMapper gpaHistoryMapper,
                       GradeMapper gradeMapper,
                       AlertMapper alertMapper,
                       DepartmentMapper departmentMapper) {
        this.majorMapper = majorMapper;
        this.studentMapper = studentMapper;
        this.userMapper = userMapper;
        this.gpaHistoryMapper = gpaHistoryMapper;
        this.gradeMapper = gradeMapper;
        this.alertMapper = alertMapper;
        this.departmentMapper = departmentMapper;
    }

    public DeanDashboardDTO dashboard(Long userId, String term) {
        List<Major> majors = collegeMajors(userId);
        List<Long> majorIds = majors.stream().map(Major::getId).toList();
        List<Student> students = collegeStudents(majorIds);
        List<Long> studentIds = students.stream().map(Student::getId).toList();
        String resolvedTerm = resolveTerm(studentIds, term);
        List<Alert> termAlerts = AlertStudentAggregator.perStudent(loadAlerts(studentIds, resolvedTerm));
        List<Grade> grades = loadGrades(studentIds, resolvedTerm);
        List<GpaHistory> histories = loadHistories(studentIds, resolvedTerm);
        Map<Long, GpaHistory> historyByStudent = histories.stream()
                .collect(Collectors.toMap(GpaHistory::getStudentId, history -> history, (a, b) -> a));
        Map<Long, Student> studentById = students.stream()
                .collect(Collectors.toMap(Student::getId, s -> s));
        int warningStudents = termAlerts.size();
        int passed = (int) grades.stream().filter(g -> "passed".equals(g.getStatus())).count();
        int intervened = (int) termAlerts.stream().filter(a -> a.getHandlerId() != null).count();

        return new DeanDashboardDTO(
                students.size(),
                average(histories.stream().map(GpaHistory::getGpa).filter(Objects::nonNull).toList(), 2),
                percent(warningStudents, students.size(), 1),
                percent(passed, grades.size(), 1),
                percent(intervened, termAlerts.size(), 1),
                departmentRanking(majors, students, resolvedTerm, termAlerts, historyByStudent),
                alertHeatmap(majors, termAlerts, studentById));
    }

    public List<String> availableTerms(Long userId) {
        List<Long> studentIds = collegeStudents(collegeMajors(userId).stream().map(Major::getId).toList())
                .stream().map(Student::getId).toList();
        return availableTerms(studentIds);
    }

    public DeanAlertStatsDTO alerts(Long userId, String term, String groupBy) {
        List<Major> majors = collegeMajors(userId);
        List<Long> majorIds = majors.stream().map(Major::getId).toList();
        List<Student> students = collegeStudents(majorIds);
        List<Long> studentIds = students.stream().map(Student::getId).toList();
        String resolvedTerm = resolveTerm(studentIds, term);
        List<Alert> alerts = AlertStudentAggregator.perStudent(loadAlerts(studentIds, resolvedTerm));
        Map<Long, Student> studentById = students.stream()
                .collect(Collectors.toMap(Student::getId, s -> s));
        Map<Long, User> users = loadUsers(students.stream().map(Student::getUserId).toList());
        Map<Long, Major> majorById = majors.stream()
                .collect(Collectors.toMap(Major::getId, major -> major));

        int red = 0;
        int orange = 0;
        int yellow = 0;
        for (Alert alert : alerts) {
            switch (alert.getLevel()) {
                case "red" -> red++;
                case "orange" -> orange++;
                case "yellow" -> yellow++;
                default -> {
                }
            }
        }

        List<DeanAlertStatsDTO.DeptRowDTO> byDept = new ArrayList<>();
        Map<String, List<Student>> studentsByName = groupStudentsByMajorName(students, majorById);
        for (Map.Entry<String, List<Student>> entry : studentsByName.entrySet()) {
            List<Student> majorStudents = entry.getValue();
            List<Long> ids = majorStudents.stream().map(Student::getId).toList();
            List<Alert> majorAlerts = alerts.stream().filter(a -> ids.contains(a.getStudentId())).toList();
            int active = majorAlerts.size();
            byDept.add(new DeanAlertStatsDTO.DeptRowDTO(
                    entry.getKey(),
                    majorStudents.size(),
                    countLevel(majorAlerts, "red"),
                    countLevel(majorAlerts, "orange"),
                    countLevel(majorAlerts, "yellow"),
                    percent(active, majorStudents.size(), 1)));
        }

        Map<String, List<Student>> byGrade = new TreeMap<>();
        for (Student student : students) {
            User user = users.get(student.getUserId());
            if (user == null || user.getGrade() == null) {
                continue;
            }
            byGrade.computeIfAbsent(user.getGrade(), k -> new ArrayList<>()).add(student);
        }
        List<DeanAlertStatsDTO.GradeRowDTO> gradeRows = byGrade.entrySet().stream()
                .map(entry -> {
                    List<Long> ids = entry.getValue().stream().map(Student::getId).toList();
                    List<Alert> gradeAlerts = alerts.stream().filter(a -> ids.contains(a.getStudentId())).toList();
                    return new DeanAlertStatsDTO.GradeRowDTO(
                            entry.getKey() + "级",
                            countLevel(gradeAlerts, "red"),
                            countLevel(gradeAlerts, "orange"),
                            countLevel(gradeAlerts, "yellow"),
                            gradeAlerts.size(),
                            gradeNote(entry.getKey()));
                })
                .toList();

        return new DeanAlertStatsDTO(
                resolvedTerm,
                new DeanAlertStatsDTO.TotalDTO(red, orange, yellow, alerts.size()),
                byDept,
                gradeRows);
    }

    public DeanAlertTrendDTO alertTrend(Long userId, String termsParam) {
        List<Long> studentIds = collegeStudents(collegeMajors(userId).stream().map(Major::getId).toList())
                .stream().map(Student::getId).toList();
        List<String> terms = parseTerms(termsParam, studentIds);
        List<Alert> alerts = loadAlerts(studentIds, null);

        List<Integer> red = new ArrayList<>();
        List<Integer> orange = new ArrayList<>();
        List<Integer> yellow = new ArrayList<>();
        for (String term : terms) {
            List<Alert> termAlerts = alerts.stream()
                    .filter(a -> term.equals(termOf(alertDate(a))))
                    .toList();
            red.add(countLevel(termAlerts, "red"));
            orange.add(countLevel(termAlerts, "orange"));
            yellow.add(countLevel(termAlerts, "yellow"));
        }
        return new DeanAlertTrendDTO(terms, red, orange, yellow);
    }

    private List<DeanDashboardDTO.DepartmentRankingDTO> departmentRanking(
            List<Major> majors, List<Student> students, String resolvedTerm,
            List<Alert> alerts, Map<Long, GpaHistory> historyByStudent) {
        Map<Long, Major> majorById = majors.stream()
                .collect(Collectors.toMap(Major::getId, major -> major));
        Map<String, List<Student>> studentsByName = groupStudentsByMajorName(students, majorById);
        List<DeanDashboardDTO.DepartmentRankingDTO> ranking = new ArrayList<>();
        for (Map.Entry<String, List<Student>> entry : studentsByName.entrySet()) {
            List<Student> majorStudents = entry.getValue();
            List<Long> ids = majorStudents.stream().map(Student::getId).toList();
            List<Alert> majorAlerts = alerts.stream().filter(a -> ids.contains(a.getStudentId())).toList();
            int warningStudents = majorAlerts.size();
            List<Grade> grades = loadGrades(ids, resolvedTerm);
            int passed = (int) grades.stream().filter(g -> "passed".equals(g.getStatus())).count();
            BigDecimal avgGpa = average(majorStudents.stream()
                    .map(student -> historyByStudent.get(student.getId()))
                    .filter(Objects::nonNull)
                    .map(GpaHistory::getGpa)
                    .filter(Objects::nonNull)
                    .toList(), 2);
            BigDecimal passRate = percent(passed, grades.size(), 1);
            BigDecimal alertRate = percent(warningStudents, majorStudents.size(), 1);
            int healthScore = healthScore(avgGpa, passRate, alertRate);
            ranking.add(new DeanDashboardDTO.DepartmentRankingDTO(
                    entry.getKey(), healthScore, passRate, avgGpa));
        }
        ranking.sort(Comparator.comparing(DeanDashboardDTO.DepartmentRankingDTO::healthScore).reversed());
        return ranking;
    }

    private List<DeanDashboardDTO.AlertHeatmapDTO> alertHeatmap(
            List<Major> majors, List<Alert> alerts, Map<Long, Student> studentById) {
        Map<Long, Major> majorById = majors.stream()
                .collect(Collectors.toMap(Major::getId, major -> major));
        Map<String, List<Alert>> alertsByName = new LinkedHashMap<>();
        for (Alert alert : alerts) {
            Student student = studentById.get(alert.getStudentId());
            Major major = student == null ? null : majorById.get(student.getMajorId());
            String name = major == null ? "未知专业" : major.getName();
            alertsByName.computeIfAbsent(name, k -> new ArrayList<>()).add(alert);
        }
        List<DeanDashboardDTO.AlertHeatmapDTO> result = new ArrayList<>();
        alertsByName.forEach((name, majorAlerts) -> {
            result.add(new DeanDashboardDTO.AlertHeatmapDTO(
                    name,
                    countLevel(majorAlerts, "yellow"),
                    countLevel(majorAlerts, "orange"),
                    countLevel(majorAlerts, "red")));
        });
        return result;
    }

    private int healthScore(BigDecimal avgGpa, BigDecimal passRate, BigDecimal alertRate) {
        double gpaScore = avgGpa.doubleValue() / 4.0 * 40;
        double passScore = passRate.doubleValue() * 0.35;
        double alertScore = (100 - alertRate.doubleValue()) * 0.25;
        return (int) Math.round(gpaScore + passScore + alertScore);
    }

    private int countLevel(List<Alert> alerts, String level) {
        return (int) alerts.stream().filter(a -> level.equals(a.getLevel())).count();
    }

    private List<String> parseTerms(String termsParam, List<Long> studentIds) {
        if (termsParam != null && !termsParam.isBlank()) {
            return List.of(termsParam.split(",")).stream()
                    .map(String::trim)
                    .filter(s -> !s.isEmpty())
                    .toList();
        }
        if (studentIds.isEmpty()) {
            return List.of();
        }
        List<String> allTerms = gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                        .in(GpaHistory::getStudentId, studentIds))
                .stream()
                .map(GpaHistory::getTerm)
                .distinct()
                .sorted(Comparator.reverseOrder())
                .limit(4)
                .toList();
        List<String> ascending = new ArrayList<>(allTerms);
        ascending.sort(String::compareTo);
        return ascending;
    }

    private LocalDate alertDate(Alert alert) {
        if (alert.getTriggerDate() != null) {
            return alert.getTriggerDate();
        }
        return alert.getCreatedAt() == null ? LocalDate.now() : alert.getCreatedAt().toLocalDate();
    }

    private String termOf(LocalDate date) {
        return AcademicTerm.fromDate(date);
    }

    private String gradeNote(String grade) {
        return switch (grade) {
            case "2021" -> "毕业前重点跟踪";
            case "2022" -> "课业压力最重学年";
            case "2023" -> "整体相对稳定";
            case "2024" -> "新生适应期";
            default -> "";
        };
    }

    private List<Major> collegeMajors(Long userId) {
        Department department = departmentMapper.selectOne(new LambdaQueryWrapper<Department>()
                .eq(Department::getDeanId, userId));
        if (department == null) {
            throw new BusinessException(403, "当前账号未配置学院数据权限");
        }
        return majorMapper.selectList(new LambdaQueryWrapper<Major>()
                .eq(Major::getDepartmentId, department.getId())
                .orderByAsc(Major::getId));
    }

    private Map<String, List<Student>> groupStudentsByMajorName(
            List<Student> students, Map<Long, Major> majorById) {
        Map<String, List<Student>> byName = new LinkedHashMap<>();
        for (Student student : students) {
            Major major = majorById.get(student.getMajorId());
            String name = major == null ? "未知专业" : major.getName();
            byName.computeIfAbsent(name, k -> new ArrayList<>()).add(student);
        }
        return byName;
    }

    private List<Student> collegeStudents(List<Long> majorIds) {
        if (majorIds.isEmpty()) {
            return List.of();
        }
        return studentMapper.selectList(new LambdaQueryWrapper<Student>()
                .in(Student::getMajorId, majorIds));
    }

    private List<Alert> loadAlerts(List<Long> studentIds, String term) {
        if (studentIds.isEmpty()) {
            return List.of();
        }
        LambdaQueryWrapper<Alert> wrapper = new LambdaQueryWrapper<Alert>()
                .in(Alert::getStudentId, studentIds);
        AcademicTerm.range(term).ifPresent(range ->
                wrapper.between(Alert::getTriggerDate, range.start(), range.end()));
        return alertMapper.selectList(wrapper);
    }

    private List<GpaHistory> loadHistories(List<Long> studentIds, String term) {
        if (studentIds.isEmpty() || term == null || term.isBlank()) {
            return List.of();
        }
        return gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .in(GpaHistory::getStudentId, studentIds)
                .eq(GpaHistory::getTerm, term));
    }

    private List<Grade> loadGrades(List<Long> studentIds, String term) {
        if (studentIds.isEmpty() || term == null || term.isBlank()) {
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
}
