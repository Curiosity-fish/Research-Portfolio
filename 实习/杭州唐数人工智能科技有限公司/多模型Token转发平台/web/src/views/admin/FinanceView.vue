<script setup lang="ts">
/* 资金与计费：充值订单 / 余额流水 / 配额审批 / 配额流水 四个 tab，按需挂载 */
import { ref } from 'vue'
import OrderTab from '@/components/finance/OrderTab.vue'
import BalanceTab from '@/components/finance/BalanceTab.vue'
import QuotaRequestTab from '@/components/finance/QuotaRequestTab.vue'
import QuotaRecordTab from '@/components/finance/QuotaRecordTab.vue'

const active = ref('orders')
</script>

<template>
  <div class="page-root">
    <div class="fin-view">
      <el-tabs v-model="active" class="fin-view__tabs">
        <el-tab-pane label="充值订单" name="orders">
          <OrderTab v-if="active === 'orders'" />
        </el-tab-pane>
        <el-tab-pane label="余额流水" name="balance">
          <BalanceTab v-if="active === 'balance'" />
        </el-tab-pane>
        <el-tab-pane label="配额审批" name="quota-requests">
          <QuotaRequestTab v-if="active === 'quota-requests'" />
        </el-tab-pane>
        <el-tab-pane label="配额流水" name="quota-records">
          <QuotaRecordTab v-if="active === 'quota-records'" />
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<style scoped>
.fin-view {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.fin-view__tabs {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.fin-view__tabs :deep(.el-tabs__header) {
  flex-shrink: 0;
}
.fin-view__tabs :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
}
.fin-view__tabs :deep(.el-tab-pane) {
  height: 100%;
  min-height: 0;
}
</style>
