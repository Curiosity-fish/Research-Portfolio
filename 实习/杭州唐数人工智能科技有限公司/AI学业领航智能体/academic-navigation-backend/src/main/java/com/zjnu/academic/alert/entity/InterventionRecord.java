package com.zjnu.academic.alert.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.extension.handlers.JacksonTypeHandler;
import lombok.Data;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;

@Data
@TableName(value = "t_intervention_record", autoResultMap = true)
public class InterventionRecord {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long alertId;

    private Long teacherId;

    private LocalDate interventionDate;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> methods;

    private String content;

    private String studentResponse;

    private String followUpPlan;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
