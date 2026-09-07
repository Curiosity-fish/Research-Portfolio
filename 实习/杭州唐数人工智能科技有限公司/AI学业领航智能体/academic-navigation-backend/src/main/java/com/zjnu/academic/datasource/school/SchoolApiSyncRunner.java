package com.zjnu.academic.datasource.school;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

@Component
public class SchoolApiSyncRunner {

    private static final Logger log = LoggerFactory.getLogger(SchoolApiSyncRunner.class);

    private final SchoolApiSyncService syncService;
    private final String provider;

    public SchoolApiSyncRunner(SchoolApiSyncService syncService,
                               @Value("${academic.datasource.provider:mock}") String provider) {
        this.syncService = syncService;
        this.provider = provider;
    }

    @EventListener(ApplicationReadyEvent.class)
    public void onApplicationReady() {
        if ("school".equalsIgnoreCase(provider)) {
            log.info("academic.datasource.provider=school, starting school API sync...");
            try {
                syncService.syncAll();
            } catch (Exception e) {
                log.error("School API sync failed at startup", e);
            }
        }
    }
}
