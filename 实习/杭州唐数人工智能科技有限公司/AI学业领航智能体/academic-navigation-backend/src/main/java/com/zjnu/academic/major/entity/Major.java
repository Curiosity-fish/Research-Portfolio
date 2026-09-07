package com.zjnu.academic.major.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("t_major")
public class Major {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String name;

    private String code;

    private Long departmentId;

    private String degreeType;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
