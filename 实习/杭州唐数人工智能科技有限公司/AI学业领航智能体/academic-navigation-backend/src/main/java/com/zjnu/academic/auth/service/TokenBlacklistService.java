package com.zjnu.academic.auth.service;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.concurrent.TimeUnit;

@Service
public class TokenBlacklistService {

    private static final Logger log = LoggerFactory.getLogger(TokenBlacklistService.class);
    private static final String BLACKLIST_PREFIX = "jwt:blacklist:";

    private final StringRedisTemplate redisTemplate;

    @Autowired
    public TokenBlacklistService(StringRedisTemplate redisTemplate) {
        this.redisTemplate = redisTemplate;
    }

    public void add(String token, long ttlSeconds) {
        try {
            if (ttlSeconds > 0) {
                redisTemplate.opsForValue().set(BLACKLIST_PREFIX + token, "1", ttlSeconds, TimeUnit.SECONDS);
            }
        } catch (Exception e) {
            // Redis 不可用时，登出黑名单功能降级，仅记录警告
            log.warn("Redis 不可用，Token 黑名单写入跳过: {}", e.getMessage());
        }
    }

    public boolean contains(String token) {
        try {
            return Boolean.TRUE.equals(redisTemplate.hasKey(BLACKLIST_PREFIX + token));
        } catch (Exception e) {
            log.warn("Redis 不可用，黑名单检查跳过: {}", e.getMessage());
            return false;
        }
    }
}

