package com.zjnu.academic.course.controller;

import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.common.security.SecurityUtil;
import com.zjnu.academic.course.dto.CourseTeacherCourseDTO;
import com.zjnu.academic.course.service.CourseTeacherService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@Tag(name = "任课教师端", description = "我的课程、课程成绩统计与所授课程预警")
@RestController
@RequestMapping("/api/course-teacher")
public class CourseTeacherController {

    private final CourseTeacherService courseTeacherService;

    public CourseTeacherController(CourseTeacherService courseTeacherService) {
        this.courseTeacherService = courseTeacherService;
    }

    @Operation(summary = "我的课程列表")
    @GetMapping("/courses")
    public Result<List<CourseTeacherCourseDTO>> courses(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(courseTeacherService.courses(
                SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "授课学期列表")
    @GetMapping("/terms")
    public Result<List<String>> terms(Authentication authentication) {
        return Result.ok(courseTeacherService.terms(
                SecurityUtil.currentUserId(authentication)));
    }

    @Operation(summary = "所授课程预警")
    @GetMapping("/alerts")
    public Result<PageResult<AlertItemDTO>> alerts(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term,
            @RequestParam(name = "page", defaultValue = "1") int page,
            @RequestParam(name = "size", defaultValue = "20") int size) {
        return Result.ok(courseTeacherService.alerts(
                SecurityUtil.currentUserId(authentication), term, page, size));
    }
}
