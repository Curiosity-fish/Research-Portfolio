package com.zjnu.academic.eagent;

import com.zjnu.academic.alert.mapper.AlertMapper;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.course.mapper.CourseMapper;
import com.zjnu.academic.course.mapper.GradeMapper;
import com.zjnu.academic.department.mapper.DepartmentMapper;
import com.zjnu.academic.major.mapper.MajorMapper;
import com.zjnu.academic.student.mapper.ProfileScoreMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.teacher.mapper.TeacherMapper;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.util.LinkedHashMap;
import java.util.Map;

@Tag(name = "E-Agent 数据库诊断", description = "供 E-Agent 验证是否已连接到本地学业数据库")
@RestController
@RequestMapping("/api/eagent")
public class EagentHealthController {

    private final DataSource dataSource;
    private final DepartmentMapper departmentMapper;
    private final MajorMapper majorMapper;
    private final TeacherMapper teacherMapper;
    private final StudentMapper studentMapper;
    private final CourseMapper courseMapper;
    private final GradeMapper gradeMapper;
    private final ProfileScoreMapper profileScoreMapper;
    private final AlertMapper alertMapper;

    @Autowired
    public EagentHealthController(DataSource dataSource,
                                  DepartmentMapper departmentMapper,
                                  MajorMapper majorMapper,
                                  TeacherMapper teacherMapper,
                                  StudentMapper studentMapper,
                                  CourseMapper courseMapper,
                                  GradeMapper gradeMapper,
                                  ProfileScoreMapper profileScoreMapper,
                                  AlertMapper alertMapper) {
        this.dataSource = dataSource;
        this.departmentMapper = departmentMapper;
        this.majorMapper = majorMapper;
        this.teacherMapper = teacherMapper;
        this.studentMapper = studentMapper;
        this.courseMapper = courseMapper;
        this.gradeMapper = gradeMapper;
        this.profileScoreMapper = profileScoreMapper;
        this.alertMapper = alertMapper;
    }

    @Operation(summary = "验证本地数据库连接", description = "返回当前数据库名、数据库时间和关键业务表行数")
    @GetMapping("/health")
    public Result<Map<String, Object>> health() {
        Map<String, Object> data = new LinkedHashMap<>();
        try (Connection connection = dataSource.getConnection();
             PreparedStatement statement = connection.prepareStatement("SELECT DATABASE(), NOW()");
             ResultSet resultSet = statement.executeQuery()) {
            if (resultSet.next()) {
                data.put("status", "connected");
                data.put("database", resultSet.getString(1));
                data.put("databaseTime", resultSet.getTimestamp(2).toString());
            } else {
                throw new BusinessException(500, "无法读取当前数据库信息");
            }
        } catch (java.sql.SQLException e) {
            throw new BusinessException(500, "本地数据库连接失败: " + e.getMessage());
        }

        Map<String, Long> counts = new LinkedHashMap<>();
        counts.put("departments", departmentMapper.selectCount(null));
        counts.put("majors", majorMapper.selectCount(null));
        counts.put("teachers", teacherMapper.selectCount(null));
        counts.put("students", studentMapper.selectCount(null));
        counts.put("courses", courseMapper.selectCount(null));
        counts.put("grades", gradeMapper.selectCount(null));
        counts.put("profileScores", profileScoreMapper.selectCount(null));
        counts.put("alerts", alertMapper.selectCount(null));
        data.put("counts", counts);
        return Result.ok(data);
    }
}
