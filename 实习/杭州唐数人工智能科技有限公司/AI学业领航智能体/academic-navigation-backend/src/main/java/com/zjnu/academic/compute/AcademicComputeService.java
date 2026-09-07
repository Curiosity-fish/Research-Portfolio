package com.zjnu.academic.compute;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zjnu.academic.alert.entity.Alert;
import com.zjnu.academic.alert.mapper.AlertMapper;
import com.zjnu.academic.course.entity.Course;
import com.zjnu.academic.course.entity.CourseClass;
import com.zjnu.academic.course.entity.Grade;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.health.entity.HealthReport;
import com.zjnu.academic.health.mapper.HealthReportMapper;
import com.zjnu.academic.rawdata.entity.Attendance;
import com.zjnu.academic.rawdata.entity.Competition;
import com.zjnu.academic.rawdata.entity.Psychology;
import com.zjnu.academic.rawdata.entity.Volunteer;
import com.zjnu.academic.rawdata.mapper.AttendanceMapper;
import com.zjnu.academic.rawdata.mapper.CompetitionMapper;
import com.zjnu.academic.rawdata.mapper.PsychologyMapper;
import com.zjnu.academic.rawdata.mapper.VolunteerMapper;
import com.zjnu.academic.student.entity.GpaHistory;
import com.zjnu.academic.student.entity.ProfileScore;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.ProfileScoreMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;

/**
 * 派生数据计算服务：从校方原始数据（成绩、体测、竞赛、志愿、考勤、心理测评）
 * 计算 GPA 历史、五维画像和预警，结果落库供各端接口直接读取。
 */
@Service
public class AcademicComputeService {

    private static final DateTimeFormatter DATETIME = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm");

    private final StudentMapper studentMapper;
    private final GradeMapper gradeMapper;
    private final CourseClassMapper courseClassMapper;
    private final CourseMapper courseMapper;
    private final GpaHistoryMapper gpaHistoryMapper;
    private final ProfileScoreMapper profileScoreMapper;
    private final HealthReportMapper healthReportMapper;
    private final AlertMapper alertMapper;
    private final CompetitionMapper competitionMapper;
    private final VolunteerMapper volunteerMapper;
    private final AttendanceMapper attendanceMapper;
    private final PsychologyMapper psychologyMapper;

    @Autowired
    public AcademicComputeService(StudentMapper studentMapper,
                                  GradeMapper gradeMapper,
                                  CourseClassMapper courseClassMapper,
                                  CourseMapper courseMapper,
                                  GpaHistoryMapper gpaHistoryMapper,
                                  ProfileScoreMapper profileScoreMapper,
                                  HealthReportMapper healthReportMapper,
                                  AlertMapper alertMapper,
                                  CompetitionMapper competitionMapper,
                                  VolunteerMapper volunteerMapper,
                                  AttendanceMapper attendanceMapper,
                                  PsychologyMapper psychologyMapper) {
        this.studentMapper = studentMapper;
        this.gradeMapper = gradeMapper;
        this.courseClassMapper = courseClassMapper;
        this.courseMapper = courseMapper;
        this.gpaHistoryMapper = gpaHistoryMapper;
        this.profileScoreMapper = profileScoreMapper;
        this.healthReportMapper = healthReportMapper;
        this.alertMapper = alertMapper;
        this.competitionMapper = competitionMapper;
        this.volunteerMapper = volunteerMapper;
        this.attendanceMapper = attendanceMapper;
        this.psychologyMapper = psychologyMapper;
    }

    public Map<String, Integer> recomputeAll() {
        Map<String, Integer> counts = new LinkedHashMap<>();
        counts.put("gpaHistory", recomputeGpaHistory());
        counts.put("studentRanks", syncStudentRank());
        counts.put("profileScores", recomputeProfileScores());
        counts.put("alerts", recomputeAlerts());
        return counts;
    }

