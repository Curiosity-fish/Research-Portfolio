package com.zjnu.academic.common.exception;

import com.zjnu.academic.common.result.Result;
import org.junit.jupiter.api.Test;
import org.springframework.http.ResponseEntity;

import static org.junit.jupiter.api.Assertions.assertEquals;

class GlobalExceptionHandlerTest {

    private final GlobalExceptionHandler handler = new GlobalExceptionHandler();

    @Test
    void businessExceptionMapsCodeToHttpStatus() {
        ResponseEntity<Result<Void>> response =
                handler.handleBusinessException(new BusinessException(401, "未登录"));
        assertEquals(401, response.getStatusCode().value());
        assertEquals(401, response.getBody().getCode());
        assertEquals("未登录", response.getBody().getMessage());
    }

    @Test
    void invalidCodeFallsBackToBadRequest() {
        ResponseEntity<Result<Void>> response =
                handler.handleBusinessException(new BusinessException(600, "自定义错误"));
        assertEquals(400, response.getStatusCode().value());
        assertEquals(600, response.getBody().getCode());
    }
}
