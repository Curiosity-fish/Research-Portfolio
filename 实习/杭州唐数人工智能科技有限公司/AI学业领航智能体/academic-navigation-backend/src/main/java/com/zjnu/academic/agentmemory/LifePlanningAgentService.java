package com.zjnu.academic.agentmemory;

import com.zjnu.academic.common.exception.BusinessException;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class LifePlanningAgentService {

    private final Mem0MemoryService memoryService;
    private final EagentLifePlanningService eagentService;
    private final StudentPlanningContextService planningContextService;

    public LifePlanningAgentService(Mem0MemoryService memoryService, EagentLifePlanningService eagentService,
                                    StudentPlanningContextService planningContextService) {
        this.memoryService = memoryService;
        this.eagentService = eagentService;
        this.planningContextService = planningContextService;
    }

    public LifePlanningChatResponse chat(Long userId, LifePlanningChatRequest request) {
        if (request == null || request.message() == null || request.message().isBlank()) {
            throw new BusinessException(400, "请输入问题");
        }
        String message = request.message().trim();
        if (message.length() > 2000) {
            throw new BusinessException(400, "单次消息不能超过 2000 个字符");
        }
        MemoryLookupResult lookup = memoryService.search(userId, message);
        String currentContext = planningContextService.currentContext(userId);
        String answer = eagentService.chat(userId, message, request.history(), lookup.memories(), currentContext);
        boolean requestedMemory = Boolean.TRUE.equals(request.remember());
        boolean eligibleForMemory = MemorySafetyPolicy.canStore(message);
        boolean remembered = requestedMemory && eligibleForMemory && memoryService.remember(userId, message);
        String memoryNotice = null;
        if (requestedMemory && !eligibleForMemory) {
            memoryNotice = "本次内容包含敏感信息或不适合作为长期记忆，未保存。";
        } else if (requestedMemory && (!lookup.available() || !remembered)) {
            memoryNotice = "长期记忆服务暂不可用，本次内容未保存。";
        } else if (remembered) {
            memoryNotice = "已保存为长期记忆。";
        }
        return new LifePlanningChatResponse(answer, lookup.available(), remembered, memoryNotice);
    }

    public List<MemoryItemDTO> listMemories(Long userId) {
        return memoryService.list(userId);
    }

    public void deleteMemory(Long userId, String memoryId) {
        if (memoryId == null || memoryId.isBlank()) {
            throw new BusinessException(400, "记忆标识不能为空");
        }
        memoryService.delete(userId, memoryId);
    }
}
