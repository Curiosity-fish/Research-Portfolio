package com.zjnu.academic.auth.service;

import com.zjnu.academic.auth.dto.LoginResponse;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.util.JwtUtil;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.service.UserService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.security.crypto.password.PasswordEncoder;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class AuthServiceTest {

    @Mock
    private UserService userService;

    @Mock
    private JwtUtil jwtUtil;

    @Mock
    private PasswordEncoder passwordEncoder;

    @Mock
    private TokenBlacklistService tokenBlacklistService;

    private AuthService authService;

    @BeforeEach
    void setUp() {
        authService = new AuthService(userService, jwtUtil, passwordEncoder, tokenBlacklistService);
    }

    @Test
    void meReturnsUserInfoForExistingActiveUser() {
        User user = new User();
        user.setId(1L);
        user.setAccount("2022001");
        user.setName("李明");
        user.setRole("student");
        user.setCollege("数学与计算机科学学院");
        user.setMajor("计算机科学与技术");
        user.setClassName("计科2201");
        user.setGrade("2022");
        when(userService.findById(1L)).thenReturn(user);

        LoginResponse.UserInfo info = authService.me(1L);

        assertEquals(1L, info.getId());
        assertEquals("2022001", info.getAccount());
        assertEquals("李明", info.getName());
        assertEquals("student", info.getRole());
        assertEquals("数学与计算机科学学院", info.getCollege());
        assertEquals("计算机科学与技术", info.getMajor());
        assertEquals("计科2201", info.getClassName());
        assertEquals("2022", info.getGrade());
    }

    @Test
    void meThrowsWhenUserMissing() {
        when(userService.findById(99L)).thenReturn(null);

        BusinessException ex = assertThrows(BusinessException.class, () -> authService.me(99L));
        assertEquals(401, ex.getCode());
    }
}
