<script setup lang="ts">
/* 告警管理：规则 / 记录 两个 tab，按需挂载保证表格高度测算正确 */
import { ref } from 'vue'
import RuleTab from '@/components/alerts/RuleTab.vue'
import RecordTab from '@/components/alerts/RecordTab.vue'

const active = ref('rules')
</script>

<template>
  <div class="page-root">
    <div class="alerts-view">
      <el-tabs v-model="active" class="alerts-view__tabs">
        <el-tab-pane label="告警规则" name="rules">
          <RuleTab v-if="active === 'rules'" />
        </el-tab-pane>
        <el-tab-pane label="告警记录" name="records">
          <RecordTab v-if="active === 'records'" />
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<style scoped>
.alerts-view {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.alerts-view__tabs {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
/* EP tabs 结构：header 固定，内容区占满剩余高度 */
.alerts-view__tabs :deep(.el-tabs__header) {
  flex-shrink: 0;
}
.alerts-view__tabs :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
}
.alerts-view__tabs :deep(.el-tab-pane) {
  height: 100%;
  min-height: 0;
}
</style>
