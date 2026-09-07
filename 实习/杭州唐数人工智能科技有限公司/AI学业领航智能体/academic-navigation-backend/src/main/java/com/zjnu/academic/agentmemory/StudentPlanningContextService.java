package com.zjnu.academic.agentmemory;

import com.zjnu.academic.development.dto.DevelopmentPathDTO;
import com.zjnu.academic.development.service.DevelopmentService;
import com.zjnu.academic.student.dto.AcademicProfileDTO;
import com.zjnu.academic.student.dto.ProfileDimensionDTO;
import com.zjnu.academic.student.dto.StudentDashboardDTO;
import com.zjnu.academic.student.service.StudentService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;

/** Builds ephemeral, current academic context. This content is never sent to Mem0. */
@Service
public class StudentPlanningContextService {

    private static final Logger log = LoggerFactory.getLogger(StudentPlanningContextService.class);

    private final StudentService studentService;
    private final DevelopmentService developmentService;

    public StudentPlanningContextService(StudentService studentService, DevelopmentService developmentService) {
        this.studentService = studentService;
        this.developmentService = developmentService;
    }

    public String currentContext(Long userId) {
        try {
            StudentDashboardDTO dashboard = studentService.dashboard(userId, null);
            AcademicProfileDTO profile = studentService.profile(userId, null);
            List<DevelopmentPathDTO> paths = developmentService.paths(userId);

            StringBuilder context = new StringBuilder("以下为学校系统提供的当前学业摘要，仅限本次回复使用；不得保存、复述为长期档案或推测不存在的数据。\n");
            context.append("GPA：").append(value(dashboard.gpa()))
                    .append("；专业排名：").append(value(dashboard.rank())).append('/').append(value(dashboard.totalStudents()))
                    .append("；已获学分：").append(value(dashboard.credits())).append('/').append(value(dashboard.totalCredits()))
                    .append('\n');

            List<ProfileDimensionDTO> safeDimensions = profile.dimensions() == null ? List.of()
                    : profile.dimensions().stream()
                    .filter(this::isSafeAcademicDimension)
                    .limit(6)
                    .toList();
            if (!safeDimensions.isEmpty()) {
                context.append("五维画像：");
                for (ProfileDimensionDTO dimension : safeDimensions) {
                    context.append(dimension.label()).append(' ').append(value(dimension.score())).append("；");
                }
                context.append('\n');
            }
            if (paths != null && !paths.isEmpty()) {
                context.append("发展路径匹配度：");
                for (DevelopmentPathDTO path : paths) {
                    context.append(path.label()).append(' ').append(path.matchScore()).append("/100；");
                }
            }
            return context.toString();
        } catch (Exception e) {
            log.warn("Could not build current student planning context: {}", e.getMessage());
            return "学校当前学业摘要暂不可用。不得编造 GPA、排名、预警或画像数据。";
        }
    }

    private static String value(Object value) {
        return value == null ? "暂缺" : String.valueOf(value);
    }

    private boolean isSafeAcademicDimension(ProfileDimensionDTO dimension) {
        String key = dimension.key() == null ? "" : dimension.key().toLowerCase();
        String label = dimension.label() == null ? "" : dimension.label();
        return !key.contains("health")
                && !key.contains("psych")
                && !label.contains("健康")
                && !label.contains("心理");
    }
}
