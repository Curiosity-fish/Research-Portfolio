package com.zjnu.academic.agentmemory;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "academic.agent-memory")
public class AgentMemoryProperties {

    private final Mem0 mem0 = new Mem0();
    private final Eagent eagent = new Eagent();

    public Mem0 getMem0() {
        return mem0;
    }

    public Eagent getEagent() {
        return eagent;
    }

    public static class Mem0 {
        private boolean enabled;
        private String baseUrl = "http://localhost:8888";
        private String apiKey = "";
        private String identitySecret = "";
        private int timeoutSeconds = 8;

        public boolean isEnabled() { return enabled; }
        public void setEnabled(boolean enabled) { this.enabled = enabled; }
        public String getBaseUrl() { return baseUrl; }
        public void setBaseUrl(String baseUrl) { this.baseUrl = baseUrl; }
        public String getApiKey() { return apiKey; }
        public void setApiKey(String apiKey) { this.apiKey = apiKey; }
        public String getIdentitySecret() { return identitySecret; }
        public void setIdentitySecret(String identitySecret) { this.identitySecret = identitySecret; }
        public int getTimeoutSeconds() { return timeoutSeconds; }
        public void setTimeoutSeconds(int timeoutSeconds) { this.timeoutSeconds = timeoutSeconds; }
    }

    public static class Eagent {
        private String baseUrl = "";
        private String workflowId = "";
        private String apiKey = "";
        private int timeoutSeconds = 45;

        public String getBaseUrl() { return baseUrl; }
        public void setBaseUrl(String baseUrl) { this.baseUrl = baseUrl; }
        public String getWorkflowId() { return workflowId; }
        public void setWorkflowId(String workflowId) { this.workflowId = workflowId; }
        public String getApiKey() { return apiKey; }
        public void setApiKey(String apiKey) { this.apiKey = apiKey; }
        public int getTimeoutSeconds() { return timeoutSeconds; }
        public void setTimeoutSeconds(int timeoutSeconds) { this.timeoutSeconds = timeoutSeconds; }
    }
}
