package com.zjnu.academic.department.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.zjnu.academic.department.entity.Department;
import org.apache.ibatis.annotations.Mapper;

@Mapper
public interface DepartmentMapper extends BaseMapper<Department> {
}
