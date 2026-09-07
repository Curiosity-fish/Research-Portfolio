package com.zjnu.academic.common.util;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.JwtException;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.util.Date;
import java.util.Map;

@Component
public class JwtUtil {

    private static final Logger log = LoggerFactory.getLogger(JwtUtil.class);

    @Value("${jwt.secret}")
    private String secret;

    @Value("${jwt.expiration}")
    private long expiration; // 秒

    private SecretKey getKey() {
        return Keys.hmacShaKeyFor(secret.getBytes(StandardCharsets.UTF_8));
    }

    /**
     * 生成 Token
     */
    public String generate(Long userId, String account, String role) {
        Date now = new Date();
        Date exp = new Date(now.getTime() + expiration * 1000);

        return Jwts.builder()
                .subject(String.valueOf(userId))
                .claims(Map.of(
                        "account", account,
                        "role", role
                ))
                .issuedAt(now)
                .expiration(exp)
                .signWith(getKey())
                .compact();
    }

    /**
     * 解析 Token，失败返回 null
     */
    public Claims parse(String token) {
        try {
            return Jwts.parser()
                    .verifyWith(getKey())
                    .build()
                    .parseSignedClaims(token)
                    .getPayload();
        } catch (JwtException | IllegalArgumentException e) {
            log.debug("JWT 解析失败: {}", e.getMessage());
            return null;
        }
    }

    /**
     * 获取 Token 剩余有效秒数（用于登出时设置 Redis TTL）
     */
    public long getRemainingSeconds(Claims claims) {
        long exp = claims.getExpiration().getTime();
        long now = System.currentTimeMillis();
        return Math.max(0, (exp - now) / 1000);
    }
}
