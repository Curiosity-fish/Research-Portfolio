package com.zjnu.academic.rawdata.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("t_attendance")
public class Attendance {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long studentId;

    private String term;

    private Integer totalClasses;

    private Integer absentCount;

    private Integer lateCount;

    private String note;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
