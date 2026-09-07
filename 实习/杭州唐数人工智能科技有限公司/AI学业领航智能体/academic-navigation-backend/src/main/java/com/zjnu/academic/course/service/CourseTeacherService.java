package com.zjnu.academic.course.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.alert.service.AlertService;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.course.dto.CourseClassStatDTO;
import com.zjnu.academic.course.dto.CourseTeacherCourseDTO;
import com.zjnu.academic.course.dto.TermFailStudentDTO;
import com.zjnu.academic.course.entity.Course;
import com.zjnu.academic.course.entity.CourseClass;
import com.zjnu.academic.course.entity.Grade;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.teacher.dto.FailStudentDTO;
import com.zjnu.academic.teacher.entity.Teacher;
import com.zjnu.academic.teacher.mapper.TeacherMapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;
import java.util.stream.Collectors;

@Service
public class CourseTeacherService {

    private final TeacherMapper teacherMapper;
    private final CourseMapper courseMapper;
    private final CourseClassMapper courseClassMapper;
    private final GradeMapper gradeMapper;
    private final StudentMapper studentMapper;
    private final UserMapper userMapper;
    private final AlertService alertService;

    public CourseTeacherService(TeacherMapper teacherMapper,
                                CourseMapper courseMapper,
                                CourseClassMapper courseClassMapper,
                                GradeMapper gradeMapper,
                                StudentMapper studentMapper,
                                UserMapper userMapper,
                                AlertService alertService) {
        this.teacherMapper = teacherMapper;
        this.courseMapper = courseMapper;
        this.courseClassMapper = courseClassMapper;
        this.gradeMapper = gradeMapper;
        this.studentMapper = studentMapper;
        this.userMapper = userMapper;
        this.alertService = alertService;
    }

    public List<CourseTeacherCourseDTO> courses(Long userId, String term) {
        Teacher teacher = teacherMapper.selectOne(
                new LambdaQueryWrapper<Teacher>().eq(Teacher::getUserId, userId));
        if (teacher == null) {
            return List.of();
        }

        String resolvedTerm = term == null || term.isBlank()
                ? terms(userId).stream().findFirst().orElse("") : term;
        LambdaQueryWrapper<CourseClass> wrapper = new LambdaQueryWrapper<CourseClass>()
                .eq(CourseClass::getTeacherId, teacher.getId())
                .eq(CourseClass::getTerm, resolvedTerm);
        List<CourseClass> courseClasses = courseClassMapper.selectList(wrapper);
        if (courseClasses.isEmpty()) {
            return List.of();
        }

        List<Long> courseIds = courseClasses.stream()
                .map(CourseClass::getCourseId)
                .distinct()
                .toList();
        Map<Long, Course> courses = courseMapper.selectBatchIds(courseIds).stream()
                .collect(Collectors.toMap(Course::getId, c -> c));
        List<Long> courseClassIds = courseClasses.stream().map(CourseClass::getId).toList();
        List<Grade> grades = gradeMapper.selectList(
                new LambdaQueryWrapper<Grade>().in(Grade::getCourseClassId, courseClassIds));
        Map<Long, List<Grade>> gradesByClass = grades.stream()
                .collect(Collectors.groupingBy(Grade::getCourseClassId));
        List<Long> studentIds = grades.stream().map(Grade::getStudentId).distinct().toList();
        Map<Long, Student> students = studentIds.isEmpty()
                ? Map.of()
                : studentMapper.selectBatchIds(studentIds).stream()
                        .collect(Collectors.toMap(Student::getId, s -> s));
        List<Long> userIds = students.values().stream().map(Student::getUserId).distinct().toList();
        Map<Long, User> users = userIds.isEmpty()
                ? Map.of()
                : userMapper.selectBatchIds(userIds).stream()
                        .collect(Collectors.toMap(User::getId, u -> u));
        Set<Long> warnedStudentIds = warnedStudentIds(studentIds, students, resolvedTerm);
        Map<Long, List<FailedCourseGrade>> termFailuresByStudent =
                termFailuresByStudent(studentIds, resolvedTerm);

        Map<Long, List<CourseClass>> classesByCourse = courseClasses.stream()
                .collect(Collectors.groupingBy(CourseClass::getCourseId, LinkedHashMap::new, Collectors.toList()));

        return classesByCourse.entrySet().stream()
                .sorted(Map.Entry.comparingByKey())
                .map(entry -> buildCourse(
                        entry.getKey(),
                        entry.getValue(),
                        courses,
                        gradesByClass,
                        students,
                        users,
                        warnedStudentIds,
                        termFailuresByStudent))
                .toList();
    }

