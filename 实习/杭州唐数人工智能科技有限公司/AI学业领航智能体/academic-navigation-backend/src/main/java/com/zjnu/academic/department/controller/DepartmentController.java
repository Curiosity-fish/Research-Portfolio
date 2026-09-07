package com.zjnu.academic.department.controller;

import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.common.security.SecurityUtil;
import com.zjnu.academic.department.dto.CourseHeatmapDTO;
import com.zjnu.academic.department.dto.DepartmentAlertDTO;
import com.zjnu.academic.department.dto.DepartmentCourseDTO;
import com.zjnu.academic.department.dto.DepartmentOverviewDTO;
import com.zjnu.academic.department.dto.PlanAnalysisDTO;
import com.zjnu.academic.department.service.DepartmentService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@Tag(name = "系主任端", description = "专业态势、课程热度、课程明细、培养方案与专业预警")
@RestController
@RequestMapping("/api/department")
public class DepartmentController {

    private final DepartmentService departmentService;

    @Autowired
    public DepartmentController(DepartmentService departmentService) {
        this.departmentService = departmentService;
    }

    @Operation(summary = "可选学期", description = "返回当前专业学生数据中存在的学期")
    @GetMapping("/terms")
    public Result<List<String>> terms(Authentication authentication) {
        return Result.ok(departmentService.availableTerms(
                SecurityUtil.currentUserId(authentication)));
    }

    @Operation(summary = "专业态势总览")
    @GetMapping("/overview")
    public Result<DepartmentOverviewDTO> overview(
            Authentication authentication,
            @RequestParam(required = false) String term) {
        return Result.ok(departmentService.overview(
                SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "课程热度热力图", description = "直接返回 ECharts 兼容的 courses/grades/heatData")
    @GetMapping("/courses/heatmap")
    public Result<CourseHeatmapDTO> courseHeatmap(
            Authentication authentication,
            @RequestParam(required = false) String term) {
        return Result.ok(departmentService.courseHeatmap(
                SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "课程明细", description = "各课程各年级均分/通过率/高分率/不及格率")
    @GetMapping("/courses")
    public Result<List<DepartmentCourseDTO>> courses(
            Authentication authentication,
            @RequestParam(required = false) String term) {
        return Result.ok(departmentService.courses(
                SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "培养方案分析", description = "维度顺序固定：公共基础、专业必修、专业选修、实践环节、通识教育")
    @GetMapping("/plan-analysis")
    public Result<PlanAnalysisDTO> planAnalysis(
            Authentication authentication,
            @RequestParam(required = false) String grade) {
        return Result.ok(departmentService.planAnalysis(
                SecurityUtil.currentUserId(authentication), grade));
    }

    @Operation(summary = "专业预警列表")
    @GetMapping("/alerts")
    public Result<PageResult<DepartmentAlertDTO>> alerts(
            Authentication authentication,
            @RequestParam(required = false) String term,
            @RequestParam(required = false) String grade,
            @RequestParam(required = false) String level,
            @RequestParam(required = false) String status,
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int size) {
        return Result.ok(departmentService.alerts(
                SecurityUtil.currentUserId(authentication),
                term, grade, level, status, page, size));
    }
}
