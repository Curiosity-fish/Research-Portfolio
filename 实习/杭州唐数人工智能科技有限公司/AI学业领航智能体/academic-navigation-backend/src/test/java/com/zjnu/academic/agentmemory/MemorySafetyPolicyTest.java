package com.zjnu.academic.agentmemory;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class MemorySafetyPolicyTest {

    @Test
    void acceptsConfirmedLowSensitivityPlanningPreference() {
        assertTrue(MemorySafetyPolicy.canStore("我计划毕业后优先在杭州找教育科技方向的工作"));
    }

    @Test
    void rejectsSensitiveOrOversizedContent() {
        assertFalse(MemorySafetyPolicy.canStore("我的 GPA 是 3.8，请记住"));
        assertFalse(MemorySafetyPolicy.canStore("我最近有心理问题，请记住"));
        assertFalse(MemorySafetyPolicy.canStore("我这门课得了 90 分，请记住"));
        assertFalse(MemorySafetyPolicy.canStore("x".repeat(361)));
    }
}
