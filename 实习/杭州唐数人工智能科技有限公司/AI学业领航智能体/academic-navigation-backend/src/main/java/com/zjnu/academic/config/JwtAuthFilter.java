package com.zjnu.academic.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.zjnu.academic.auth.service.TokenBlacklistService;
import com.zjnu.academic.common.result.Result;
import com.zjnu.academic.common.util.JwtUtil;
import io.jsonwebtoken.Claims;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.List;

@Component
public class JwtAuthFilter extends OncePerRequestFilter {

    private static final Logger log = LoggerFactory.getLogger(JwtAuthFilter.class);

    private final JwtUtil jwtUtil;
    private final TokenBlacklistService tokenBlacklistService;
    private final ObjectMapper objectMapper;

    @Autowired
    public JwtAuthFilter(JwtUtil jwtUtil,
                         TokenBlacklistService tokenBlacklistService,
                         ObjectMapper objectMapper) {
        this.jwtUtil = jwtUtil;
        this.tokenBlacklistService = tokenBlacklistService;
        this.objectMapper = objectMapper;
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain chain) throws ServletException, IOException {
        // 登录与 CAS 回调不需要 Token，其余 /api/auth/**（如 /me、/logout）仍需校验
        String path = request.getRequestURI();
        if (path.equals("/api/auth/login") || path.startsWith("/api/auth/cas")) {
            chain.doFilter(request, response);
            return;
        }

        String token = extractToken(request);

        if (!StringUtils.hasText(token)) {
            chain.doFilter(request, response);
            return;
        }

        // 检查黑名单（已登出的 Token）
        if (tokenBlacklistService.contains(token)) {
            writeUnauthorized(response, "Token 已失效，请重新登录");
            return;
        }

        Claims claims = jwtUtil.parse(token);
        if (claims == null) {
            writeUnauthorized(response, "Token 无效或已过期");
            return;
        }

        String role = claims.get("role", String.class);
        String account = claims.get("account", String.class);
        Long userId = Long.parseLong(claims.getSubject());

        var auth = new UsernamePasswordAuthenticationToken(
                userId,
                account,
                List.of(new SimpleGrantedAuthority("ROLE_" + role.toUpperCase()))
        );
        SecurityContextHolder.getContext().setAuthentication(auth);

        chain.doFilter(request, response);
    }

    private String extractToken(HttpServletRequest request) {
        String bearer = request.getHeader("Authorization");
        if (StringUtils.hasText(bearer) && bearer.startsWith("Bearer ")) {
            return bearer.substring(7);
        }
        return null;
    }

    private void writeUnauthorized(HttpServletResponse response, String message) throws IOException {
        response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
        response.setContentType(MediaType.APPLICATION_JSON_VALUE);
        response.setCharacterEncoding(StandardCharsets.UTF_8.name());
        response.getWriter().write(
                objectMapper.writeValueAsString(Result.unauthorized(message))
        );
    }
}
