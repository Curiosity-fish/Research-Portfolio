package com.zjnu.academic.common.util;

import com.zjnu.academic.common.exception.BusinessException;

public class RoleDetector {

    private RoleDetector() {}

    /**
     * 根据工号/学号前缀识别角色
     * 规则：L→院领导 D→系主任 T→班主任 C→任课老师 纯数字→学生
     */
    public static String detect(String account) {
        if (account == null || account.isBlank()) {
            throw new BusinessException("工号/学号不能为空");
        }
        if (account.startsWith("L")) return "dean";
        if (account.startsWith("D")) return "department";
        if (account.startsWith("T")) return "teacher";
        if (account.startsWith("C")) return "course_teacher";
        if (account.matches("\\d+"))  return "student";
        throw new BusinessException("工号格式不正确，无法识别身份");
    }
}
