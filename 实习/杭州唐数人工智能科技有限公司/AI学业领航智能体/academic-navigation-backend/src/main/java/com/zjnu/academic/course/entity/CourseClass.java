package com.zjnu.academic.course.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("t_course_class")
public class CourseClass {

    @TableId(type = IdType.AUTO)
    private Long id;

    private Long courseId;

    private String term;

    private String className;

    private Long teacherId;

    private Integer maxStudents;

    private Integer enrolledStudents;

    private String status;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}
