package com.zjnu.academic.datasource.school;

import com.fasterxml.jackson.databind.JsonNode;
import com.zjnu.academic.datasource.AcademicDataSource;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.context.annotation.Primary;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * 校方教务接口数据源：provider=school 时生效。
 * 业务侧按需调用教务系统 REST 接口，当前提供 GPA/排名/学分三个查询方法。
 */
@Primary
@Component
@ConditionalOnProperty(name = "academic.datasource.provider", havingValue = "school")
public class SchoolApiDataSource implements AcademicDataSource {

    private final SchoolApiClient client;

    public SchoolApiDataSource(SchoolApiClient client) {
        this.client = client;
    }

    @Override
    public double getGpa(String studentAccount, String semester) {
        List<JsonNode> history = client.fetchStudentGpaHistory(studentAccount);
        JsonNode target = findTerm(history, semester);
        return target == null ? 0 : target.path("gpa").asDouble(0);
    }

    @Override
    public int[] getRank(String studentAccount, String semester) {
        List<JsonNode> history = client.fetchStudentGpaHistory(studentAccount);
        JsonNode target = findTerm(history, semester);
        if (target == null) {
            return new int[]{0, 0};
        }
        return new int[]{
                target.path("rank").asInt(0),
                target.path("total_students").asInt(0)
        };
    }

    @Override
    public int[] getCredits(String studentAccount) {
        JsonNode student = client.fetchStudent(studentAccount);
        if (student == null || student.isMissingNode()) {
            return new int[]{0, 0};
        }
        return new int[]{
                (int) Math.round(student.path("total_credits").asDouble(0)),
                (int) Math.round(student.path("required_credits").asDouble(0))
        };
    }

    private JsonNode findTerm(List<JsonNode> history, String semester) {
        if (history == null || history.isEmpty()) {
            return null;
        }
        if (semester != null && !semester.isBlank()) {
            return history.stream()
                    .filter(node -> semester.equals(node.path("term").asText()))
                    .findFirst()
                    .orElse(null);
        }
        return history.get(history.size() - 1);
    }
}
