package com.zjnu.academic.auth.controller;

import com.zjnu.academic.auth.dto.LoginRequest;
import com.zjnu.academic.auth.dto.LoginResponse;
import com.zjnu.academic.auth.service.AuthService;
import com.zjnu.academic.common.result.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.core.Authentication;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@Tag(name = "认证", description = "登录、登出接口")
@RestController
@RequestMapping("/api/auth")
public class AuthController {

    private final AuthService authService;

    @Autowired
    public AuthController(AuthService authService) {
        this.authService = authService;
    }

    @Operation(summary = "登录", description = "工号/学号 + 密码，系统自动识别身份")
    @PostMapping("/login")
    public Result<LoginResponse> login(@RequestBody LoginRequest req) {
        return Result.ok(authService.login(req));
    }

    @Operation(summary = "当前用户", description = "根据 Token 返回当前登录用户信息，供前端刷新页面后恢复登录态")
    @GetMapping("/me")
    public Result<LoginResponse.UserInfo> me(Authentication authentication) {
        Long userId = (Long) authentication.getPrincipal();
        return Result.ok(authService.me(userId));
    }

    @Operation(summary = "登出", description = "将当前 Token 加入黑名单")
    @PostMapping("/logout")
    public Result<Void> logout(HttpServletRequest request) {
        String token = extractToken(request);
        if (StringUtils.hasText(token)) {
            authService.logout(token);
        }
        return Result.ok();
    }

    private String extractToken(HttpServletRequest request) {
        String bearer = request.getHeader("Authorization");
        if (StringUtils.hasText(bearer) && bearer.startsWith("Bearer ")) {
            return bearer.substring(7);
        }
        return null;
    }
}
