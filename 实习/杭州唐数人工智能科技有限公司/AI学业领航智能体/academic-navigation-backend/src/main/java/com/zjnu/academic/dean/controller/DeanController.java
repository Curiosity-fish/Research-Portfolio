package com.zjnu.academic.dean.controller;

import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.common.security.SecurityUtil;
import com.zjnu.academic.dean.dto.DeanAlertStatsDTO;
import com.zjnu.academic.dean.dto.DeanAlertTrendDTO;
import com.zjnu.academic.dean.dto.DeanDashboardDTO;
import com.zjnu.academic.dean.service.DeanService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@Tag(name = "院长端", description = "全院仪表盘、预警统计与预警趋势")
@RestController
@RequestMapping("/api/dean")
public class DeanController {

    private final DeanService deanService;

    @Autowired
    public DeanController(DeanService deanService) {
        this.deanService = deanService;
    }

    @Operation(summary = "可选学期", description = "返回当前学院学生数据中存在的学期")
    @GetMapping("/terms")
    public Result<List<String>> terms(Authentication authentication) {
        return Result.ok(deanService.availableTerms(
                SecurityUtil.currentUserId(authentication)));
    }

    @Operation(summary = "全院仪表盘", description = "全院 KPI、专业排名与预警分布")
    @GetMapping("/dashboard")
    public Result<DeanDashboardDTO> dashboard(
            Authentication authentication,
            @RequestParam(required = false) String term) {
        return Result.ok(deanService.dashboard(
                SecurityUtil.currentUserId(authentication), term));
    }

    @Operation(summary = "全院预警统计", description = "按专业/年级统计预警总览")
    @GetMapping("/alerts")
    public Result<DeanAlertStatsDTO> alerts(
            Authentication authentication,
            @RequestParam(required = false) String term,
            @RequestParam(required = false) String groupBy) {
        return Result.ok(deanService.alerts(
                SecurityUtil.currentUserId(authentication), term, groupBy));
    }

    @Operation(summary = "预警趋势", description = "近四学期预警趋势，terms 参数逗号分隔")
    @GetMapping("/alerts/trend")
    public Result<DeanAlertTrendDTO> alertTrend(
            Authentication authentication,
            @RequestParam(required = false) String terms) {
        return Result.ok(deanService.alertTrend(
                SecurityUtil.currentUserId(authentication), terms));
    }
}
