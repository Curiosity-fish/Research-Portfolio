package com.zjnu.academic.agentmemory;

import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.common.security.SecurityUtil;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@Tag(name = "人生规划长期记忆", description = "受登录身份约束的低敏长期记忆与人生规划对话代理")
@RestController
@RequestMapping("/api/agents/life-planning")
public class LifePlanningAgentController {

    private final LifePlanningAgentService service;

    public LifePlanningAgentController(LifePlanningAgentService service) {
        this.service = service;
    }

    @Operation(summary = "人生规划对话", description = "后端注入长期记忆并代理 E-Agent；仅学生可调用")
    @PostMapping("/chat")
    public Result<LifePlanningChatResponse> chat(Authentication authentication,
                                                   @RequestBody LifePlanningChatRequest request) {
        requireStudent(authentication);
        return Result.ok(service.chat(SecurityUtil.currentUserId(authentication), request));
    }

    @Operation(summary = "读取我的长期记忆")
    @GetMapping("/memories")
    public Result<List<MemoryItemDTO>> memories(Authentication authentication) {
        requireStudent(authentication);
        return Result.ok(service.listMemories(SecurityUtil.currentUserId(authentication)));
    }

    @Operation(summary = "删除我的一条长期记忆")
    @DeleteMapping("/memories/{memoryId}")
    public Result<Void> deleteMemory(Authentication authentication, @PathVariable String memoryId) {
        requireStudent(authentication);
        service.deleteMemory(SecurityUtil.currentUserId(authentication), memoryId);
        return Result.ok();
    }

    private void requireStudent(Authentication authentication) {
        if (!"student".equals(SecurityUtil.currentRole(authentication))) {
            throw new com.zjnu.academic.common.exception.BusinessException(403, "仅学生可使用人生规划智能体");
        }
    }
}
