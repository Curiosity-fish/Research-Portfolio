package com.zjnu.academic.datasource.school;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.zjnu.academic.alert.entity.Alert;
import com.zjnu.academic.alert.entity.InterventionRecord;
import com.zjnu.academic.alert.mapper.AlertMapper;
import com.zjnu.academic.alert.mapper.InterventionRecordMapper;
import com.zjnu.academic.course.entity.Course;
import com.zjnu.academic.course.entity.CourseClass;
import com.zjnu.academic.course.entity.Grade;
import com.zjnu.academic.course.mapper.CourseClassMapper;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.department.entity.Department;
import com.zjnu.academic.department.mapper.DepartmentMapper;
import com.zjnu.academic.health.entity.HealthReport;
import com.zjnu.academic.health.mapper.HealthReportMapper;
import com.zjnu.academic.major.entity.Major;
import com.zjnu.academic.major.mapper.MajorMapper;
import com.zjnu.academic.student.entity.GpaHistory;
import com.zjnu.academic.student.entity.ProfileScore;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.entity.StudentClass;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.ProfileScoreMapper;
import com.zjnu.academic.student.mapper.StudentClassMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.teacher.entity.Teacher;
import com.zjnu.academic.teacher.mapper.TeacherMapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * 从校方教务 REST 接口拉取样本数据并 upsert 到本地库。
 * 业务层无需感知数据来源，同步完成后现有接口即可读到新数据。
 */
@Service
public class SchoolApiSyncService {

    private static final Logger log = LoggerFactory.getLogger(SchoolApiSyncService.class);

    private final SchoolApiClient client;
    private final DepartmentMapper departmentMapper;
    private final MajorMapper majorMapper;
    private final StudentClassMapper studentClassMapper;
    private final TeacherMapper teacherMapper;
    private final UserMapper userMapper;
    private final StudentMapper studentMapper;
    private final CourseMapper courseMapper;
    private final CourseClassMapper courseClassMapper;
    private final GradeMapper gradeMapper;
    private final GpaHistoryMapper gpaHistoryMapper;
    private final ProfileScoreMapper profileScoreMapper;
    private final HealthReportMapper healthReportMapper;
    private final AlertMapper alertMapper;
    private final InterventionRecordMapper interventionRecordMapper;
    private final ObjectMapper objectMapper;

    public SchoolApiSyncService(SchoolApiClient client,
                                DepartmentMapper departmentMapper,
                                MajorMapper majorMapper,
                                StudentClassMapper studentClassMapper,
                                TeacherMapper teacherMapper,
                                UserMapper userMapper,
                                StudentMapper studentMapper,
                                CourseMapper courseMapper,
                                CourseClassMapper courseClassMapper,
                                GradeMapper gradeMapper,
                                GpaHistoryMapper gpaHistoryMapper,
                                ProfileScoreMapper profileScoreMapper,
                                HealthReportMapper healthReportMapper,
                                AlertMapper alertMapper,
                                InterventionRecordMapper interventionRecordMapper) {
        this.client = client;
        this.departmentMapper = departmentMapper;
        this.majorMapper = majorMapper;
        this.studentClassMapper = studentClassMapper;
        this.teacherMapper = teacherMapper;
        this.userMapper = userMapper;
        this.studentMapper = studentMapper;
        this.courseMapper = courseMapper;
        this.courseClassMapper = courseClassMapper;
        this.gradeMapper = gradeMapper;
        this.gpaHistoryMapper = gpaHistoryMapper;
        this.profileScoreMapper = profileScoreMapper;
        this.healthReportMapper = healthReportMapper;
        this.alertMapper = alertMapper;
        this.interventionRecordMapper = interventionRecordMapper;
        this.objectMapper = new ObjectMapper()
                .registerModule(new JavaTimeModule())
                .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS)
                .disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES)
                .setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
    }

    public Map<String, Integer> syncAll() {
        Map<String, Integer> counts = new LinkedHashMap<>();
        syncList("/api/v1/departments", departmentMapper, Department.class, counts, "departments");
        syncList("/api/v1/majors", majorMapper, Major.class, counts, "majors");
        syncList("/api/v1/classes", studentClassMapper, StudentClass.class, counts, "classes");
        syncList("/api/v1/users", userMapper, User.class, counts, "users");
        syncList("/api/v1/teachers", teacherMapper, Teacher.class, counts, "teachers");

        List<JsonNode> students = fetchAllStudents();
        for (JsonNode student : students) {
            upsertById(studentMapper, student, Student.class);
        }
        counts.put("students", students.size());

        syncList("/api/v1/courses", courseMapper, Course.class, counts, "courses");
        syncList("/api/v1/course-classes", courseClassMapper, CourseClass.class, counts, "courseClasses");

        for (JsonNode student : students) {
            String studentId = student.path("student_id").asText();
            syncStudentChildren(studentId, counts);
        }

        syncList("/api/v1/interventions", interventionRecordMapper, InterventionRecord.class, counts, "interventions");
        log.info("School API sync finished: {}", counts);
        return counts;
    }

    private void syncStudentChildren(String studentId, Map<String, Integer> counts) {
        List<JsonNode> grades = client.fetchStudentGrades(studentId, null);
        for (JsonNode node : grades) {
            upsertById(gradeMapper, node, Grade.class);
        }
        counts.merge("grades", grades.size(), Integer::sum);

        List<JsonNode> gpaHistory = client.fetchStudentGpaHistory(studentId);
        for (JsonNode node : gpaHistory) {
            upsertById(gpaHistoryMapper, node, GpaHistory.class);
        }
        counts.merge("gpaHistory", gpaHistory.size(), Integer::sum);

        List<JsonNode> profiles = client.fetchStudentProfileScores(studentId, null);
        for (JsonNode node : profiles) {
            upsertById(profileScoreMapper, node, ProfileScore.class);
        }
        counts.merge("profileScores", profiles.size(), Integer::sum);

        List<JsonNode> health = client.fetchStudentHealthReports(studentId);
        for (JsonNode node : health) {
            upsertById(healthReportMapper, node, HealthReport.class);
        }
        counts.merge("healthReports", health.size(), Integer::sum);

        List<JsonNode> alerts = client.fetchStudentAlerts(studentId);
        for (JsonNode node : alerts) {
            upsertById(alertMapper, node, Alert.class);
        }
        counts.merge("alerts", alerts.size(), Integer::sum);
    }

    private List<JsonNode> fetchAllStudents() {
        List<JsonNode> result = new ArrayList<>();
        int page = 1;
        int size = 200;
        while (true) {
            JsonNode data = client.fetchStudents(page, size);
            data.path("list").forEach(result::add);
            int total = data.path("total").asInt(0);
            if (result.size() >= total || data.path("list").size() == 0) {
                break;
            }
            page++;
        }
        return result;
    }

    private <T> void syncList(String path,
                              BaseMapper<T> mapper,
                              Class<T> type,
                              Map<String, Integer> counts,
                              String key) {
        List<JsonNode> nodes = client.list(path);
        for (JsonNode node : nodes) {
            upsertById(mapper, node, type);
        }
        counts.put(key, nodes.size());
    }

    private <T> void upsertById(BaseMapper<T> mapper, JsonNode node, Class<T> type) {
        long id = node.path("id").asLong(0);
        if (id <= 0) {
            return;
        }
        T entity = objectMapper.convertValue(node, type);
        T existing = mapper.selectById(id);
        if (existing == null) {
            mapper.insert(entity);
        } else {
            mapper.updateById(entity);
        }
    }
}
