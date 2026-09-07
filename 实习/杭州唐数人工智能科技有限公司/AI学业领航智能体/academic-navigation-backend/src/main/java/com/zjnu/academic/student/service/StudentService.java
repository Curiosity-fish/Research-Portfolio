package com.zjnu.academic.student.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.alert.service.AlertService;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.course.entity.Course;
import com.zjnu.academic.course.entity.CourseClass;
import com.zjnu.academic.course.entity.Grade;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.health.entity.HealthReport;
import com.zjnu.academic.health.mapper.HealthReportMapper;
import com.zjnu.academic.student.dto.AcademicProfileDTO;
import com.zjnu.academic.student.dto.GpaTrendItemDTO;
import com.zjnu.academic.student.dto.HealthReportDTO;
import com.zjnu.academic.student.dto.ProfileDimensionDTO;
import com.zjnu.academic.student.dto.RankTrendItemDTO;
import com.zjnu.academic.student.dto.StudentDTO;
import com.zjnu.academic.student.dto.StudentDashboardDTO;
import com.zjnu.academic.student.dto.StudentCourseGradeDTO;
import com.zjnu.academic.student.dto.StudentTermGradesDTO;
import com.zjnu.academic.student.entity.GpaHistory;
import com.zjnu.academic.student.entity.ProfileScore;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.ProfileScoreMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Comparator;
import java.util.Map;
import java.util.function.Function;
import java.util.stream.Collectors;
import java.math.BigDecimal;
import java.math.RoundingMode;

@Service
public class StudentService {

    private final StudentMapper studentMapper;
    private final GpaHistoryMapper gpaHistoryMapper;
    private final ProfileScoreMapper profileScoreMapper;
    private final HealthReportMapper healthReportMapper;
    private final UserMapper userMapper;
    private final AlertService alertService;
    private final GradeMapper gradeMapper;
    private final CourseClassMapper courseClassMapper;
    private final CourseMapper courseMapper;

    @Autowired
    public StudentService(StudentMapper studentMapper,
                          GpaHistoryMapper gpaHistoryMapper,
                          ProfileScoreMapper profileScoreMapper,
                          HealthReportMapper healthReportMapper,
                          UserMapper userMapper,
                          AlertService alertService,
                          GradeMapper gradeMapper,
                          CourseClassMapper courseClassMapper,
                          CourseMapper courseMapper) {
        this.studentMapper = studentMapper;
        this.gpaHistoryMapper = gpaHistoryMapper;
        this.profileScoreMapper = profileScoreMapper;
        this.healthReportMapper = healthReportMapper;
        this.userMapper = userMapper;
        this.alertService = alertService;
        this.gradeMapper = gradeMapper;
        this.courseClassMapper = courseClassMapper;
        this.courseMapper = courseMapper;
    }

    public StudentDashboardDTO dashboard(Long userId, String term) {
        Student student = requireStudent(userId);
        User user = userMapper.selectById(student.getUserId());
        String grade = user == null ? null : user.getGrade();
        String resolvedTerm = resolveTerm(student.getId(), term);
        StudentTermGradesDTO termGrades = buildTermGrades(student, resolvedTerm);
        GpaHistory history = findHistory(student.getId(), resolvedTerm);

        List<GpaTrendItemDTO> trend = buildGpaTrend(student.getId(), grade, resolvedTerm);
        List<AlertItemDTO> alerts = alertService.list(
                userId, "student", resolvedTerm, null, null, null, 1, 1000).list();
        int alertCount = alerts.size();
        HealthReport health = latestHealth(student.getId(), resolvedTerm);
        ClassRank classRank = classRank(student, resolvedTerm);

        return new StudentDashboardDTO(
                history == null ? termGrades.gpa() : history.getGpa(),
                history == null ? student.getRank() : history.getRank(),
                history == null ? student.getTotalStudents() : history.getTotalStudents(),
                classRank.rank(),
                classRank.total(),
                termGrades.earnedCredits(),
                student.getRequiredCredits(),
                termGrades.courseCount(),
                alertCount,
                health == null ? 0 : health.getTotalScore(),
                trend,
                alerts.stream().limit(3).toList());
    }

