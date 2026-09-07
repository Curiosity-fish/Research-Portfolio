package com.zjnu.academic.rawdata.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("t_volunteer")
public class Volunteer {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long studentId;

    private String term;

    private BigDecimal hours;

    private String description;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