    private int syncStudentRank() {
        List<GpaHistory> histories = gpaHistoryMapper.selectList(null);
        Map<Long, GpaHistory> latest = new HashMap<>();
        for (GpaHistory history : histories) {
            GpaHistory previous = latest.get(history.getStudentId());
            if (previous == null || history.getTerm().compareTo(previous.getTerm()) > 0) {
                latest.put(history.getStudentId(), history);
            }
        }
        int count = 0;
        for (Student student : studentMapper.selectList(null)) {
            GpaHistory history = latest.get(student.getId());
            if (history == null || history.getRank() == null || history.getTotalStudents() == null) {
                continue;
            }
            student.setRank(history.getRank());
            student.setTotalStudents(history.getTotalStudents());
            student.setGpa(history.getGpa());
            studentMapper.updateById(student);
            count++;
        }
        return count;
    }

    public int recomputeGpaHistory() {
        List<Grade> grades = gradeMapper.selectList(null);
        if (grades.isEmpty()) {
            return 0;
        }
        Map<Long, Long> classCourse = courseClassMapper.selectList(null).stream()
                .collect(Collectors.toMap(CourseClass::getId, CourseClass::getCourseId, (a, b) -> a));
        Map<Long, BigDecimal> courseCredits = courseMapper.selectBatchIds(
                        classCourse.values().stream().distinct().toList()).stream()
                .collect(Collectors.toMap(Course::getId, Course::getCredits, (a, b) -> a));

        Map<String, List<Grade>> byStudentTerm = new HashMap<>();
        for (Grade grade : grades) {
            String key = grade.getStudentId() + "|" + grade.getTerm();
            byStudentTerm.computeIfAbsent(key, k -> new ArrayList<>()).add(grade);
        }
        Map<String, BigDecimal> termGpa = new HashMap<>();
        for (Map.Entry<String, List<Grade>> entry : byStudentTerm.entrySet()) {
            BigDecimal denominator = BigDecimal.ZERO;
            BigDecimal numerator = BigDecimal.ZERO;
            for (Grade grade : entry.getValue()) {
                if (grade.getScore() == null) {
                    continue;
                }
                BigDecimal credit = courseCredits.getOrDefault(
                        classCourse.getOrDefault(grade.getCourseClassId(), -1L), BigDecimal.ONE);
                denominator = denominator.add(credit);
                BigDecimal gradePoint = grade.getGradePoint() == null
                        ? BigDecimal.ZERO : grade.getGradePoint();
                numerator = numerator.add(gradePoint.multiply(credit));
            }
            if (denominator.signum() > 0) {
                termGpa.put(entry.getKey(),
                        numerator.divide(denominator, 2, RoundingMode.HALF_UP));
            }
        }

        Map<Long, Student> students = studentMapper.selectList(null).stream()
                .collect(Collectors.toMap(Student::getId, s -> s));
        Map<String, List<TermGpa>> byMajorTerm = new HashMap<>();
        for (Map.Entry<String, BigDecimal> entry : termGpa.entrySet()) {
            String[] parts = entry.getKey().split("\\|");
            Long studentId = Long.valueOf(parts[0]);
            String term = parts[1];
            Student student = students.get(studentId);
            String groupKey = (student == null ? -1L : student.getMajorId()) + "|" + term;
            byMajorTerm.computeIfAbsent(groupKey, k -> new ArrayList<>())
                    .add(new TermGpa(studentId, term, entry.getValue()));
        }

        Map<String, RankInfo> ranks = new HashMap<>();
        for (Map.Entry<String, List<TermGpa>> entry : byMajorTerm.entrySet()) {
            List<TermGpa> list = entry.getValue();
            list.sort(Comparator.comparing(TermGpa::gpa).reversed());
            double avg = list.stream().mapToDouble(t -> t.gpa().doubleValue()).average().orElse(0);
            for (int i = 0; i < list.size(); i++) {
                ranks.put(list.get(i).studentId() + "|" + list.get(i).term(),
                        new RankInfo(i + 1, list.size(),
                                BigDecimal.valueOf(avg).setScale(2, RoundingMode.HALF_UP)));
            }
        }

        int count = 0;
        for (Map.Entry<String, BigDecimal> entry : termGpa.entrySet()) {
            String[] parts = entry.getKey().split("\\|");
            Long studentId = Long.valueOf(parts[0]);
            String term = parts[1];
            RankInfo rank = ranks.get(entry.getKey());
            GpaHistory history = gpaHistoryMapper.selectOne(new LambdaQueryWrapper<GpaHistory>()
                    .eq(GpaHistory::getStudentId, studentId)
                    .eq(GpaHistory::getTerm, term));
            if (history == null) {
                history = new GpaHistory();
                history.setStudentId(studentId);
                history.setTerm(term);
                history.setGpa(entry.getValue());
                history.setAvgGpa(rank == null ? null : rank.avgGpa());
                history.setRank(rank == null ? null : rank.rank());
                history.setTotalStudents(rank == null ? null : rank.total());
                gpaHistoryMapper.insert(history);
            } else {
                history.setGpa(entry.getValue());
                history.setAvgGpa(rank == null ? null : rank.avgGpa());
                history.setRank(rank == null ? null : rank.rank());
                history.setTotalStudents(rank == null ? null : rank.total());
                gpaHistoryMapper.updateById(history);
            }
            count++;
        }
        return count;
    }

