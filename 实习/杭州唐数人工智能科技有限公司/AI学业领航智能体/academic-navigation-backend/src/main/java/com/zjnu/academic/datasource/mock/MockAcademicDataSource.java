package com.zjnu.academic.datasource.mock;

import com.zjnu.academic.datasource.AcademicDataSource;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.context.annotation.Primary;
import org.springframework.stereotype.Component;

/**
 * Mock 数据源实现
 * 在校方接口就绪前使用此实现，返回稳定的本地演示数据。
 * 切换方式：实现 SchoolApiDataSource 后，将 @Primary 移到新实现上即可。
 */
@Primary
@Component
@ConditionalOnProperty(name = "academic.datasource.provider", havingValue = "mock", matchIfMissing = true)
public class MockAcademicDataSource implements AcademicDataSource {

    @Override
    public double getGpa(String studentAccount, String semester) {
        return 3.62;
    }

    @Override
    public int[] getRank(String studentAccount, String semester) {
        return new int[]{15, 120};
    }

    @Override
    public int[] getCredits(String studentAccount) {
        return new int[]{89, 176};
    }
}
