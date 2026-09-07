package com.zjnu.academic;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import com.zjnu.academic.config.SchoolApiProperties;
import com.zjnu.academic.config.LlmProperties;

@EnableConfigurationProperties({SchoolApiProperties.class, LlmProperties.class})
@SpringBootApplication
public class AcademicNavigationApplication {
    public static void main(String[] args) {
        SpringApplication.run(AcademicNavigationApplication.class, args);
    }
}
