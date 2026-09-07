<script setup lang="ts">
import type { ResourceItem } from '@/types'

const props = defineProps<{ resource: ResourceItem }>()

const typeConfig: Record<ResourceItem['type'], { label: string; icon: string; color: string; bg: string }> = {
  video:    { label: '视频课程', icon: 'VideoPlay',  color: '#2A4D99', bg: 'rgba(42,77,153,0.08)' },
  doc:      { label: '文档资料', icon: 'Document',   color: '#3db87e', bg: 'rgba(61,184,126,0.08)' },
  exercise: { label: '题库练习', icon: 'EditPen',    color: '#f09a4e', bg: 'rgba(240,154,78,0.08)' },
  book:     { label: '参考书目', icon: 'Reading',    color: '#8b5cf6', bg: 'rgba(139,92,246,0.08)' },
}

const diffConfig = {
  easy:   { label: '入门', color: '#3db87e' },
  medium: { label: '进阶', color: '#f09a4e' },
  hard:   { label: '深入', color: '#e04538' },
}
</script>

<template>
  <div class="rcard">
    <div class="rcard-top">
      <div class="rcard-type-wrap" :style="{ background: typeConfig[resource.type].bg, color: typeConfig[resource.type].color }">
        <el-icon><component :is="typeConfig[resource.type].icon" /></el-icon>
        <span>{{ typeConfig[resource.type].label }}</span>
      </div>
      <div class="rcard-match">
        <span class="match-num">{{ resource.matchScore }}</span>
        <span class="match-pct">%</span>
        <span class="match-label">匹配</span>
      </div>
    </div>

    <div class="rcard-body">
      <div class="rcard-subject">{{ resource.subject }}</div>
      <div class="rcard-title">{{ resource.title }}</div>
      <div class="rcard-source-row">
        <span class="rcard-source">{{ resource.source }}</span>
        <span v-if="resource.duration" class="rcard-duration">{{ resource.duration }}</span>
      </div>
    </div>

    <div class="rcard-match-bar">
      <div class="match-bar-track">
        <div
          class="match-bar-fill"
          :style="{ width: resource.matchScore + '%', background: resource.matchScore >= 85 ? '#3db87e' : resource.matchScore >= 70 ? '#2A4D99' : '#f09a4e' }"
        />
      </div>
    </div>

    <div class="rcard-reason">
      <el-icon class="reason-icon"><InfoFilled /></el-icon>
      <span>{{ resource.matchReason }}</span>
    </div>

    <div class="rcard-footer">
      <div class="rcard-tags">
        <span v-for="tag in resource.tags.slice(0, 2)" :key="tag" class="rcard-tag">{{ tag }}</span>
        <span class="rcard-diff" :style="{ color: diffConfig[resource.difficulty].color }">
          {{ diffConfig[resource.difficulty].label }}
        </span>
      </div>
      <button class="rcard-btn">
        <el-icon><ArrowRight /></el-icon>
      </button>
    </div>
  </div>
</template>

<style scoped>
.rcard {
  background: #fff;
  border-radius: 16px;
  padding: 18px 20px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: box-shadow 0.18s ease, transform 0.18s ease;
}
.rcard:hover {
  box-shadow: 0 4px 8px rgba(10,19,41,0.06), 0 16px 40px rgba(10,19,41,0.08);
  transform: translateY(-2px);
}

.rcard-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.rcard-type-wrap {
  display: flex; align-items: center; gap: 5px;
  padding: 4px 10px; border-radius: 8px;
  font-size: 12px; font-weight: 600;
}
.rcard-type-wrap .el-icon { font-size: 13px; }

.rcard-match {
  display: flex;
  align-items: baseline;
  gap: 1px;
}
.match-num { font-size: 22px; font-weight: 800; color: #111827; line-height: 1; font-variant-numeric: tabular-nums; }
.match-pct { font-size: 13px; font-weight: 700; color: #6b7280; }
.match-label { font-size: 11px; color: #9ca3af; margin-left: 3px; }

.rcard-body { display: flex; flex-direction: column; gap: 4px; }
.rcard-subject { font-size: 11.5px; font-weight: 600; color: #9facc5; text-transform: uppercase; letter-spacing: 0.04em; }
.rcard-title { font-size: 14.5px; font-weight: 700; color: #111827; line-height: 1.4; }
.rcard-source-row { display: flex; align-items: center; gap: 8px; margin-top: 2px; }
.rcard-source { font-size: 12px; color: #6b7280; }
.rcard-duration {
  font-size: 11.5px; color: #9ca3af;
  padding: 1px 7px; background: #f4f6fa; border-radius: 6px;
}

.rcard-match-bar {}
.match-bar-track { width: 100%; height: 4px; background: #f0f2f7; border-radius: 99px; overflow: hidden; }
.match-bar-fill { height: 100%; border-radius: 99px; transition: width 0.6s cubic-bezier(0.34,1.56,0.64,1); }

.rcard-reason {
  display: flex; align-items: flex-start; gap: 6px;
  padding: 8px 11px;
  background: #f8f9fc; border-radius: 9px;
  font-size: 12px; color: #6b7280; line-height: 1.5;
}
.reason-icon { font-size: 13px; color: #b4bed2; flex-shrink: 0; margin-top: 1px; }

.rcard-footer { display: flex; align-items: center; justify-content: space-between; }
.rcard-tags { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.rcard-tag {
  font-size: 11.5px; color: #6b7280;
  background: #f0f2f7; padding: 2px 8px; border-radius: 6px;
}
.rcard-diff { font-size: 11.5px; font-weight: 700; }

.rcard-btn {
  width: 28px; height: 28px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  background: rgba(42,77,153,0.07); color: #2A4D99;
  border: none; cursor: pointer;
  transition: background 0.15s ease;
}
.rcard-btn:hover { background: rgba(42,77,153,0.15); }
</style>
