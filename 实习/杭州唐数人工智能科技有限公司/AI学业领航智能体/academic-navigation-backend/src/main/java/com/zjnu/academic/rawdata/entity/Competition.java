package com.zjnu.academic.rawdata.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("t_competition")
public class Competition {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long studentId;

    private String term;

    private String competitionName;

    private String level;

    private String award;

    private Integer points;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
