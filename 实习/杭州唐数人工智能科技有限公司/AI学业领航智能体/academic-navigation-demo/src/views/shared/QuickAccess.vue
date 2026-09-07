<script setup lang="ts">
interface SystemLink {
  id: string
  name: string
  desc: string
  icon: string
  color: string
  bg: string
  url: string
  tag?: string
}

interface SystemGroup {
  label: string
  accentColor: string
  items: SystemLink[]
}

const groups: SystemGroup[] = [
  {
    label: '教学管理',
    accentColor: '#2A4D99',
    items: [
      {
        id: 'jwxt',
        name: '教务系统',
        desc: '选课 · 成绩查询 · 课表',
        icon: 'School',
        color: '#2A4D99',
        bg: 'rgba(42,77,153,0.07)',
        url: 'http://jwglxt.zjnu.edu.cn',
        tag: '常用',
      },
      {
        id: 'bb',
        name: '网络教学平台',
        desc: '课程资料 · 作业提交 · 直播',
        icon: 'Monitor',
        color: '#0077cc',
        bg: 'rgba(0,119,204,0.07)',
        url: 'http://bb.zjnu.edu.cn',
        tag: '常用',
      },
      {
        id: 'thesis',
        name: '毕业论文管理系统',
        desc: '选题 · 任务书 · 进度提交',
        icon: 'Document',
        color: '#5b21b6',
        bg: 'rgba(91,33,182,0.07)',
        url: 'http://bylw.zjnu.edu.cn',
      },
      {
        id: 'exam',
        name: '考试报名系统',
        desc: '四六级 · 研究生 · 专业认证',
        icon: 'EditPen',
        color: '#be5d18',
        bg: 'rgba(190,93,24,0.07)',
        url: 'http://exam.zjnu.edu.cn',
      },
    ],
  },
  {
    label: '校园生活',
    accentColor: '#3db87e',
    items: [
      {
        id: 'card',
        name: '一卡通系统',
        desc: '余额查询 · 消费记录 · 充值',
        icon: 'CreditCard',
        color: '#3db87e',
        bg: 'rgba(61,184,126,0.07)',
        url: 'http://card.zjnu.edu.cn',
        tag: '常用',
      },
      {
        id: 'lib',
        name: '图书馆系统',
        desc: '馆藏检索 · 借阅记录 · 续借',
        icon: 'Reading',
        color: '#c8762a',
        bg: 'rgba(200,118,42,0.07)',
        url: 'http://lib.zjnu.edu.cn',
      },
      {
        id: 'dorm',
        name: '宿舍管理系统',
        desc: '报修 · 电费查询 · 门禁记录',
        icon: 'House',
        color: '#0e7490',
        bg: 'rgba(14,116,144,0.07)',
        url: 'http://dorm.zjnu.edu.cn',
      },
      {
        id: 'net',
        name: '校园网服务',
        desc: 'VPN · 网络认证 · 套餐管理',
        icon: 'Connection',
        color: '#047857',
        bg: 'rgba(4,120,87,0.07)',
        url: 'http://net.zjnu.edu.cn',
      },
    ],
  },
  {
    label: '综合服务',
    accentColor: '#f09a4e',
    items: [
      {
        id: 'mail',
        name: '学校邮箱',
        desc: '官方邮件 · 通知接收',
        icon: 'Message',
        color: '#e04538',
        bg: 'rgba(224,69,56,0.07)',
        url: 'https://mail.zjnu.edu.cn',
      },
      {
        id: 'job',
        name: '就业信息网',
        desc: '校招 · 招聘会 · 就业指导',
        icon: 'Suitcase',
        color: '#f09a4e',
        bg: 'rgba(240,154,78,0.07)',
        url: 'http://job.zjnu.edu.cn',
        tag: '推荐',
      },
      {
        id: 'mental',
        name: '心理健康服务',
        desc: '测评 · 预约咨询 · 资源',
        icon: 'Sunny',
        color: '#7c3aed',
        bg: 'rgba(124,58,237,0.07)',
        url: 'http://xl.zjnu.edu.cn',
      },
      {
        id: 'info',
        name: '学工信息系统',
        desc: '奖助学金 · 学籍信息 · 综测',
        icon: 'UserFilled',
        color: '#0369a1',
        bg: 'rgba(3,105,161,0.07)',
        url: 'http://xgxt.zjnu.edu.cn',
      },
    ],
  },
]

