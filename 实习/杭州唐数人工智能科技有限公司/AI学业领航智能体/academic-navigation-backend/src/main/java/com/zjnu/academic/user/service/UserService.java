package com.zjnu.academic.user.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class UserService {

    private final UserMapper userMapper;

    @Autowired
    public UserService(UserMapper userMapper) {
        this.userMapper = userMapper;
    }

    public User findByAccount(String account) {
        return userMapper.selectOne(
                new LambdaQueryWrapper<User>()
                        .eq(User::getAccount, account)
                        .eq(User::getIsActive, 1)
        );
    }

    public User findById(Long id) {
        return userMapper.selectOne(
                new LambdaQueryWrapper<User>()
                        .eq(User::getId, id)
                        .eq(User::getIsActive, 1)
        );
    }
}
