package com.zjnu.academic.agentmemory;

import java.time.Instant;

public record MemoryItemDTO(String id, String memory, Double score, Instant createdAt) {
}
