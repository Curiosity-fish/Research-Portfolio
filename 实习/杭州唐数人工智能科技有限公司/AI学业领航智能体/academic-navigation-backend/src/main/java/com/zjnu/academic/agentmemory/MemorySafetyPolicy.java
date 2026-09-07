package com.zjnu.academic.agentmemory;

import java.util.List;

final class MemorySafetyPolicy {

    private static final List<String> SENSITIVE_MARKERS = List.of(
            "身份证", "手机号", "手机号码", "电话", "住址", "密码", "token", "验证码", "学号",
            "gpa", "绩点", "成绩", "排名", "预警", "心理", "抑郁", "自杀", "自残", "病历",
            "健康", "体测", "考勤", "家庭住址", "银行卡");

    private static final List<String> PLANNING_MARKERS = List.of(
            "目标", "计划", "规划", "打算", "准备", "希望", "想要", "倾向", "偏好", "喜欢",
            "不喜欢", "优先", "愿意", "不愿意", "安排", "地点", "城市", "预算", "时间", "方向",
            "考研", "留学", "就业", "考公", "考编", "实习", "证书", "复盘", "行动");

    private MemorySafetyPolicy() {
    }

    static boolean canStore(String text) {
        if (text == null || text.isBlank() || text.length() > 360) {
            return false;
        }
        String normalized = text.toLowerCase();
        return SENSITIVE_MARKERS.stream().noneMatch(normalized::contains)
                && PLANNING_MARKERS.stream().anyMatch(normalized::contains);
    }
}
