package com.zjnu.academic.student.controller;

import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.common.security.SecurityUtil;
import com.zjnu.academic.student.dto.AcademicProfileDTO;
import com.zjnu.academic.student.dto.GpaTrendItemDTO;
import com.zjnu.academic.student.dto.HealthReportDTO;
import com.zjnu.academic.student.dto.StudentDashboardDTO;
import com.zjnu.academic.student.dto.StudentTermGradesDTO;
import com.zjnu.academic.student.service.StudentService;
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

@Tag(name = "学生端", description = "学业总览、画像、GPA 趋势、个人预警与体测报告")
@RestController
@RequestMapping("/api/student")
public class StudentController {

    private final StudentService studentService;

    @Autowired
    public StudentController(StudentService studentService) {
        this.studentService = studentService;
    }

    @Operation(summary = "学业总览")
    @GetMapping("/dashboard")
    public Result<StudentDashboardDTO> dashboard(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(studentService.dashboard(SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "可查看成绩的学期")
    @GetMapping("/terms")
    public Result<List<String>> terms(Authentication authentication) {
        return Result.ok(studentService.availableTerms(SecurityUtil.currentUserId(authentication)));
    }

    @Operation(summary = "指定学期课程成绩")
    @GetMapping("/grades")
    public Result<StudentTermGradesDTO> grades(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(studentService.termGrades(SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "学业画像")
    @GetMapping("/profile")
    public Result<AcademicProfileDTO> profile(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(studentService.profile(SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "GPA 历史")
    @GetMapping("/gpa-trend")
    public Result<List<GpaTrendItemDTO>> gpaTrend(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(studentService.gpaTrend(SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "个人预警")
    @GetMapping("/alerts")
    public Result<PageResult<AlertItemDTO>> alerts(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term,
            @RequestParam(name = "page", defaultValue = "1") int page,
            @RequestParam(name = "size", defaultValue = "20") int size) {
        return Result.ok(studentService.alerts(SecurityUtil.currentUserId(authentication), term, page, size));
    }

    @Operation(summary = "体测报告列表")
    @GetMapping("/health-reports")
    public Result<List<HealthReportDTO>> healthReports(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(studentService.healthReports(SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "浣撴祴鎶ュ憡璇︽儏")
    @GetMapping("/health-reports/{id}")
    public Result<HealthReportDTO> healthReportDetail(
            Authentication authentication,
            @PathVariable Long id) {
        return Result.ok(studentService.healthReportDetail(SecurityUtil.currentUserId(authentication), id));
    }
}
