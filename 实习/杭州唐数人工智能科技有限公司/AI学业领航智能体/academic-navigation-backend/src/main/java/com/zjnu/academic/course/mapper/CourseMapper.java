package com.zjnu.academic.course.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.zjnu.academic.course.entity.Course;
import org.apache.ibatis.annotations.Mapper;

@Mapper
public interface CourseMapper extends BaseMapper<Course> {
}