    private ClassRank classRank(Student student, String term) {
        if (student.getClassId() == null) {
            return new ClassRank(0, 0);
        }
        List<Student> classmates = studentMapper.selectList(new LambdaQueryWrapper<Student>()
                .eq(Student::getClassId, student.getClassId()));
        List<Long> classmateIds = classmates.stream().map(Student::getId).toList();
        if (classmateIds.isEmpty()) {
            return new ClassRank(0, 0);
        }
        List<GpaHistory> histories = gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .in(GpaHistory::getStudentId, classmateIds)
                .eq(GpaHistory::getTerm, term));
        List<GpaHistory> sorted = histories.stream()
                .sorted(Comparator.comparing(
                        (GpaHistory h) -> h.getGpa() == null ? BigDecimal.ZERO : h.getGpa()).reversed())
                .toList();
        for (int i = 0; i < sorted.size(); i++) {
            if (sorted.get(i).getStudentId().equals(student.getId())) {
                return new ClassRank(i + 1, sorted.size());
            }
        }
        return new ClassRank(0, sorted.size());
    }

    private record ClassRank(int rank, int total) {
    }

    public AcademicProfileDTO profile(Long userId, String term) {
        Student student = requireStudent(userId);
        User user = userMapper.selectById(student.getUserId());
        String grade = user == null ? null : user.getGrade();
        String resolvedTerm = resolveTerm(student.getId(), term);

        List<ProfileScore> scores = profileScoreMapper.selectList(new LambdaQueryWrapper<ProfileScore>()
                .eq(ProfileScore::getStudentId, student.getId())
                .eq(ProfileScore::getTerm, resolvedTerm))
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

        List<GpaHistory> histories = loadHistories(student.getId()).stream()
                .filter(history -> resolvedTerm.isBlank() || history.getTerm().compareTo(resolvedTerm) <= 0)
                .toList();
        return new AcademicProfileDTO(
                dimensions,
                toGpaTrend(histories, grade),
                toRankTrend(histories, grade));
    }

    public List<GpaTrendItemDTO> gpaTrend(Long userId, String term) {
        Student student = requireStudent(userId);
        User user = userMapper.selectById(student.getUserId());
        String resolvedTerm = resolveTerm(student.getId(), term);
        return buildGpaTrend(student.getId(), user == null ? null : user.getGrade(), resolvedTerm);
    }

    public List<String> availableTerms(Long userId) {
        Student student = requireStudent(userId);
        return availableTermsForStudent(student.getId());
    }

    public StudentTermGradesDTO termGrades(Long userId, String term) {
        Student student = requireStudent(userId);
        return buildTermGrades(student, resolveTerm(student.getId(), term));
    }

    public PageResult<AlertItemDTO> alerts(Long userId, String term, int page, int size) {
        return alertService.list(userId, "student", term, null, null, null, page, size);
    }

    public List<HealthReportDTO> healthReports(Long userId, String term) {
        Student student = requireStudent(userId);
        LambdaQueryWrapper<HealthReport> query = new LambdaQueryWrapper<HealthReport>()
                .eq(HealthReport::getStudentId, student.getId());
        if (term != null && !term.isBlank()) {
            query.eq(HealthReport::getTerm, term);
        }
        List<HealthReport> reports = healthReportMapper.selectList(query.orderByDesc(HealthReport::getTerm));
        return reports.stream().map(report -> new HealthReportDTO(
                String.valueOf(report.getId()),
                student.getStudentId(),
                report.getTerm(),
                report.getTotalScore(),
                report.getItems() == null ? List.of() : report.getItems(),
                report.getReportUrl()
        )).toList();
    }

    public HealthReportDTO healthReportDetail(Long userId, Long id) {
        Student student = requireStudent(userId);
        HealthReport report = healthReportMapper.selectById(id);
        if (report == null || !report.getStudentId().equals(student.getId())) {
            throw new BusinessException(404, "体测报告不存在");
        }
        return new HealthReportDTO(
                String.valueOf(report.getId()),
                student.getStudentId(),
                report.getTerm(),
                report.getTotalScore(),
                report.getItems() == null ? List.of() : report.getItems(),
                report.getReportUrl());
    }

    public StudentDTO toStudentDTO(Student student) {
        User user = userMapper.selectById(student.getUserId());
        return new StudentDTO(
                String.valueOf(student.getId()),
                user == null ? "" : user.getName(),
                student.getStudentId(),
                user == null ? "" : user.getCollege(),
                user == null ? "" : user.getMajor(),
                user == null ? "" : user.getClassName(),
                user == null ? "" : user.getGrade(),
                student.getGpa(),
                student.getRank(),
                student.getTotalStudents(),
                student.getAlertLevel(),
                null);
    }

    public Student requireStudent(Long userId) {
        Student student = studentMapper.selectOne(
                new LambdaQueryWrapper<Student>().eq(Student::getUserId, userId));
        if (student == null) {
            throw new BusinessException(404, "学生档案不存在");
        }
        return student;
    }

    private List<GpaTrendItemDTO> buildGpaTrend(Long studentId, String grade, String throughTerm) {
        List<GpaHistory> histories = loadHistories(studentId);
        if (throughTerm != null && !throughTerm.isBlank()) {
            histories = histories.stream()
                    .filter(history -> history.getTerm().compareTo(throughTerm) <= 0)
                    .toList();
        }
        return toGpaTrend(histories, grade);
    }

    private List<GpaHistory> loadHistories(Long studentId) {
        return gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .eq(GpaHistory::getStudentId, studentId)
                .orderByAsc(GpaHistory::getTerm));
    }

    private List<GpaTrendItemDTO> toGpaTrend(List<GpaHistory> histories, String grade) {
        return histories.stream()
                .map(h -> new GpaTrendItemDTO(termLabel(h.getTerm(), grade), h.getGpa(), h.getAvgGpa()))
                .toList();
    }

    private List<RankTrendItemDTO> toRankTrend(List<GpaHistory> histories, String grade) {
        return histories.stream()
                .map(h -> new RankTrendItemDTO(termLabel(h.getTerm(), grade), h.getRank(), h.getTotalStudents()))
                .toList();
    }

    private String resolveTerm(Long studentId, String term) {
        if (term != null && !term.isBlank()) {
            return term;
        }
        List<String> terms = availableTermsForStudent(studentId);
        if (!terms.isEmpty()) {
            return terms.get(0);
        }
        return loadHistories(studentId).stream().map(GpaHistory::getTerm)
                .max(String::compareTo).orElse("");
    }

    private HealthReport latestHealth(Long studentId, String term) {
        LambdaQueryWrapper<HealthReport> query = new LambdaQueryWrapper<HealthReport>()
                .eq(HealthReport::getStudentId, studentId);
        if (term != null && !term.isBlank()) {
            query.eq(HealthReport::getTerm, term);
        }
        List<HealthReport> reports = healthReportMapper.selectList(query.orderByDesc(HealthReport::getTerm));
        if (reports.isEmpty()) {
            return null;
        }
        return reports.get(0);
    }

    private List<String> availableTermsForStudent(Long studentId) {
        List<String> gradeTerms = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                        .eq(Grade::getStudentId, studentId))
                .stream()
                .map(Grade::getTerm)
                .toList();
        List<String> historyTerms = loadHistories(studentId).stream()
                .map(GpaHistory::getTerm)
                .toList();
        return java.util.stream.Stream.concat(gradeTerms.stream(), historyTerms.stream())
                .filter(value -> value != null && !value.isBlank())
                .distinct()
                .sorted(Comparator.reverseOrder())
                .toList();
    }

    private GpaHistory findHistory(Long studentId, String term) {
        return gpaHistoryMapper.selectOne(new LambdaQueryWrapper<GpaHistory>()
                .eq(GpaHistory::getStudentId, studentId)
                .eq(GpaHistory::getTerm, term));
    }

    private StudentTermGradesDTO buildTermGrades(Student student, String term) {
        List<Grade> grades = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                .eq(Grade::getStudentId, student.getId())
                .eq(Grade::getTerm, term)
                .orderByAsc(Grade::getCourseClassId));

        Map<Long, CourseClass> courseClasses = grades.isEmpty() ? Map.of()
                : courseClassMapper.selectBatchIds(grades.stream().map(Grade::getCourseClassId).distinct().toList())
                .stream().collect(Collectors.toMap(CourseClass::getId, Function.identity()));
        List<Long> courseIds = courseClasses.values().stream().map(CourseClass::getCourseId).distinct().toList();
        Map<Long, Course> courses = courseIds.isEmpty() ? Map.of()
                : courseMapper.selectBatchIds(courseIds).stream()
                .collect(Collectors.toMap(Course::getId, Function.identity()));

        List<StudentCourseGradeDTO> courseGrades = grades.stream().map(grade -> {
            CourseClass courseClass = courseClasses.get(grade.getCourseClassId());
            Course course = courseClass == null ? null : courses.get(courseClass.getCourseId());
            return new StudentCourseGradeDTO(
                    course == null ? "" : course.getCode(),
                    course == null ? "未知课程" : course.getName(),
                    course == null ? BigDecimal.ZERO : course.getCredits(),
                    grade.getScore(),
                    grade.getGradePoint(),
                    grade.getStatus());
        }).sorted(Comparator.comparing(StudentCourseGradeDTO::courseCode)).toList();

        BigDecimal earnedCredits = courseGrades.stream()
                .filter(this::isPassed)
                .map(StudentCourseGradeDTO::credits)
                .reduce(BigDecimal.ZERO, BigDecimal::add);
        int passedCount = (int) courseGrades.stream().filter(this::isPassed).count();
        int failedCount = (int) courseGrades.stream().filter(this::isFailed).count();
        GpaHistory history = findHistory(student.getId(), term);
        BigDecimal gpa = history == null ? calculateGpa(courseGrades) : history.getGpa();

        return new StudentTermGradesDTO(
                term,
                gpa,
                history == null ? null : history.getAvgGpa(),
                history == null ? null : history.getRank(),
                history == null ? null : history.getTotalStudents(),
                earnedCredits,
                courseGrades.size(),
                passedCount,
                failedCount,
                courseGrades);
    }

    private boolean isPassed(StudentCourseGradeDTO grade) {
        return "passed".equals(grade.status())
                || (grade.score() != null && grade.score().compareTo(BigDecimal.valueOf(60)) >= 0);
    }

    private boolean isFailed(StudentCourseGradeDTO grade) {
        return "failed".equals(grade.status())
                || (grade.score() != null && grade.score().compareTo(BigDecimal.valueOf(60)) < 0);
    }

    private BigDecimal calculateGpa(List<StudentCourseGradeDTO> grades) {
        BigDecimal weightedPoints = BigDecimal.ZERO;
        BigDecimal credits = BigDecimal.ZERO;
        for (StudentCourseGradeDTO grade : grades) {
            if (grade.gradePoint() == null || grade.credits() == null) {
                continue;
            }
            weightedPoints = weightedPoints.add(grade.gradePoint().multiply(grade.credits()));
            credits = credits.add(grade.credits());
        }
        return credits.signum() == 0 ? BigDecimal.ZERO
                : weightedPoints.divide(credits, 2, RoundingMode.HALF_UP);
    }

    private String termLabel(String term, String grade) {
        if (term == null || term.length() < 9) {
            return term;
        }
        int year;
        try {
            year = Integer.parseInt(term.substring(0, 4));
        } catch (NumberFormatException e) {
            return term;
        }
        int gradeYear;
        try {
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