    public int recomputeProfileScores() {
        List<GpaHistory> histories = gpaHistoryMapper.selectList(null);
        Map<String, List<Grade>> gradesByTerm = gradeMapper.selectList(null).stream()
                .collect(Collectors.groupingBy(g -> profileKey(g.getStudentId(), g.getTerm())));
        Map<String, List<HealthReport>> healthByTerm = healthReportMapper.selectList(null).stream()
                .collect(Collectors.groupingBy(r -> profileKey(r.getStudentId(), r.getTerm())));
        Map<String, List<Competition>> competitionByTerm = competitionMapper.selectList(null).stream()
                .collect(Collectors.groupingBy(c -> profileKey(c.getStudentId(), c.getTerm())));
        Map<String, List<Volunteer>> volunteerByTerm = volunteerMapper.selectList(null).stream()
                .collect(Collectors.groupingBy(v -> profileKey(v.getStudentId(), v.getTerm())));
        Map<String, List<Psychology>> psychologyByTerm = psychologyMapper.selectList(null).stream()
                .collect(Collectors.groupingBy(p -> profileKey(p.getStudentId(), p.getTerm())));
        Map<String, List<Attendance>> attendanceByTerm = attendanceMapper.selectList(null).stream()
                .collect(Collectors.groupingBy(a -> profileKey(a.getStudentId(), a.getTerm())));

        int count = 0;
        for (GpaHistory history : histories) {
            Long studentId = history.getStudentId();
            String term = history.getTerm();
            String key = profileKey(studentId, term);
            BigDecimal gpa = history.getGpa();
            int rank = history.getRank() == null ? 0 : history.getRank();
            int total = history.getTotalStudents() == null ? 0 : history.getTotalStudents();
            List<Grade> termGrades = gradesByTerm.getOrDefault(key, List.of());
            int failedCount = countFailedGrades(termGrades);
            int avgScore = avgScore(termGrades);
            int healthScore = healthScore(healthByTerm.get(key));
            int competitionScore = competitionScore(competitionByTerm.get(key));
            int volunteerHours = volunteerHours(volunteerByTerm.get(key));
            int psychologyScore = psychologyScore(psychologyByTerm.get(key));
            int attendanceRate = attendanceRate(attendanceByTerm.get(key));

            upsertProfile(studentId, term, "academic", "学业成绩",
                    academicScore(gpa, rank, total, avgScore, failedCount), 72, 100,
                    "基于当学期GPA、排名、挂科数与课程均分计算",
                    List.of("GPA: " + fmt(gpa), "专业排名: " + rank + " / " + total,
                            "挂科门数: " + failedCount, "课程均分: " + avgScore));
            upsertProfile(studentId, term, "practice", "实践能力",
                    competitionScore, 68, 100, "基于竞赛经历与实践积分计算",
                    List.of("竞赛积分: " + competitionScore, "竞赛记录: " + countCompetitions(competitionByTerm.get(key)) + " 项"));
            upsertProfile(studentId, term, "culture", "人文素养",
                    psychologyScore, 75, 100, "基于心理测评结果计算",
                    List.of("心理测评: " + psychologyScore + " 分"));
            upsertProfile(studentId, term, "health", "身心健康",
                    healthScore, 80, 100, "基于体测总分计算",
                    List.of("体测总分: " + healthScore));
            upsertProfile(studentId, term, "quality", "综合素质",
                    thoughtScore(volunteerHours), 85, 100, "基于志愿服务时长计算",
                    List.of("志愿时长: " + volunteerHours + " 小时", "出勤率: " + attendanceRate + "%"));
            count += 5;
        }
        return count;
    }

