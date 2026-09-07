package com.zjnu.academic.course.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@TableName("t_grade")
public class Grade {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long studentId;

    private Long courseClassId;

    private BigDecimal score;

    private BigDecimal gradePoint;

    private String status;

    private String term;

    private LocalDate examDate;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
