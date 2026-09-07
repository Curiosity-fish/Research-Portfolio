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
@TableName(value = "t_job_cache", autoResultMap = true)
public class JobCache {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String platform;

    private String platformLabel;

    private String companyName;

    private String companySize;

    private String companyStage;

    private String jobTitle;

    private String salaryRange;

    private String city;

    private String district;

    private String education;

    private String experience;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> tags;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> highlights;

    private Integer matchScore;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> matchReasons;

    private String publishDate;

    private String sourceUrl;

    private String category;

    private LocalDateTime publishedAt;

    private LocalDateTime crawledAt;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