    public int recomputeAlerts() {
        List<Student> students = studentMapper.selectList(null);
        List<Grade> grades = gradeMapper.selectList(null);
        Map<Long, Long> classCourse = courseClassMapper.selectList(null).stream()
                .collect(Collectors.toMap(CourseClass::getId, CourseClass::getCourseId, (a, b) -> a));
        Map<Long, String> courseNames = courseMapper.selectBatchIds(
                        classCourse.values().stream().distinct().toList()).stream()
                .collect(Collectors.toMap(Course::getId, Course::getName, (a, b) -> a));
        Map<Long, List<Grade>> gradesByStudent = grades.stream()
                .collect(Collectors.groupingBy(Grade::getStudentId));
        Map<Long, String> latestTermByStudent = gpaHistoryMapper.selectList(null).stream()
                .collect(Collectors.toMap(GpaHistory::getStudentId, GpaHistory::getTerm,
                        (a, b) -> a.compareTo(b) >= 0 ? a : b));
        Map<Long, List<Attendance>> attendanceByStudent = attendanceMapper.selectList(null).stream()
                .collect(Collectors.groupingBy(Attendance::getStudentId));
        int count = 0;
        for (Student student : students) {
            // 每个学生只保留 1 条未闭环预警，保证预警人数不超过学生人数
            alertMapper.delete(new LambdaQueryWrapper<Alert>()
                    .eq(Alert::getStudentId, student.getId())
                    .in(Alert::getStatus, List.of("pending", "processing")));
            List<Grade> studentGrades = gradesByStudent.getOrDefault(student.getId(), List.of());
            String latestTerm = latestTermByStudent.getOrDefault(student.getId(), "2024-2025-1");
            int failed = (int) studentGrades.stream()
                    .filter(g -> latestTerm.equals(g.getTerm()))
                    .filter(g -> "failed".equals(g.getStatus())
                            || (g.getScore() != null && g.getScore().doubleValue() < 60))
                    .count();
            int absences = attendanceByStudent.getOrDefault(student.getId(), List.of()).stream()
                    .filter(a -> latestTerm.equals(a.getTerm()))
                    .mapToInt(a -> a.getAbsentCount() == null ? 0 : a.getAbsentCount())
                    .sum();
            AcademicAlertPolicy.Assessment assessment = AcademicAlertPolicy.assess(
                    failed, absences);
            String level = assessment.level();
            if ("none".equals(level)) {
                if (!"none".equals(student.getAlertLevel())) {
                    student.setAlertLevel("none");
                    studentMapper.updateById(student);
                }
                continue;
            }
            List<String> failedNames = studentGrades.stream()
                    .filter(g -> latestTerm.equals(g.getTerm()))
                    .filter(g -> "failed".equals(g.getStatus())
                            || (g.getScore() != null && g.getScore().doubleValue() < 60))
                    .map(g -> courseNames.getOrDefault(classCourse.get(g.getCourseClassId()), "未知课程"))
                    .distinct()
                    .limit(6)
                    .toList();
            String type = failed > 0 && absences > 0 ? "复合风险预警" : assessment.primaryRisk();
            String triggerEvent = "挂科 " + failed + " 门；缺勤 " + absences + " 次";
            if (!"none".equals(level)) {
                Alert alert = new Alert();
                alert.setStudentId(student.getId());
                alert.setLevel(level);
                alert.setType(type);
                alert.setTitle(alertTitle(level, type));
                alert.setDescription("系统根据挂科数与缺勤次数判定：当学期不及格 " + failed
                        + " 门，缺勤 " + absences + " 次");
                alert.setFailedCourses(failedNames);
                LocalDate triggerDate = termExamDate(latestTerm);
                alert.setTriggerDate(triggerDate);
                alert.setStatus("pending");
                alert.setSuggestion(alertSuggestion(level, failed));
                alert.setTriggerEvent(triggerEvent);
                alert.setPushedAt(triggerDate.atTime(14, 0));
                alertMapper.insert(alert);
                count++;
            }
            if (!level.equals(student.getAlertLevel())) {
                student.setAlertLevel(level);
                studentMapper.updateById(student);
            }
        }
        return count;
    }