    public List<String> terms(Long userId) {
        Teacher teacher = teacherMapper.selectOne(
                new LambdaQueryWrapper<Teacher>().eq(Teacher::getUserId, userId));
        if (teacher == null) {
            return List.of();
        }
        List<CourseClass> classes = courseClassMapper.selectList(new LambdaQueryWrapper<CourseClass>()
                .eq(CourseClass::getTeacherId, teacher.getId()));
        if (classes.isEmpty()) {
            return List.of();
        }
        Set<Long> gradedClassIds = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                        .in(Grade::getCourseClassId, classes.stream().map(CourseClass::getId).toList())
                        .select(Grade::getCourseClassId))
                .stream().map(Grade::getCourseClassId).collect(Collectors.toSet());
        return classes.stream()
                .filter(courseClass -> gradedClassIds.contains(courseClass.getId()))
                .map(CourseClass::getTerm)
                .filter(value -> value != null && !value.isBlank())
                .distinct().sorted(Comparator.reverseOrder()).toList();
    }

    public PageResult<AlertItemDTO> alerts(Long userId, String term, int page, int size) {
        Teacher teacher = teacherMapper.selectOne(
                new LambdaQueryWrapper<Teacher>().eq(Teacher::getUserId, userId));
        if (teacher == null) {
            return PageResult.of(List.of(), 0, page, size);
        }
        String resolvedTerm = term == null || term.isBlank()
                ? terms(userId).stream().findFirst().orElse("") : term;
        List<CourseClass> taughtClasses = courseClassMapper.selectList(new LambdaQueryWrapper<CourseClass>()
                .eq(CourseClass::getTeacherId, teacher.getId())
                .eq(CourseClass::getTerm, resolvedTerm));
        if (taughtClasses.isEmpty()) {
            return PageResult.of(List.of(), 0, page, size);
        }
        List<Long> classIds = taughtClasses.stream().map(CourseClass::getId).toList();
        List<Grade> taughtGrades = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                .in(Grade::getCourseClassId, classIds));
        List<Long> studentIds = taughtGrades.stream().map(Grade::getStudentId).distinct().toList();
        if (studentIds.isEmpty()) {
            return PageResult.of(List.of(), 0, page, size);
        }
        List<Long> courseIds = taughtClasses.stream().map(CourseClass::getCourseId).distinct().toList();
        Map<Long, Course> taughtCourses = courseMapper.selectBatchIds(courseIds).stream()
                .collect(Collectors.toMap(Course::getId, course -> course));
        Map<Long, CourseClass> taughtClassById = taughtClasses.stream()
                .collect(Collectors.toMap(CourseClass::getId, courseClass -> courseClass));
        Map<Long, Set<Long>> courseIdsByStudent = new HashMap<>();
        for (Grade grade : taughtGrades) {
            CourseClass courseClass = taughtClassById.get(grade.getCourseClassId());
            if (courseClass != null && taughtCourses.containsKey(courseClass.getCourseId())) {
                courseIdsByStudent.computeIfAbsent(grade.getStudentId(), ignored -> new LinkedHashSet<>())
                        .add(courseClass.getCourseId());
            }
        }
        Map<Long, Student> scopedStudents = studentMapper.selectBatchIds(studentIds).stream()
                .collect(Collectors.toMap(Student::getId, student -> student));
        Map<String, List<Course>> coursesByStudentNo = new HashMap<>();
        for (Map.Entry<Long, Set<Long>> entry : courseIdsByStudent.entrySet()) {
            Student student = scopedStudents.get(entry.getKey());
            if (student == null) {
                continue;
            }
            List<Course> enrolledTaughtCourses = entry.getValue().stream()
                    .map(taughtCourses::get)
                    .filter(Objects::nonNull)
                    .sorted(Comparator.comparing(Course::getId))
                    .toList();
            coursesByStudentNo.put(student.getStudentId(), enrolledTaughtCourses);
        }

        PageResult<AlertItemDTO> result = alertService.listForStudents(
                studentIds, resolvedTerm, null, null, null, 1, Integer.MAX_VALUE);
        Map<String, AlertItemDTO> latestAlertByStudent = result.list().stream()
                .collect(Collectors.toMap(
                        AlertItemDTO::studentId,
                        alert -> alert,
                        (latest, ignored) -> latest,
                        LinkedHashMap::new));
        List<AlertItemDTO> expanded = latestAlertByStudent.values().stream()
                .flatMap(alert -> coursesByStudentNo.getOrDefault(alert.studentId(), List.of()).stream()
                        .map(course -> withTaughtCourse(alert, course)))
                .sorted(Comparator.comparing(AlertItemDTO::date).reversed()
                        .thenComparing(AlertItemDTO::studentId)
                        .thenComparing(AlertItemDTO::course))
                .toList();
        int safePage = Math.max(page, 1);
        int safeSize = Math.max(size, 1);
        int from = Math.min((safePage - 1) * safeSize, expanded.size());
        int to = Math.min(from + safeSize, expanded.size());
        return PageResult.of(expanded.subList(from, to), expanded.size(), safePage, safeSize);
    }

    private AlertItemDTO withTaughtCourse(AlertItemDTO alert, Course course) {
        return new AlertItemDTO(
                alert.id() + "-" + course.getId(), alert.studentId(), alert.studentName(),
                alert.type(), alert.level(), alert.title(), alert.description(), course.getName(),
                alert.failedCourses(), alert.date(),
                alert.status(), alert.suggestion(), alert.triggerEvent(), alert.pushedAt());
    }

    private Set<Long> warnedStudentIds(List<Long> studentIds,
                                       Map<Long, Student> students,
                                       String term) {
        if (studentIds.isEmpty()) {
            return Set.of();
        }
        Set<String> warnedStudentNumbers = alertService.listForStudents(
                        studentIds, term, null, null, null, 1, Integer.MAX_VALUE)
                .list().stream()
                .map(AlertItemDTO::studentId)
                .collect(Collectors.toSet());
        return students.entrySet().stream()
                .filter(entry -> warnedStudentNumbers.contains(entry.getValue().getStudentId()))
                .map(Map.Entry::getKey)
                .collect(Collectors.toSet());
    }

    private Map<Long, List<FailedCourseGrade>> termFailuresByStudent(List<Long> studentIds,
                                                                      String term) {
        if (studentIds.isEmpty()) {
            return Map.of();
        }
        List<Grade> failedGrades = gradeMapper.selectList(new LambdaQueryWrapper<Grade>()
                .in(Grade::getStudentId, studentIds)
                .eq(Grade::getTerm, term)
                .and(wrapper -> wrapper.eq(Grade::getStatus, "failed")
                        .or().eq(Grade::getStatus, "retake")
                        .or().lt(Grade::getScore, 60)));
        if (failedGrades.isEmpty()) {
            return Map.of();
        }
        List<Long> failedClassIds = failedGrades.stream()
                .map(Grade::getCourseClassId).distinct().toList();
        Map<Long, CourseClass> failedClasses = courseClassMapper.selectBatchIds(failedClassIds).stream()
                .collect(Collectors.toMap(CourseClass::getId, courseClass -> courseClass));
        List<Long> failedCourseIds = failedClasses.values().stream()
                .map(CourseClass::getCourseId).distinct().toList();
        Map<Long, Course> failedCourses = courseMapper.selectBatchIds(failedCourseIds).stream()
                .collect(Collectors.toMap(Course::getId, course -> course));

        return failedGrades.stream().collect(Collectors.groupingBy(
                Grade::getStudentId,
                Collectors.mapping(grade -> {
                    CourseClass courseClass = failedClasses.get(grade.getCourseClassId());
                    Course course = courseClass == null ? null : failedCourses.get(courseClass.getCourseId());
                    return new FailedCourseGrade(
                            course == null ? "未知课程" : course.getName(),
                            grade.getScore());
                }, Collectors.toList())));
    }

    private CourseTeacherCourseDTO buildCourse(Long courseId,
                                               List<CourseClass> classes,
                                               Map<Long, Course> courses,
                                               Map<Long, List<Grade>> gradesByClass,
                                               Map<Long, Student> students,
                                               Map<Long, User> users,
                                               Set<Long> warnedStudentIds,
                                               Map<Long, List<FailedCourseGrade>> termFailuresByStudent) {
        Course course = courses.get(courseId);
        List<Grade> allGrades = classes.stream()
                .flatMap(c -> gradesByClass.getOrDefault(c.getId(), List.of()).stream())
                .toList();
        int totalStudents = (int) allGrades.stream().map(Grade::getStudentId).distinct().count();
        double avgScore = average(allGrades.stream().map(Grade::getScore).filter(Objects::nonNull).toList());
        int passed = (int) allGrades.stream().filter(g -> "passed".equals(g.getStatus())).count();
        double passRate = percent(passed, allGrades.size());
        int highRisk = (int) allGrades.stream()
                .map(Grade::getStudentId)
                .distinct()
                .filter(warnedStudentIds::contains)
                .count();
        List<Integer> distribution = scoreDistribution(allGrades);

        List<CourseClassStatDTO> classStats = classes.stream()
                .map(c -> buildClassStat(c, gradesByClass.getOrDefault(c.getId(), List.of()),
                        students, users, warnedStudentIds, termFailuresByStudent))
                .toList();

        String term = classes.stream().map(CourseClass::getTerm).findFirst().orElse("");
        return new CourseTeacherCourseDTO(
                course == null ? String.valueOf(courseId) : String.valueOf(course.getId()),
                course == null ? "未知课程" : course.getName(),
                course == null ? "" : course.getCode(),
                term,
                totalStudents,
                round1(avgScore),
                round1(passRate),
                highRisk,
                distribution,
                classStats);
    }

    private CourseClassStatDTO buildClassStat(CourseClass courseClass,
                                              List<Grade> grades,
                                              Map<Long, Student> students,
                                              Map<Long, User> users,
                                              Set<Long> warnedStudentIds,
                                              Map<Long, List<FailedCourseGrade>> termFailuresByStudent) {
        double avg = average(grades.stream().map(Grade::getScore).filter(Objects::nonNull).toList());
        int passed = (int) grades.stream().filter(g -> "passed".equals(g.getStatus())).count();
        double pass = percent(passed, grades.size());
        double high = grades.stream().map(Grade::getScore).filter(Objects::nonNull)
                .mapToDouble(Number::doubleValue).max().orElse(0);
        double low = grades.stream().map(Grade::getScore).filter(Objects::nonNull)
                .mapToDouble(Number::doubleValue).min().orElse(0);
        int risk = (int) grades.stream()
                .map(Grade::getStudentId)
                .distinct()
                .filter(warnedStudentIds::contains)
                .count();
        List<FailStudentDTO> failStudents = grades.stream()
                .filter(g -> "failed".equals(g.getStatus()) || "retake".equals(g.getStatus())
                        || (g.getScore() != null && g.getScore().doubleValue() < 60))
                .map(g -> {
                    Student student = students.get(g.getStudentId());
                    User user = student == null ? null : users.get(student.getUserId());
                    return new FailStudentDTO(
                            user == null ? "" : user.getName(),
                            student == null ? "" : student.getStudentId(),
                            g.getScore() == null ? null : g.getScore());
                })
                .sorted(Comparator.comparing(FailStudentDTO::score,
                        Comparator.nullsLast(Comparator.naturalOrder())))
                .toList();
        List<TermFailStudentDTO> termFailStudents = grades.stream()
                .map(Grade::getStudentId)
                .distinct()
                .filter(termFailuresByStudent::containsKey)
                .map(studentId -> {
                    Student student = students.get(studentId);
                    User user = student == null ? null : users.get(student.getUserId());
                    List<FailedCourseGrade> failures = termFailuresByStudent.get(studentId);
                    List<String> failedCourses = failures.stream()
                            .map(FailedCourseGrade::courseName)
                            .distinct()
                            .sorted()
                            .toList();
                    java.math.BigDecimal lowestScore = failures.stream()
                            .map(FailedCourseGrade::score)
                            .filter(Objects::nonNull)
                            .min(Comparator.naturalOrder())
                            .orElse(null);
                    return new TermFailStudentDTO(
                            user == null ? "" : user.getName(),
                            student == null ? "" : student.getStudentId(),
                            failedCourses,
                            lowestScore);
                })
                .sorted(Comparator.comparing(TermFailStudentDTO::lowestScore,
                                Comparator.nullsLast(Comparator.naturalOrder()))
                        .thenComparing(TermFailStudentDTO::studentId))
                .toList();
        return new CourseClassStatDTO(
                courseClass.getClassName(),
                round1(avg),
                round1(pass),
                round1(high),
                round1(low),
                grades.size(),
                risk,
                failStudents,
                termFailStudents);
    }

    private record FailedCourseGrade(String courseName, java.math.BigDecimal score) {
    }

    private List<Integer> scoreDistribution(List<Grade> grades) {
        int[] buckets = new int[6];
        for (Grade grade : grades) {
            if (grade.getScore() == null) {
                continue;
            }
            double score = grade.getScore().doubleValue();
            if (score < 60) {
                buckets[0]++;
            } else if (score < 70) {
                buckets[1]++;
            } else if (score < 80) {
                buckets[2]++;
            } else if (score < 90) {
                buckets[3]++;
            } else if (score < 95) {
                buckets[4]++;
            } else {
                buckets[5]++;
            }
        }
        List<Integer> result = new ArrayList<>();
        for (int bucket : buckets) {
            result.add(bucket);
        }
        return result;
    }

    private double average(List<java.math.BigDecimal> values) {
        if (values.isEmpty()) {
            return 0;
        }
        return values.stream().mapToDouble(java.math.BigDecimal::doubleValue).average().orElse(0);
    }

    private double percent(int part, int total) {
        if (total == 0) {
            return 0;
        }
        return part * 100.0 / total;
    }

    private double round1(double value) {
        return Math.round(value * 10) / 10.0;
    }
}
