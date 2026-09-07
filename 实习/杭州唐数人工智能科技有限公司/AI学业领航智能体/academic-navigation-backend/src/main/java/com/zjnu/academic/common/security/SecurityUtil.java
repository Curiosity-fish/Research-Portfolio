package com.zjnu.academic.common.security;

import com.zjnu.academic.common.exception.BusinessException;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.GrantedAuthority;

public final class SecurityUtil {

    private SecurityUtil() {
    }

    public static Long currentUserId(Authentication authentication) {
        if (authentication == null || !(authentication.getPrincipal() instanceof Long userId)) {
            throw new BusinessException(401, "请先登录");
        }
        return userId;
    }

    public static String currentRole(Authentication authentication) {
        if (authentication == null || authentication.getAuthorities().isEmpty()) {
            throw new BusinessException(401, "请先登录");
        }
        GrantedAuthority authority = authentication.getAuthorities().iterator().next();
        String role = authority.getAuthority();
        if (role.startsWith("ROLE_")) {
            role = role.substring(5);
        }
        return role.toLowerCase();
    }
}
