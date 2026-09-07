package com.zjnu.academic.health.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.extension.handlers.JacksonTypeHandler;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;

@Data
@TableName(value = "t_health_report", autoResultMap = true)
public class HealthReport {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long studentId;

    private String term;

    private Integer totalScore;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<Map<String, Object>> items;

    private String reportUrl;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
