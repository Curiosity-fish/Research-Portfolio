package com.zjnu.academic.teacher.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("t_teacher")
public class Teacher {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String teacherId;

    private Long userId;

    private Long departmentId;

    private String title;

    private Integer isClassAdvisor;

    private Long classId;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
