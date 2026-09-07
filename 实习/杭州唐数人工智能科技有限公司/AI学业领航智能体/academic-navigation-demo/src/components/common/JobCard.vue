<script setup lang="ts">
import type { JobItem } from '@/types'

defineProps<{ job: JobItem }>()

const platformColor: Record<JobItem['platform'], string> = {
  boss:    '#00b37e',
  '51job': '#e45b3c',
  zhilian: '#004ede',
  liepin:  '#f5820a',
}

function openJob(url: string) {
  window.open(url, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <div class="jcard">
    <!-- 头部：公司 + 平台标记 -->
    <div class="jcard-top">
      <div class="company-avatar">{{ job.companyName[0] }}</div>
      <div class="company-info">
        <div class="company-name">{{ job.companyName }}</div>
        <div class="company-meta">{{ job.companySize }} · {{ job.companyStage }}</div>
      </div>
      <div
        class="platform-badge"
        :style="{ color: platformColor[job.platform], background: `${platformColor[job.platform]}14` }"
      >{{ job.platformLabel }}</div>
    </div>

    <!-- 职位标题 + 薪资 -->
    <div class="jcard-main">
      <div class="job-title">{{ job.jobTitle }}</div>
      <div class="job-salary">{{ job.salaryRange }}</div>
    </div>

    <!-- 位置 + 要求 -->
    <div class="jcard-meta-row">
      <span class="meta-chip">
        <el-icon><Location /></el-icon>{{ job.city }}·{{ job.district }}
      </span>
      <span class="meta-chip">
        <el-icon><School /></el-icon>{{ job.education }}
      </span>
      <span class="meta-chip">
        <el-icon><Briefcase /></el-icon>{{ job.experience }}
      </span>
    </div>

    <!-- 技能标签 -->
    <div class="jcard-tags">
      <span v-for="tag in job.tags.slice(0, 4)" :key="tag" class="skill-tag">{{ tag }}</span>
    </div>

    <!-- 匹配度 -->
    <div class="jcard-match">
      <div class="match-header">
        <span class="match-label">画像匹配度</span>
        <span
          class="match-score"
          :style="{ color: job.matchScore >= 85 ? '#3db87e' : job.matchScore >= 70 ? '#2A4D99' : '#f09a4e' }"
        >{{ job.matchScore }}%</span>
      </div>
      <div class="match-bar-track">
        <div
          class="match-bar-fill"
          :style="{
            width: job.matchScore + '%',
            background: job.matchScore >= 85 ? '#3db87e' : job.matchScore >= 70 ? '#2A4D99' : '#f09a4e'
          }"
        />
      </div>
      <ul class="match-reasons">
        <li v-for="r in job.matchReasons.slice(0, 2)" :key="r">{{ r }}</li>
      </ul>
    </div>

    <!-- 底部：亮点 + 操作 -->
    <div class="jcard-footer">
      <div class="highlights">
        <span v-for="h in job.highlights.slice(0, 2)" :key="h" class="highlight-chip">{{ h }}</span>
      </div>
      <button class="btn-view" @click="openJob(job.sourceUrl)">
        查看详情
        <el-icon><ArrowRight /></el-icon>
      </button>
    </div>

    <div class="publish-date">发布于 {{ job.publishDate }}</div>
  </div>
</template>

<style scoped>
.jcard {
  background: #fff;
  border-radius: 16px;
  padding: 18px 20px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
  display: flex; flex-direction: column; gap: 12px;
  transition: box-shadow 0.18s ease, transform 0.18s ease;
}
.jcard:hover {
  box-shadow: 0 4px 8px rgba(10,19,41,0.06), 0 16px 40px rgba(10,19,41,0.09);
  transform: translateY(-2px);
}

.jcard-top { display: flex; align-items: center; gap: 10px; }
.company-avatar {
  width: 36px; height: 36px; border-radius: 10px; flex-shrink: 0;
  background: linear-gradient(135deg, #edf0f6, #d8dff0);
  display: flex; align-items: center; justify-content: center;
  font-size: 15px; font-weight: 800; color: #2A4D99;
}
.company-info { flex: 1; min-width: 0; }
.company-name { font-size: 13px; font-weight: 700; color: #111827; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.company-meta { font-size: 11.5px; color: #9facc5; margin-top: 1px; }
.platform-badge {
  font-size: 11px; font-weight: 700;
  padding: 3px 8px; border-radius: 6px; flex-shrink: 0;
}

.jcard-main { display: flex; align-items: baseline; justify-content: space-between; gap: 8px; }
.job-title { font-size: 15px; font-weight: 800; color: #101d3e; line-height: 1.3; }
.job-salary { font-size: 14.5px; font-weight: 800; color: #e04538; white-space: nowrap; flex-shrink: 0; }

.jcard-meta-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.meta-chip {
  display: flex; align-items: center; gap: 3px;
  font-size: 12px; color: #6b7280;
}
.meta-chip .el-icon { font-size: 12px; color: #b4bed2; }

.jcard-tags { display: flex; gap: 6px; flex-wrap: wrap; }
.skill-tag {
  font-size: 11.5px; color: #374151;
  background: #f0f2f7; padding: 2px 8px; border-radius: 6px;
}

.jcard-match {}
.match-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.match-label { font-size: 12px; font-weight: 600; color: #9facc5; }
.match-score { font-size: 14px; font-weight: 800; }
.match-bar-track { width: 100%; height: 4px; background: #f0f2f7; border-radius: 99px; overflow: hidden; margin-bottom: 8px; }
.match-bar-fill { height: 100%; border-radius: 99px; transition: width 0.6s cubic-bezier(0.34,1.56,0.64,1); }
.match-reasons {
  margin: 0; padding: 0; list-style: none;
  display: flex; flex-direction: column; gap: 3px;
}
.match-reasons li {
  font-size: 11.5px; color: #6b7280; line-height: 1.5;
  padding-left: 10px; position: relative;
}
.match-reasons li::before {
  content: ''; position: absolute; left: 0; top: 7px;
  width: 3px; height: 3px; border-radius: 50%; background: #3db87e;
}

.jcard-footer { display: flex; align-items: center; justify-content: space-between; }
.highlights { display: flex; gap: 6px; flex-wrap: wrap; }
.highlight-chip {
  font-size: 11.5px; color: #2A4D99;
  background: rgba(42,77,153,0.07);
  padding: 2px 8px; border-radius: 6px;
}

.btn-view {
  display: flex; align-items: center; gap: 5px;
  padding: 7px 14px; border-radius: 9px;
  background: #2A4D99; color: #fff;
  font-size: 12.5px; font-weight: 700;
  border: none; cursor: pointer; flex-shrink: 0;
  transition: background 0.15s ease, transform 0.15s ease;
}
.btn-view:hover { background: #3a5fba; transform: translateY(-1px); }
.btn-view:active { transform: translateY(0); }
.btn-view .el-icon { font-size: 12px; }

.publish-date { font-size: 11px; color: #c5cdd8; text-align: right; margin-top: -4px; }
</style>
