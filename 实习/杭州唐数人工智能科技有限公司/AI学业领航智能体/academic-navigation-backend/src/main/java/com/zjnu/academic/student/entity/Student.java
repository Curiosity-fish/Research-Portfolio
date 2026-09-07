package com.zjnu.academic.student.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("t_student")
public class Student {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String studentId;

    private Long userId;

    private Long classId;

    private Long majorId;

    private Long departmentId;

    private BigDecimal gpa;

    @TableField("`rank`")
    private Integer rank;

    private Integer totalStudents;

    private String alertLevel;

    private BigDecimal totalCredits;

    private BigDecimal requiredCredits;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
