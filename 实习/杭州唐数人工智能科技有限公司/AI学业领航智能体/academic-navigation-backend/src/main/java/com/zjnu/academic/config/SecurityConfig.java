package com.zjnu.academic.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.zjnu.academic.common.result.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.MediaType;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

import java.nio.charset.StandardCharsets;

@Configuration
@EnableWebSecurity
public class SecurityConfig {

    private final JwtAuthFilter jwtAuthFilter;
    private final ObjectMapper objectMapper;

    @Autowired
    public SecurityConfig(JwtAuthFilter jwtAuthFilter, ObjectMapper objectMapper) {
        this.jwtAuthFilter = jwtAuthFilter;
        this.objectMapper = objectMapper;
    }

    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        http
            .csrf(AbstractHttpConfigurer::disable)
            .sessionManagement(s -> s.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            .authorizeHttpRequests(auth -> auth
                // 公开接口
                .requestMatchers(
                    "/api/auth/login",
                    "/api/eagent/health",
                    "/swagger-ui/**",
                    "/swagger-ui.html",
                    "/api-docs/**",
                    "/actuator/health"
                ).permitAll()
                // 登出与当前用户信息需要有效 Token
                .requestMatchers("/api/auth/me", "/api/auth/logout").authenticated()
                // 五端角色级 URL 鉴权
                .requestMatchers("/api/student/**").hasRole("STUDENT")
                .requestMatchers("/api/teacher/**").hasAnyRole("TEACHER", "COURSE_TEACHER")
                .requestMatchers("/api/course-teacher/**").hasRole("COURSE_TEACHER")
                .requestMatchers("/api/department/**").hasRole("DEPARTMENT")
                .requestMatchers("/api/dean/**").hasRole("DEAN")
                .requestMatchers("/api/school-sync/**").hasRole("DEAN")
                .requestMatchers("/api/alerts/**", "/api/agents/**", "/api/development/**").authenticated()
                // 其余接口需要登录
                .anyRequest().authenticated()
            )
            // 未认证时返回 JSON 而非重定向
            .exceptionHandling(e -> e
                .authenticationEntryPoint((req, res, ex) -> {
                    res.setStatus(401);
                    res.setContentType(MediaType.APPLICATION_JSON_VALUE);
                    res.setCharacterEncoding(StandardCharsets.UTF_8.name());
                    res.getWriter().write(
                        objectMapper.writeValueAsString(Result.unauthorized("请先登录"))
                    );
                })
                .accessDeniedHandler((req, res, ex) -> {
                    res.setStatus(403);
                    res.setContentType(MediaType.APPLICATION_JSON_VALUE);
                    res.setCharacterEncoding(StandardCharsets.UTF_8.name());
                    res.getWriter().write(
                        objectMapper.writeValueAsString(Result.forbidden("无权限访问"))
                    );
                })
            )
            .addFilterBefore(jwtAuthFilter, UsernamePasswordAuthenticationFilter.class);

        return http.build();
    }

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder(12);
    }
}
