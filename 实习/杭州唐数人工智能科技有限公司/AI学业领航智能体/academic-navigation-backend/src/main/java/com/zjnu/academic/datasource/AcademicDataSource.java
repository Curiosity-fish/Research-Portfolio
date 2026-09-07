package com.zjnu.academic.datasource;

/**
 * 校方学业数据来源接口
 *
 * 业务层只依赖此接口，不直接访问数据库或外部 API。
 * 当前由 MockAcademicDataSource 实现；
 * 等校方提供接口后，实现 SchoolApiDataSource 并切换配置即可，业务代码不动。
 */
public interface AcademicDataSource {

    /**
     * 获取学生某学期的 GPA
     * @param studentAccount 学号
     * @param semester       学期，格式 2024-2025-1
     */
    double getGpa(String studentAccount, String semester);

    /**
     * 获取学生在专业内的排名
     * @return [rank, total]，如 [15, 120]
     */
    int[] getRank(String studentAccount, String semester);

    /**
     * 获取学生已修学分和总要求学分
     * @return [earned, required]
     */
    int[] getCredits(String studentAccount);
}