    private LocalDate termExamDate(String term) {
        try {
            String[] parts = term.split("-");
            int endYear = Integer.parseInt(parts[1]);
            return "2".equals(parts[2])
                    ? LocalDate.of(endYear, 7, 10)
                    : LocalDate.of(endYear, 1, 15);
        } catch (Exception e) {
            return LocalDate.now();
        }
    }

    private void upsertProfile(Long studentId, String term, String key, String label,
                               int score, int avgScore, int maxScore,
                               String description, List<String> details) {
        ProfileScore existing = profileScoreMapper.selectOne(new LambdaQueryWrapper<ProfileScore>()
                .eq(ProfileScore::getStudentId, studentId)
                .eq(ProfileScore::getTerm, term)
                .eq(ProfileScore::getDimensionKey, key));
        int safeScore = clamp(score, 0, 100);
        if (existing == null) {
            ProfileScore scoreEntity = new ProfileScore();
            scoreEntity.setStudentId(studentId);
            scoreEntity.setTerm(term);
            scoreEntity.setDimensionKey(key);
            scoreEntity.setLabel(label);
            scoreEntity.setScore(safeScore);
            scoreEntity.setAvgScore(avgScore);
            scoreEntity.setMaxScore(maxScore);
            scoreEntity.setDescription(description);
            scoreEntity.setDetails(details);
            profileScoreMapper.insert(scoreEntity);
        } else {
            existing.setLabel(label);
            existing.setScore(safeScore);
            existing.setAvgScore(avgScore);
            existing.setMaxScore(maxScore);
            existing.setDescription(description);
            existing.setDetails(details);
            profileScoreMapper.updateById(existing);
        }
    }

    private int countFailedGrades(List<Grade> grades) {
        return (int) grades.stream()
                .filter(g -> "failed".equals(g.getStatus())
                        || (g.getScore() != null && g.getScore().doubleValue() < 60))
                .count();
    }

    private int avgScore(List<Grade> grades) {
        return (int) Math.round(grades.stream()
                .map(Grade::getScore)
                .filter(java.util.Objects::nonNull)
                .mapToDouble(BigDecimal::doubleValue)
                .average()
                .orElse(0));
    }

    private int healthScore(List<HealthReport> reports) {
        if (reports == null || reports.isEmpty()) {
            return 75;
        }
        return reports.stream()
                .filter(r -> r.getTotalScore() != null)
                .findFirst()
                .map(HealthReport::getTotalScore)
                .orElse(75);
    }

