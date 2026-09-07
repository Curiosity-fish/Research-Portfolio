package com.zjnu.academic.development.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.extension.handlers.JacksonTypeHandler;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

@Data
@TableName(value = "t_school_cache", autoResultMap = true)
public class SchoolCache {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String type;

    private String name;

    private String nameZh;

    private String country;

    private String countryEmoji;

    private String location;

    @TableField("`rank`")
    private String rank;

    private String tier;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> programs;

    private String admissionGpa;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> examRequirements;

    private String languageRequirement;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> highlights;

    private Integer matchScore;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> matchReasons;

    private String officialUrl;

    private String deadline;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
