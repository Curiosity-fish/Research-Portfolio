<script setup lang="ts">
/* 转发管理：平台 / 账号 / 模型 / 分组绑定 四个 tab，按需挂载保证表格高度测算正确 */
import { ref } from 'vue'
import PlatformTab from '@/components/relay/PlatformTab.vue'
import AccountTab from '@/components/relay/AccountTab.vue'
import ModelTab from '@/components/relay/ModelTab.vue'
import GroupBindTab from '@/components/relay/GroupBindTab.vue'

const active = ref('platform')
</script>

<template>
  <div class="page-root">
    <div class="relay-view">
      <el-tabs v-model="active" class="relay-view__tabs">
        <el-tab-pane label="平台管理" name="platform">
          <PlatformTab v-if="active === 'platform'" />
        </el-tab-pane>
        <el-tab-pane label="账号管理" name="account">
          <AccountTab v-if="active === 'account'" />
        </el-tab-pane>
        <el-tab-pane label="模型管理" name="model">
          <ModelTab v-if="active === 'model'" />
        </el-tab-pane>
        <el-tab-pane label="分组绑定" name="group">
          <GroupBindTab v-if="active === 'group'" />
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<style scoped>
.relay-view {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.relay-view__tabs {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
/* EP tabs 结构：header 固定，内容区占满剩余高度 */
.relay-view__tabs :deep(.el-tabs__header) {
  flex-shrink: 0;
}
.relay-view__tabs :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
}
.relay-view__tabs :deep(.el-tab-pane) {
  height: 100%;
  min-height: 0;
}
</style>
