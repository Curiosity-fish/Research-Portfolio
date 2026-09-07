package com.zjnu.academic.department.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.extension.handlers.JacksonTypeHandler;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

@Data
@TableName(value = "t_major_plan", autoResultMap = true)
public class MajorPlan {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long majorId;

    private String grade;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> dimensions;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<Integer> standard;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<Integer> actual;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> tips;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
