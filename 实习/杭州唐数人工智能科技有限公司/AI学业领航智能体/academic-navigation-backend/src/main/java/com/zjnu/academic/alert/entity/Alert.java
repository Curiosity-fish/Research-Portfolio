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
@TableName(value = "t_alert", autoResultMap = true)
public class Alert {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long studentId;

    private String level;

    private String type;

    private String title;

    private String description;

    private String course;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> failedCourses;

    private LocalDate triggerDate;

    private String status;

    private String suggestion;

    private String triggerEvent;

    private LocalDateTime pushedAt;

    private Long handlerId;

    private String handleNote;

    private LocalDateTime studentConfirmedAt;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
