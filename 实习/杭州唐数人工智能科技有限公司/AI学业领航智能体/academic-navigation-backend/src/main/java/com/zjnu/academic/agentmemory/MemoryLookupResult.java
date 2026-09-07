package com.zjnu.academic.agentmemory;

import java.util.List;

public record MemoryLookupResult(List<MemoryItemDTO> memories, boolean available) {
}
