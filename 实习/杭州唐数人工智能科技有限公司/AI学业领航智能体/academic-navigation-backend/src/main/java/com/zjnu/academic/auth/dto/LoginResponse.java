package com.zjnu.academic.auth.dto;

public class LoginResponse {
    private String token;
    private UserInfo user;

    public LoginResponse() {}

    private LoginResponse(Builder builder) {
        this.token = builder.token;
        this.user = builder.user;
    }

    public static Builder builder() {
        return new Builder();
    }

    public String getToken() {
        return token;
    }

    public void setToken(String token) {
        this.token = token;
    }

    public UserInfo getUser() {
        return user;
    }

    public void setUser(UserInfo user) {
        this.user = user;
    }

    @Override
    public String toString() {
        return "LoginResponse{token='" + token + "', user=" + user + "}";
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        LoginResponse that = (LoginResponse) o;
        return java.util.Objects.equals(token, that.token) &&
                java.util.Objects.equals(user, that.user);
    }

    @Override
    public int hashCode() {
        return java.util.Objects.hash(token, user);
    }

    public static class Builder {
        private String token;
        private UserInfo user;

        public Builder token(String token) {
            this.token = token;
            return this;
        }

        public Builder user(UserInfo user) {
            this.user = user;
            return this;
        }

        public LoginResponse build() {
            return new LoginResponse(this);
        }
    }

    public static class UserInfo {
        private Long id;
        private String account;
        private String name;
        private String role;
        private String college;
        private String major;
        private String className;
        private String grade;

        public UserInfo() {}

        private UserInfo(Builder builder) {
            this.id = builder.id;
            this.account = builder.account;
            this.name = builder.name;
            this.role = builder.role;
            this.college = builder.college;
            this.major = builder.major;
            this.className = builder.className;
            this.grade = builder.grade;
        }

        public static Builder builder() {
            return new Builder();
        }

        public Long getId() { return id; }
        public void setId(Long id) { this.id = id; }

        public String getAccount() { return account; }
        public void setAccount(String account) { this.account = account; }

        public String getName() { return name; }
        public void setName(String name) { this.name = name; }

        public String getRole() { return role; }
        public void setRole(String role) { this.role = role; }

        public String getCollege() { return college; }
        public void setCollege(String college) { this.college = college; }

        public String getMajor() { return major; }
        public void setMajor(String major) { this.major = major; }

        public String getClassName() { return className; }
        public void setClassName(String className) { this.className = className; }

        public String getGrade() { return grade; }
        public void setGrade(String grade) { this.grade = grade; }

        @Override
        public String toString() {
            return "UserInfo{id=" + id + ", account='" + account + "', name='" + name +
                    "', role='" + role + "'}";
        }

        @Override
        public boolean equals(Object o) {
            if (this == o) return true;
            if (o == null || getClass() != o.getClass()) return false;
            UserInfo that = (UserInfo) o;
            return java.util.Objects.equals(id, that.id) &&
                    java.util.Objects.equals(account, that.account) &&
                    java.util.Objects.equals(name, that.name) &&
                    java.util.Objects.equals(role, that.role) &&
                    java.util.Objects.equals(college, that.college) &&
                    java.util.Objects.equals(major, that.major) &&
                    java.util.Objects.equals(className, that.className) &&
                    java.util.Objects.equals(grade, that.grade);
        }

        @Override
        public int hashCode() {
            return java.util.Objects.hash(id, account, name, role, college, major, className, grade);
        }

        public static class Builder {
            private Long id;
            private String account;
            private String name;
            private String role;
            private String college;
            private String major;
            private String className;
            private String grade;

            public Builder id(Long id) { this.id = id; return this; }
            public Builder account(String account) { this.account = account; return this; }
            public Builder name(String name) { this.name = name; return this; }
            public Builder role(String role) { this.role = role; return this; }
            public Builder college(String college) { this.college = college; return this; }
            public Builder major(String major) { this.major = major; return this; }
            public Builder className(String className) { this.className = className; return this; }
            public Builder grade(String grade) { this.grade = grade; return this; }

            public UserInfo build() {
                return new UserInfo(this);
            }
        }
    }
}
