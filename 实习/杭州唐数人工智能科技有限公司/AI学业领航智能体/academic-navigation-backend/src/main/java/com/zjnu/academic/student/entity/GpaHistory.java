package com.zjnu.academic.student.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("t_gpa_history")
public class GpaHistory {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long studentId;

    private String term;

    private BigDecimal gpa;

    private BigDecimal avgGpa;

    @TableField("`rank`")
    private Integer rank;

    private Integer totalStudents;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
