package com.zjnu.academic.development.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zjnu.academic.common.exception.BusinessException;
import com.zjnu.academic.common.result.PageResult;
import com.zjnu.academic.development.dto.DevelopmentAnalyzeDTO;
import com.zjnu.academic.development.dto.DevelopmentAnalyzeRequest;
import com.zjnu.academic.development.dto.DevelopmentPathDTO;
import com.zjnu.academic.development.dto.JobItemDTO;
import com.zjnu.academic.development.dto.SchoolItemDTO;
import com.zjnu.academic.development.entity.JobCache;
import com.zjnu.academic.development.entity.SchoolCache;
import com.zjnu.academic.development.mapper.JobCacheMapper;
import com.zjnu.academic.development.mapper.SchoolCacheMapper;
import com.zjnu.academic.student.entity.GpaHistory;
import com.zjnu.academic.student.entity.ProfileScore;
import com.zjnu.academic.student.entity.Student;
import com.zjnu.academic.student.mapper.GpaHistoryMapper;
import com.zjnu.academic.student.mapper.ProfileScoreMapper;
import com.zjnu.academic.student.mapper.StudentMapper;
import com.zjnu.academic.user.entity.User;
import com.zjnu.academic.user.mapper.UserMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class DevelopmentService {

    private final StudentMapper studentMapper;
    private final UserMapper userMapper;
    private final ProfileScoreMapper profileScoreMapper;
    private final GpaHistoryMapper gpaHistoryMapper;
    private final JobCacheMapper jobCacheMapper;
    private final SchoolCacheMapper schoolCacheMapper;

    @Autowired
    public DevelopmentService(StudentMapper studentMapper,
                              UserMapper userMapper,
                              ProfileScoreMapper profileScoreMapper,
                              GpaHistoryMapper gpaHistoryMapper,
                              JobCacheMapper jobCacheMapper,
                              SchoolCacheMapper schoolCacheMapper) {
        this.studentMapper = studentMapper;
        this.userMapper = userMapper;
        this.profileScoreMapper = profileScoreMapper;
        this.gpaHistoryMapper = gpaHistoryMapper;
        this.jobCacheMapper = jobCacheMapper;
        this.schoolCacheMapper = schoolCacheMapper;
    }

    public List<DevelopmentPathDTO> paths(Long userId) {
        Student student = requireStudent(userId);
        double gpa = student.getGpa() == null ? 3.0 : student.getGpa().doubleValue();
        Map<String, Integer> scores = profileScores(student.getId());
        int academic = scores.getOrDefault("academic", 75);
        int practice = scores.getOrDefault("practice", 62);
        int quality = scores.getOrDefault("quality", 74);
        int health = scores.getOrDefault("health", 80);
        int culture = scores.getOrDefault("culture", 80);

        int graduate = clamp((int) Math.round(56 + academic * 0.28 + practice * 0.08 + culture * 0.04), 30, 95);
        int overseas = clamp((int) Math.round(52 + academic * 0.20 + practice * 0.14 + health * 0.06), 30, 95);
        int civil = clamp((int) Math.round(44 + culture * 0.24 + quality * 0.16 + academic * 0.10), 30, 95);
        int institution = clamp((int) Math.round(46 + culture * 0.20 + quality * 0.16 + academic * 0.14), 30, 95);
        int employment = clamp((int) Math.round(52 + practice * 0.26 + academic * 0.16 + quality * 0.06), 30, 95);

        return List.of(
                buildPath("graduate", "考研深造", graduate, gpa, academic, practice, culture, quality, health),
                buildPath("overseas", "出国留学", overseas, gpa, academic, practice, culture, quality, health),
                buildPath("civil", "国考公务员", civil, gpa, academic, practice, culture, quality, health),
                buildPath("institution", "事业单位考编", institution, gpa, academic, practice, culture, quality, health),
                buildPath("employment", "直接就业", employment, gpa, academic, practice, culture, quality, health));
    }

    public PageResult<JobItemDTO> jobs(String category, int page, int size) {
        LambdaQueryWrapper<JobCache> wrapper = new LambdaQueryWrapper<>();
        if (category != null && !category.isBlank()) {
            wrapper.eq(JobCache::getCategory, category);
        }
        wrapper.orderByDesc(JobCache::getMatchScore);
        List<JobItemDTO> dtos = jobCacheMapper.selectList(wrapper).stream()
                .map(this::toJobDto)
                .toList();
        int from = Math.min((page - 1) * size, dtos.size());
        int to = Math.min(from + size, dtos.size());
        return PageResult.of(dtos.subList(from, to), dtos.size(), page, size);
    }

    public PageResult<SchoolItemDTO> schools(String type, String tier, int page, int size) {
        if (type == null || type.isBlank() || !List.of("grad", "overseas").contains(type)) {
            throw new BusinessException(400, "院校类型必须为 grad 或 overseas");
        }
        LambdaQueryWrapper<SchoolCache> wrapper = new LambdaQueryWrapper<SchoolCache>()
                .eq(SchoolCache::getType, type);
        if (tier != null && !tier.isBlank()) {
            wrapper.eq(SchoolCache::getTier, tier);
        }
        wrapper.orderByDesc(SchoolCache::getMatchScore);
        List<SchoolItemDTO> dtos = schoolCacheMapper.selectList(wrapper).stream()
                .map(this::toSchoolDto)
                .toList();
        int from = Math.min((page - 1) * size, dtos.size());
        int to = Math.min(from + size, dtos.size());
        return PageResult.of(dtos.subList(from, to), dtos.size(), page, size);
    }

    public DevelopmentAnalyzeDTO analyze(Long userId, DevelopmentAnalyzeRequest request) {
        if (request == null || request.targetKey() == null || request.targetKey().isBlank()) {
            throw new BusinessException(400, "请选择发展目标");
        }
        DevelopmentPathDTO path = paths(userId).stream()
                .filter(p -> p.key().equals(request.targetKey()))
                .findFirst()
                .orElseThrow(() -> new BusinessException(400, "未知发展目标"));
        int matchScore = path.matchScore();
        if (request.targetScore() != null) {
            matchScore = clamp((matchScore + request.targetScore()) / 2, 30, 95);
        }
        List<String> suggestions = new ArrayList<>();
        path.gapItems().stream()
                .filter(gap -> "high".equals(gap.urgent()))
                .forEach(gap -> suggestions.add("优先补齐" + gap.dimension() + "：目标 " + gap.target()));
        if (matchScore < 60) {
            suggestions.add("当前匹配度较低，建议先补齐核心差距再确认目标");
        } else if (matchScore >= 85) {
            suggestions.add("匹配度较高，建议按里程碑持续推进并定期复盘");
        }
        return new DevelopmentAnalyzeDTO(
                path.key(),
                request.targetLabel() == null || request.targetLabel().isBlank() ? path.label() : request.targetLabel(),
                matchScore,
                path.gapItems(),
                path.actionItems(),
                path.milestones(),
                suggestions.stream().limit(5).toList());
    }

    private DevelopmentPathDTO buildPath(String key, String label, int matchScore, double gpa,
                                         int academic, int practice, int culture,
                                         int quality, int health) {
        LocalDate now = LocalDate.now();
        String m1 = now.minusMonths(1).toString();
        String m2 = now.plusMonths(1).toString();
        String m3 = now.plusMonths(3).toString();
        String m4 = now.plusMonths(5).toString();
        String m5 = now.plusMonths(7).toString();
        String d1 = now.plusMonths(1).toString();
        String d2 = now.plusMonths(2).toString();
        String d3 = now.plusMonths(3).toString();
        String d4 = now.plusMonths(4).toString();
        String d5 = now.plusMonths(5).toString();

        return switch (key) {
            case "graduate" -> new DevelopmentPathDTO(
                    key, label, matchScore,
                    List.of(
                            gap("学业成绩", round(gpa), "3.80+", academic < 75 ? "high" : "medium"),
                            gap("英语水平", "CET-6 528", "CET-6 570+ / 精读 80+", "high"),
                            gap("竞赛经历", "课程项目 2 项", "省级以上奖项", academic >= 85 ? "low" : "medium"),
                            gap("科研基础", "无论文/项目", "参与 1 项科研", "medium")),
                    List.of(
                            action("报名考研英语强化班", "系统提升阅读与写作能力，目标 CET-6 570+", "high", d1),
                            action("参加数学建模竞赛", "锻炼建模与团队协作能力，提升竞赛背景", "medium", d3),
                            action("联系导师参与科研", "争取进入实验室，积累科研经验", "low", d5),
                            action("提升核心课成绩", "本学期重点巩固专业核心课程", "high", d2)),
                    milestones(
                            new String[][]{{"确定目标院校与专业", m1, "done", "结合地域与学科排名完成目标确认"},
                                    {"完成考研报名", m2, "current", "关注报名截止时间"},
                                    {"初试冲刺复习", m3, "pending", "三轮复习与模拟训练"},
                                    {"参加初试", m4, "pending", "政治、英语、专业课"},
                                    {"复试准备", m5, "pending", "专业课面试与英文自我介绍"}}));
            case "overseas" -> new DevelopmentPathDTO(
                    key, label, matchScore,
                    List.of(
                            gap("托福/雅思", "未考取", "托福 100+ / 雅思 7.0+", "high"),
                            gap("学业成绩", round(gpa), "3.7+ (4.0 制)", "medium"),
                            gap("英文科研成果", "无", "至少 1 篇英文论文/Report", "high"),
                            gap("推荐信", "未准备", "3 封教授推荐信", "medium")),
                    List.of(
                            action("备考托福", "目标托福 105+，建议系统学习", "high", d1),
                            action("联系海外导师", "通过邮件了解实验室研究方向", "medium", d3),
                            action("准备英文成绩单", "申请英文版成绩单和在读证明", "low", d5)),
                    milestones(
                            new String[][]{{"确定意向国家/学校", m1, "done", "综合预算与专业排名确认"},
                                    {"完成托福考试", m2, "current", "目标 100 分以上"},
                                    {"完善文书与简历", m3, "pending", "PS/CV/推荐信同步准备"},
                                    {"提交申请材料", m4, "pending", "关注各校截止日期"},
                                    {"收到录取通知", m5, "pending", "确认入读院校"}}));
            case "civil" -> new DevelopmentPathDTO(
                    key, label, matchScore,
                    List.of(
                            gap("行测能力", "未系统训练", "模拟题 75+", "high"),
                            gap("申论写作", "未训练", "掌握政论文框架", "high"),
                            gap("时事政治", "一般了解", "系统学习近年政策", "medium"),
                            gap("专业加分", "无行政经验", "学生会/志愿服务经历", "low")),
                    List.of(
                            action("行测系统训练", "数量关系、言语理解、判断推理三大模块", "high", d1),
                            action("订阅时政学习资源", "每日阅读人民日报/半月谈", "medium", d1),
                            action("参加志愿服务活动", "积累社会服务经历，丰富简历", "low", d3)),
                    milestones(
                            new String[][]{{"了解报考条件", m1, "done", "核对岗位与专业要求"},
                                    {"开始行测系统学习", m2, "current", "每日 2 小时专项训练"},
                                    {"国考报名", m3, "pending", "国家公务员考试报名"},
                                    {"参加笔试", m4, "pending", "行测与申论"},
                                    {"面试准备", m5, "pending", "结构化面试训练"}}));
            case "institution" -> new DevelopmentPathDTO(
                    key, label, matchScore,
                    List.of(
                            gap("职业能力测验", "未系统训练", "模拟题 70+", "high"),
                            gap("综合应用写作", "未训练", "掌握机关文体格式", "high"),
                            gap("专业知识积累", "专业课合格", "相关资格证书", "medium"),
                            gap("面试表达", "无结构化训练", "熟悉结构化面试框架", "medium")),
                    List.of(
                            action("关注招考公告", "筛选适合专业的编制岗位", "high", d1),
                            action("职测模块学习", "重点攻克判断推理与资料分析", "high", d2),
                            action("考取资格证书", "教师资格证或计算机等级证书", "medium", d4),
                            action("准备结构化面试", "熟悉综合分析、人际、组织协调题型", "medium", d5)),
                    milestones(
                            new String[][]{{"确定目标单位类型", m1, "done", "按专业筛选事业单位方向"},
                                    {"完成职测基础学习", m2, "current", "完成一轮模块训练"},
                                    {"统考报名", m3, "pending", "省/市事业单位统考"},
                                    {"参加笔试", m4, "pending", "职测与综合应用"},
                                    {"面试及体检", m5, "pending", "结构化面试与体检"}}));
            default -> new DevelopmentPathDTO(
                    key, label, matchScore,
                    List.of(
                            gap("实习经历", "无实习", "1 段互联网/研发实习", "high"),
                            gap("项目作品集", "2 个课程项目", "3-5 个完整作品", "high"),
                            gap("算法刷题", "LeetCode 50 题", "LeetCode 200+ 题", "medium"),
                            gap("技术栈深度", "Java 基础", "后端/前端方向专精", "medium")),
                    List.of(
                            action("投递暑期实习", "目标头部互联网公司实习", "high", d1),
                            action("完善作品集", "整理课程项目并补充 README", "high", d1),
                            action("坚持刷题", "每天 2-3 道，重点攻克动态规划", "medium", d3),
                            action("学习主流框架", "Spring Boot / React 选择一个方向深入", "medium", d5)),
                    milestones(
                            new String[][]{{"确定求职方向", m1, "done", "后端/前端/算法三选一"},
                                    {"获得暑期实习 offer", m2, "current", "完成笔试与面试准备"},
                                    {"完成实习获得转正评估", m3, "pending", "争取转正机会"},
                                    {"参加秋招", m4, "pending", "校招黄金期 9-11 月"},
                                    {"签署就业协议", m5, "pending", "确认薪酬与工作地点"}}));
        };
    }

    private DevelopmentPathDTO.GapItemDTO gap(String dimension, Object current, Object target, String urgent) {
        return new DevelopmentPathDTO.GapItemDTO(dimension, current, target, urgent);
    }

    private DevelopmentPathDTO.ActionItemDTO action(String title, String description, String urgency, String deadline) {
        return new DevelopmentPathDTO.ActionItemDTO(title, description, urgency, deadline);
    }

    private List<DevelopmentPathDTO.MilestoneItemDTO> milestones(String[][] rows) {
        List<DevelopmentPathDTO.MilestoneItemDTO> list = new ArrayList<>();
        for (String[] row : rows) {
            list.add(new DevelopmentPathDTO.MilestoneItemDTO(row[0], row[1], row[2], row[3]));
        }
        return list;
    }

    private JobItemDTO toJobDto(JobCache job) {
        return new JobItemDTO(
                String.valueOf(job.getId()),
                job.getPlatform(),
                job.getPlatformLabel(),
                job.getCompanyName(),
                job.getCompanySize(),
                job.getCompanyStage(),
                job.getJobTitle(),
                job.getSalaryRange(),
                job.getCity(),
                job.getDistrict(),
                job.getEducation(),
                job.getExperience(),
                job.getTags() == null ? List.of() : job.getTags(),
                job.getHighlights() == null ? List.of() : job.getHighlights(),
                job.getMatchScore() == null ? 0 : job.getMatchScore(),
                job.getMatchReasons() == null ? List.of() : job.getMatchReasons(),
                job.getPublishDate(),
                job.getSourceUrl(),
                job.getCategory());
    }

    private SchoolItemDTO toSchoolDto(SchoolCache school) {
        return new SchoolItemDTO(
                String.valueOf(school.getId()),
                school.getType(),
                school.getName(),
                school.getNameZh(),
                school.getCountry(),
                school.getCountryEmoji(),
                school.getLocation(),
                school.getRank(),
                school.getTier(),
                school.getPrograms() == null ? List.of() : school.getPrograms(),
                school.getAdmissionGpa(),
                school.getExamRequirements() == null ? List.of() : school.getExamRequirements(),
                school.getLanguageRequirement(),
                school.getHighlights() == null ? List.of() : school.getHighlights(),
                school.getMatchScore() == null ? 0 : school.getMatchScore(),
                school.getMatchReasons() == null ? List.of() : school.getMatchReasons(),
                school.getOfficialUrl(),
                school.getDeadline());
    }

    private Map<String, Integer> profileScores(Long studentId) {
        List<GpaHistory> histories = gpaHistoryMapper.selectList(new LambdaQueryWrapper<GpaHistory>()
                .eq(GpaHistory::getStudentId, studentId)
                .orderByDesc(GpaHistory::getTerm));
        String term = histories.isEmpty() ? "2024-2025-1" : histories.get(0).getTerm();
        List<ProfileScore> scores = profileScoreMapper.selectList(new LambdaQueryWrapper<ProfileScore>()
                .eq(ProfileScore::getStudentId, studentId)
                .eq(ProfileScore::getTerm, term));
        Map<String, Integer> map = new HashMap<>();
        for (ProfileScore score : scores) {
            map.put(score.getDimensionKey(), score.getScore());
        }
        return map;
    }

    private Student requireStudent(Long userId) {
        Student student = studentMapper.selectOne(new LambdaQueryWrapper<Student>()
                .eq(Student::getUserId, userId));
        if (student == null) {
            throw new BusinessException(404, "学生档案不存在");
        }
        return student;
    }

    private double round(double value) {
        return BigDecimal.valueOf(value).setScale(2, java.math.RoundingMode.HALF_UP).doubleValue();
    }

    private int clamp(int value, int min, int max) {
        return Math.max(min, Math.min(max, value));
    }
}
