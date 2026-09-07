package com.zjnu.academic.course.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("t_course")
public class Course {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String code;

    private String name;

    private BigDecimal credits;

    private String type;

    private Long departmentId;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
