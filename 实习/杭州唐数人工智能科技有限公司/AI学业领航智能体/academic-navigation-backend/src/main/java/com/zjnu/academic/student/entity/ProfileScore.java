package com.zjnu.academic.student.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableField;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.extension.handlers.JacksonTypeHandler;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.Comparator;
import java.util.List;

@Data
@TableName(value = "t_profile_score", autoResultMap = true)
public class ProfileScore {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long studentId;

    private String term;

    private String dimensionKey;

    private String label;

    private Integer score;

    private Integer avgScore;

    private Integer maxScore;

    private String description;

    @TableField(typeHandler = JacksonTypeHandler.class)
    private List<String> details;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;

    // Canonical display order for the five-dimension radar chart
    public static final List<String> DISPLAY_ORDER = List.of(
            "academic", "practice", "quality", "culture", "health");

    public static Comparator<ProfileScore> byDisplayOrder() {
        return Comparator.comparingInt(score -> {
            int index = DISPLAY_ORDER.indexOf(score.getDimensionKey());
            return index < 0 ? DISPLAY_ORDER.size() : index;
        });
    }
}
