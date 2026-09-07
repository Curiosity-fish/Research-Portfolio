package com.zjnu.academic.common.result;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

class ResultTest {

    @Test
    void okResultCarriesTimestamp() {
        Result<String> result = Result.ok("hello");
        assertEquals(200, result.getCode());
        assertEquals("hello", result.getData());
        assertTrue(result.getTimestamp() > 0);
    }
}
