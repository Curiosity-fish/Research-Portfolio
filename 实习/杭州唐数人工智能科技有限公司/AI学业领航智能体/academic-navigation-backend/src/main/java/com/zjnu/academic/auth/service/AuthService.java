package com.zjnu.academic.auth.service;

import com.zjnu.academic.auth.dto.LoginRequest;
import com.zjnu.academic.auth.dto.LoginResponse;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.util.JwtUtil;
import com.zjnu.academic.common.util.RoleDetector;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.service.UserService;
import io.jsonwebtoken.Claims;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@Service
public class AuthService {

    private static final Logger log = LoggerFactory.getLogger(AuthService.class);

    private final UserService userService;
    private final JwtUtil jwtUtil;
    private final PasswordEncoder passwordEncoder;
    private final TokenBlacklistService tokenBlacklistService;

    @Autowired
    public AuthService(UserService userService,
                       JwtUtil jwtUtil,
                       PasswordEncoder passwordEncoder,
                       TokenBlacklistService tokenBlacklistService) {
        this.userService = userService;
        this.jwtUtil = jwtUtil;
        this.passwordEncoder = passwordEncoder;
        this.tokenBlacklistService = tokenBlacklistService;
    }

    public LoginResponse login(LoginRequest req) {
        if (req.getAccount() == null || req.getAccount().isBlank()) {
            throw new BusinessException("请输入工号/学号");
        }
        if (req.getPassword() == null || req.getPassword().isBlank()) {
            throw new BusinessException("请输入密码");
        }

        RoleDetector.detect(req.getAccount());

        User user = userService.findByAccount(req.getAccount());
        if (user == null || !passwordEncoder.matches(req.getPassword(), user.getPassword())) {
            throw new BusinessException(401, "账号或密码错误");
        }

        String token = jwtUtil.generate(user.getId(), user.getAccount(), user.getRole());

        return LoginResponse.builder()
                .token(token)
                .user(toUserInfo(user))
                .build();
    }

    public LoginResponse.UserInfo me(Long userId) {
        User user = userService.findById(userId);
        if (user == null) {
            throw new BusinessException(401, "用户不存在或已停用");
        }
        return toUserInfo(user);
    }

    private LoginResponse.UserInfo toUserInfo(User user) {
        return LoginResponse.UserInfo.builder()
                .id(user.getId())
                .account(user.getAccount())
                .name(user.getName())
                .role(user.getRole())
                .college(user.getCollege())
                .major(user.getMajor())
                .className(user.getClassName())
                .grade(user.getGrade())
                .build();
    }

    public void logout(String token) {
        Claims claims = jwtUtil.parse(token);
        if (claims == null) return;
        long ttl = jwtUtil.getRemainingSeconds(claims);
        tokenBlacklistService.add(token, ttl);
    }

    public boolean isBlacklisted(String token) {
        return tokenBlacklistService.contains(token);
    }
}
