package com.zjnu.academic.development.controller;

import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.common.security.SecurityUtil;
import com.zjnu.academic.development.dto.DevelopmentAnalyzeDTO;
import com.zjnu.academic.development.dto.DevelopmentAnalyzeRequest;
import com.zjnu.academic.development.dto.DevelopmentPathDTO;
import com.zjnu.academic.development.dto.JobItemDTO;
import com.zjnu.academic.development.dto.SchoolItemDTO;
import com.zjnu.academic.development.service.DevelopmentService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@Tag(name = "发展引导", description = "五条发展路径、职位推荐、院校推荐与目标差距拆解")
@RestController
@RequestMapping("/api/development")
public class DevelopmentController {

    private final DevelopmentService developmentService;

    @Autowired
    public DevelopmentController(DevelopmentService developmentService) {
        this.developmentService = developmentService;
    }

    @Operation(summary = "发展路径分析", description = "按当前学生画像返回五条路径及匹配度/差距/行动/里程碑")
    @GetMapping("/paths")
    public Result<List<DevelopmentPathDTO>> paths(Authentication authentication) {
        return Result.ok(developmentService.paths(
                SecurityUtil.currentUserId(authentication)));
    }

    @Operation(summary = "职位推荐", description = "按岗位分类分页读取职位缓存")
    @GetMapping("/jobs")
    public Result<PageResult<JobItemDTO>> jobs(
            Authentication authentication,
            @RequestParam(required = false) String category,
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int size) {
        return Result.ok(developmentService.jobs(category, page, size));
    }

    @Operation(summary = "院校推荐", description = "type=grad|overseas，可按 tier 与分页筛选")
    @GetMapping("/schools")
    public Result<PageResult<SchoolItemDTO>> schools(
            Authentication authentication,
            @RequestParam String type,
            @RequestParam(required = false) String tier,
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "20") int size) {
        return Result.ok(developmentService.schools(type, tier, page, size));
    }

    @Operation(summary = "目标差距分析", description = "提交目标，返回差距反向拆解与建议")
    @PostMapping("/analyze")
    public Result<DevelopmentAnalyzeDTO> analyze(
            Authentication authentication,
            @RequestBody DevelopmentAnalyzeRequest request) {
        return Result.ok(developmentService.analyze(
                SecurityUtil.currentUserId(authentication), request));
    }
}
