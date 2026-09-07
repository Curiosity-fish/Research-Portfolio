package com.zjnu.academic.teacher.controller;

import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.common.security.SecurityUtil;
import com.zjnu.academic.student.dto.StudentDTO;
import com.zjnu.academic.teacher.dto.CourseStatsItemDTO;
import com.zjnu.academic.teacher.dto.GpaProgressDTO;
import com.zjnu.academic.teacher.dto.StudentDetailDTO;
import com.zjnu.academic.teacher.dto.TeacherDashboardDTO;
import com.zjnu.academic.teacher.dto.TeacherTermsDTO;
import com.zjnu.academic.teacher.service.TeacherService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@Tag(name = "班主任端", description = "班级驾驶舱、GPA 进退、课程统计、学生列表与详情")
@RestController
@RequestMapping("/api/teacher")
public class TeacherController {

    private final TeacherService teacherService;

    @Autowired
    public TeacherController(TeacherService teacherService) {
        this.teacherService = teacherService;
    }

    @Operation(summary = "班级可用学期")
    @GetMapping("/terms")
    public Result<TeacherTermsDTO> terms(Authentication authentication) {
        return Result.ok(teacherService.terms(SecurityUtil.currentUserId(authentication)));
    }

    @Operation(summary = "班级驾驶舱")
    @GetMapping("/dashboard")
    public Result<TeacherDashboardDTO> dashboard(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(teacherService.dashboard(SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "GPA 进退")
    @GetMapping("/gpa-progress")
    public Result<GpaProgressDTO> gpaProgress(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(teacherService.gpaProgress(SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "课程统计")
    @GetMapping("/course-stats")
    public Result<List<CourseStatsItemDTO>> courseStats(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(teacherService.courseStats(SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "班级学生列表")
    @GetMapping("/students")
    public Result<PageResult<StudentDTO>> students(
            Authentication authentication,
            @RequestParam(name = "keyword", required = false) String keyword,
            @RequestParam(name = "term", required = false) String term,
            @RequestParam(name = "page", defaultValue = "1") int page,
            @RequestParam(name = "size", defaultValue = "20") int size) {
        return Result.ok(teacherService.students(
                SecurityUtil.currentUserId(authentication), keyword, term, page, size));
    }

    @Operation(summary = "学生详情")
    @GetMapping("/students/{studentId}")
    public Result<StudentDetailDTO> studentDetail(
            Authentication authentication,
            @PathVariable(name = "studentId") String studentId,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(teacherService.studentDetail(
                SecurityUtil.currentUserId(authentication), studentId, term));
    }
}
