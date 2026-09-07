package com.zjnu.academic.student.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("t_class")
public class StudentClass {

    @TableId(type = IdType.AUTO)
    private Long id;

    private String name;

    private String grade;

    private Long majorId;

    private Long advisorId;

    private Integer totalStudents;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
