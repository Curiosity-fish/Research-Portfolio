package com.zjnu.academic.datasource.school;

import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.compute.AcademicComputeService;
import com.zjnu.academic.config.SchoolApiProperties;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;
import java.util.LinkedHashMap;

@RestController
@RequestMapping("/api/school-sync")
public class SchoolSyncController {

    private final SchoolApiSyncService syncService;
    private final SchoolApiProperties properties;
    private final AcademicComputeService computeService;
    private final String provider;

    public SchoolSyncController(SchoolApiSyncService syncService,
                                SchoolApiProperties properties,
                                AcademicComputeService computeService,
                                @Value("${academic.datasource.provider:mock}") String provider) {
        this.syncService = syncService;
        this.properties = properties;
        this.computeService = computeService;
        this.provider = provider;
    }

    @PostMapping
    public Result<Map<String, Integer>> sync() {
        Map<String, Integer> result = new LinkedHashMap<>(syncService.syncAll());
        computeService.recomputeAll().forEach((key, value) -> result.put("computed." + key, value));
        return Result.ok(result);
    }

    @io.swagger.v3.oas.annotations.Operation(summary = "触发派生数据计算", description = "从原始成绩/体测/竞赛/志愿/考勤/心理测评计算GPA历史、五维画像与基于挂科和缺勤的预警")
    @PostMapping("/compute")
    public Result<Map<String, Integer>> compute() {
        return Result.ok(computeService.recomputeAll());
    }

    @GetMapping("/status")
    public Result<Map<String, String>> status() {
        return Result.ok(Map.of(
                "provider", provider,
                "baseUrl", properties.getBaseUrl(),
                "tokenConfigured", String.valueOf(!properties.getToken().isBlank())
        ));
    }
}
