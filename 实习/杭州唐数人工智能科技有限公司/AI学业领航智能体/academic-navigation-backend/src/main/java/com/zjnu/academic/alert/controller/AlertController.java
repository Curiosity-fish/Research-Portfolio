package com.zjnu.academic.alert.controller;

import com.zjnu.academic.alert.dto.AlertItemDTO;
import com.zjnu.academic.alert.dto.AlertStatsDTO;
import com.zjnu.academic.alert.dto.InterventionRequest;
import com.zjnu.academic.alert.service.AlertService;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.common.security.SecurityUtil;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@Tag(name = "预警", description = "预警列表、详情、确认、干预与统计")
@RestController
@RequestMapping("/api/alerts")
public class AlertController {

    private final AlertService alertService;

    @Autowired
    public AlertController(AlertService alertService) {
        this.alertService = alertService;
    }

    @Operation(summary = "预警列表", description = "按角色数据范围返回，支持等级/状态/关键词过滤和分页")
    @GetMapping
    public Result<PageResult<AlertItemDTO>> list(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term,
            @RequestParam(name = "level", required = false) String level,
            @RequestParam(name = "status", required = false) String status,
            @RequestParam(name = "keyword", required = false) String keyword,
            @RequestParam(name = "page", defaultValue = "1") int page,
            @RequestParam(name = "size", defaultValue = "20") int size) {
        return Result.ok(alertService.list(
                SecurityUtil.currentUserId(authentication),
                SecurityUtil.currentRole(authentication),
                term, level, status, keyword, page, size));
    }

    @Operation(summary = "预警详情")
    @GetMapping("/{id}")
    public Result<AlertItemDTO> detail(
            Authentication authentication,
            @PathVariable(name = "id") Long id) {
        return Result.ok(alertService.detail(
                SecurityUtil.currentUserId(authentication),
                SecurityUtil.currentRole(authentication),
                id));
    }

    @Operation(summary = "预警未读数", description = "侧边栏红点，统计 pending 数量")
    @GetMapping("/unread-count")
    public Result<Integer> unreadCount(Authentication authentication) {
        return Result.ok(alertService.unreadCount(
                SecurityUtil.currentUserId(authentication),
                SecurityUtil.currentRole(authentication)));
    }

    @Operation(summary = "确认已看", description = "学生确认预警，pending 变为 processing")
    @PostMapping("/{id}/confirm")
    public Result<Void> confirm(Authentication authentication, @PathVariable(name = "id") Long id) {
        alertService.confirm(SecurityUtil.currentUserId(authentication), id);
        return Result.ok();
    }

    @Operation(summary = "提交干预", description = "教师端提交干预记录并更新预警状态")
    @PostMapping("/{id}/intervention")
    public Result<Void> intervention(
            Authentication authentication,
            @PathVariable(name = "id") Long id,
            @RequestBody InterventionRequest request) {
        alertService.intervention(
                SecurityUtil.currentUserId(authentication),
                SecurityUtil.currentRole(authentication),
                id, request);
        return Result.ok();
    }

    @Operation(summary = "预警统计", description = "按等级统计当前角色可见范围")
    @GetMapping("/stats")
    public Result<AlertStatsDTO> stats(
            Authentication authentication,
            @RequestParam(name = "term", required = false) String term) {
        return Result.ok(alertService.stats(
                SecurityUtil.currentUserId(authentication),
                SecurityUtil.currentRole(authentication),
                term));
    }
}
