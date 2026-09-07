package com.zjnu.academic.agentmemory;

import java.util.List;

public record LifePlanningChatRequest(String message, Boolean remember, List<ChatHistoryItem> history) {
    public record ChatHistoryItem(String role, String content) {
    }
}
