package com.zjnu.academic.department.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("t_department")
public class Department {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String name;

    private String code;

    private Long deanId;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