function openLink(url: string) {
  window.open(url, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <div class="quick-access">
    <div class="qa-header">
      <div class="qa-header-main">
        <el-icon class="qa-header-icon"><Link /></el-icon>
        <div>
          <h1 class="qa-title">快捷访问</h1>
          <p class="qa-subtitle">一键跳转学校各大系统</p>
        </div>
      </div>
    </div>

    <div class="qa-content">
      <div v-for="group in groups" :key="group.label" class="qa-group">
        <div class="qa-group-label" :style="{ borderLeftColor: group.accentColor }">
          {{ group.label }}
        </div>
        <div class="qa-grid">
          <div
            v-for="item in group.items"
            :key="item.id"
            class="qa-card"
            @click="openLink(item.url)"
          >
            <div class="qa-card-top">
              <div class="qa-card-icon" :style="{ background: item.bg, color: item.color }">
                <el-icon><component :is="item.icon" /></el-icon>
              </div>
              <span v-if="item.tag" class="qa-card-tag" :style="{ background: item.bg, color: item.color }">
                {{ item.tag }}
              </span>
            </div>
            <div class="qa-card-body">
              <div class="qa-card-title">{{ item.name }}</div>
              <div class="qa-card-desc">{{ item.desc }}</div>
            </div>
            <div class="qa-card-footer">
              <span class="qa-card-btn">
                前往
                <el-icon><ArrowRight /></el-icon>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.quick-access {
  background: linear-gradient(180deg, #f9fafb 0%, #f3f4f6 100%);
  padding: 28px 32px;
}

.qa-header {
  max-width: 1200px;
  margin: 0 auto 32px;
}

.qa-header-main {
  display: flex;
  align-items: center;
  gap: 16px;
}

.qa-header-icon {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  background: linear-gradient(135deg, #2A4D99, #3a5fba);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  box-shadow: 0 4px 12px rgba(42,77,153,0.2);
}

.qa-title {
  margin: 0;
  font-size: 28px;
  font-weight: 800;
  color: #111827;
  line-height: 1.2;
}

.qa-subtitle {
  margin: 4px 0 0;
  font-size: 14px;
  color: #6b7280;
}

.qa-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 36px;
}

.qa-group {}

.qa-group-label {
  font-size: 15px;
  font-weight: 700;
  color: #374151;
  margin-bottom: 16px;
  padding-left: 14px;
  border-left: 3px solid #2A4D99;
  line-height: 1.4;
}

.qa-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}

.qa-card {
  background: white;
  border-radius: 16px;
  padding: 20px;
  border: 1px solid rgba(10,19,41,0.06);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 12px rgba(10,19,41,0.04);
  display: flex;
  flex-direction: column;
  gap: 14px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.qa-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(10,19,41,0.06), 0 12px 32px rgba(10,19,41,0.08);
}

.qa-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.qa-card-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.qa-card-tag {
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 6px;
}

.qa-card-body {
  flex: 1;
}

.qa-card-title {
  font-size: 16px;
  font-weight: 700;
  color: #111827;
  margin-bottom: 4px;
}

.qa-card-desc {
  font-size: 13px;
  color: #6b7280;
  line-height: 1.5;
}

.qa-card-footer {}

.qa-card-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 600;
  color: #2A4D99;
  transition: gap 0.2s ease;
}

.qa-card:hover .qa-card-btn {
  gap: 8px;
}

.qa-card-btn .el-icon {
  font-size: 14px;
}
</style>