    private String profileKey(Long studentId, String term) {
        return studentId + "|" + term;
    }

    private int competitionScore(List<Competition> competitions) {
        if (competitions == null || competitions.isEmpty()) {
            return 55;
        }
        int maxPoints = competitions.stream()
                .map(c -> c.getPoints() == null ? 0 : c.getPoints())
                .max(Integer::compareTo)
                .orElse(0);
        if (maxPoints >= 90) return 90;
        if (maxPoints >= 80) return 82;
        if (maxPoints >= 70) return 74;
        return 66;
    }

    private int countCompetitions(List<Competition> competitions) {
        return competitions == null ? 0 : competitions.size();
    }

    private int volunteerHours(List<Volunteer> volunteers) {
        if (volunteers == null || volunteers.isEmpty()) {
            return 0;
        }
        return volunteers.stream()
                .map(v -> v.getHours() == null ? BigDecimal.ZERO : v.getHours())
                .mapToInt(BigDecimal::intValue)
                .sum();
    }

    private int psychologyScore(List<Psychology> psychologies) {
        if (psychologies == null || psychologies.isEmpty()) {
            return 72;
        }
        int score = psychologies.stream()
                .filter(p -> p.getScore() != null)
                .map(Psychology::getScore)
                .max(Integer::compareTo)
                .orElse(72);
        if (score >= 85) return 90;
        if (score >= 70) return 78;
        return 68;
    }

    private int attendanceRate(List<Attendance> attendances) {
        if (attendances == null || attendances.isEmpty()) {
            return 100;
        }
        int total = attendances.stream().mapToInt(a -> a.getTotalClasses() == null ? 0 : a.getTotalClasses()).sum();
        int absent = attendances.stream().mapToInt(a -> a.getAbsentCount() == null ? 0 : a.getAbsentCount()).sum();
        if (total <= 0) {
            return 100;
        }
        return Math.max(0, 100 - Math.round(absent * 100f / total));
    }

    private int academicScore(BigDecimal gpa, int rank, int total, int avgScore, int failedCount) {
        double gpaValue = gpa == null ? 0 : gpa.doubleValue();
        double rankPct = total <= 0 ? 0.5 : Math.min(1.0, rank * 1.0 / total);
        double score = 35 + gpaValue / 4.0 * 40 + (1 - rankPct) * 12 + avgScore / 100.0 * 8 - failedCount * 5;
        return clamp((int) Math.round(score), 0, 100);
    }

    private int thoughtScore(int volunteerHours) {
        if (volunteerHours >= 40) return 92;
        if (volunteerHours >= 20) return 85;
        if (volunteerHours >= 8) return 78;
        if (volunteerHours > 0) return 70;
        return 62;
    }

    private String alertTitle(String level, String type) {
        return switch (level) {
            case "red" -> "红色预警：" + type;
            case "orange" -> "橙色预警：" + type;
            default -> "黄色预警：" + type;
        };
    }

    private String alertSuggestion(String level, int failed) {
        if (failed == 0) {
            return "建议了解缺勤原因，关注学生状态并约定改进目标";
        }
        return switch (level) {
            case "red" -> "建议立即约谈并制定学业恢复计划";
            case "orange" -> "建议安排课后辅导并定期复盘";
            default -> "建议加强学习规划并跟踪成绩变化";
        };
    }

    private int clamp(int value, int min, int max) {
        return Math.max(min, Math.min(max, value));
    }

    private String fmt(BigDecimal value) {
        return value == null ? "0.00" : value.setScale(2, RoundingMode.HALF_UP).toString();
    }

    private record TermGpa(Long studentId, String term, BigDecimal gpa) {
    }

    private record RankInfo(int rank, int total, BigDecimal avgGpa) {
    }
}
