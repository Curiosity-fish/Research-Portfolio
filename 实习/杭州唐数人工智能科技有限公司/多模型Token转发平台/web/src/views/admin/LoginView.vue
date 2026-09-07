<script setup lang="ts">
/* 管理员登录页：复刻旧项目 PortalView 布局——左侧固定宽度表单区，右侧产品展示图 + 打字机标题 + 功能亮点
   整页 100vh 不滚动，窄屏（<1024px）隐藏右侧展示区 */
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const rememberMe = ref(false)
const form = reactive({ username: '', password: '' })

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' },
  ],
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await auth.loginAndFetch({ username: form.username, password: form.password })
      ElMessage.success('登录成功')
      const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/admin'
      router.push(redirect)
    } finally {
      loading.value = false
    }
  })
}

/* ── 打字机效果：/start: 从这里开始，连接更高效的校园 AI 服务 ──
    orangePart = "/start: 从这里开始，连接更高效的"（逐字打出，橙色）
   navyPart   = "校园 AI 服务"（整体淡入，藏青） */
const TYPE_TEXT = '/start: 从这里开始，连接更高效的'
const NAVY_TEXT = '校园 AI 服务'
const typeCount = ref(0)
const navyVisible = ref(false)
let typeTimer: ReturnType<typeof setInterval> | undefined
let navyTimer: ReturnType<typeof setTimeout> | undefined

onMounted(() => {
  typeTimer = setInterval(() => {
    typeCount.value += 1
    if (typeCount.value >= TYPE_TEXT.length) {
      if (typeTimer) clearInterval(typeTimer)
      typeTimer = undefined
      navyTimer = setTimeout(() => {
        navyVisible.value = true
      }, 120)
    }
  }, 70)
})

onBeforeUnmount(() => {
  if (typeTimer) clearInterval(typeTimer)
  if (navyTimer) clearTimeout(navyTimer)
})
</script>

<template>
  <div class="login-view">
    <!-- 左侧登录区 -->
    <div class="login-view__left">
      <div class="login-view__divider" />
      <div class="login-view__left-inner">
        <div class="login-view__inner">
          <!-- Logo + 平台名：图标与标题同一行居中，副标题单独一行 -->
          <div class="login-view__logo">
            <img src="/platform-logo.png" alt="校园Token管理平台" class="login-view__logo-img" />
            <div class="login-view__logo-title">校园 Token 管理平台</div>
          </div>
          <div class="login-view__logo-sub">统一管理校园 AI 服务访问与使用配额</div>

          <!-- 表单：与旧版 PortalView 一致的字段标签 + 记住我/忘记密码 -->
          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            size="large"
            label-position="top"
            hide-required-asterisk
            class="login-view__form"
            @keyup.enter="submit"
          >
            <el-form-item prop="username" label="用户名">
              <el-input
                v-model="form.username"
                placeholder="请输入用户名"
                :prefix-icon="'User'"
                autocomplete="username"
              />
            </el-form-item>
            <el-form-item prop="password" label="密码">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="请输入密码"
                :prefix-icon="'Lock'"
                autocomplete="current-password"
                show-password
              />
            </el-form-item>
            <div class="login-view__form-row">
              <el-checkbox v-model="rememberMe">记住我</el-checkbox>
              <span class="login-view__forgot">忘记密码？请联系管理员</span>
            </div>
            <el-button class="login-view__submit" type="primary" :loading="loading" @click="submit">
              登 录
            </el-button>
          </el-form>

          <!-- 安全提示 -->
          <div class="login-view__tip">
            <el-icon :size="13"><Lock /></el-icon>
            <span>请妥善保管账号与 Token，避免向他人泄露。</span>
          </div>

          <p class="login-view__copy">© {{ new Date().getFullYear() }} 校园 Token 管理平台 · 仅限内部使用</p>
        </div>
      </div>
    </div>

    <!-- 右侧展示区 -->
    <div class="login-view__right">
      <!-- 品牌氛围光晕 -->
      <div class="login-view__glow login-view__glow--top" />
      <div class="login-view__glow login-view__glow--bottom" />

      <div class="login-view__right-inner">
        <!-- 打字机标题（顶部居中，衬线体大字） -->
        <h1 class="login-view__headline" aria-label="/start: 从这里开始，连接更高效的校园 AI 服务">
          <span class="login-view__type">{{ TYPE_TEXT.slice(0, typeCount) }}</span><span
            class="login-view__caret"
            :class="{ 'login-view__caret--done': typeCount >= TYPE_TEXT.length }"
          /><span class="login-view__navy" :class="{ 'login-view__navy--visible': navyVisible }">{{ NAVY_TEXT }}</span>
        </h1>

        <!-- 产品展示图：等比完整显示，随窗口缩放不变形 -->
        <div class="login-view__shot">
          <img src="/login-right.png" alt="平台控制台预览" class="login-view__shot-img" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-view {
  height: 100vh;
  display: flex;
  overflow: hidden;
  background: var(--bg-page);
}

/* ── 左侧登录区 ── */
.login-view__left {
  width: 100%;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
  background: var(--bg-card);
}
.login-view__divider {
  display: none;
  position: absolute;
  right: 0;
  top: 32px;
  bottom: 32px;
  width: 1px;
  background: var(--border-light);
}
.login-view__left-inner {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px 20px;
  overflow: hidden;
}
.login-view__inner {
  width: 100%;
  max-width: 560px;
}

/* Logo 区：图标与标题同一行居中。字号用 clamp 随左栏宽度（34vw）流式增长，
   单行总宽 ≈ 11.2em（标题 9.1em + 图 1.78em + 间距 0.32em），
   各档位预留约 2% 余量，保证 400-464px 内容区下绝不换行 */
.login-view__logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.32em;
  margin-bottom: 12px;
  font-size: clamp(31px, 2.65vw, 46px);
}
.login-view__logo-img {
  width: 1.78em;
  height: 1.78em;
  object-fit: contain;
  flex-shrink: 0;
}
.login-view__logo-title {
  font-size: 1em;
  font-weight: 900;
  letter-spacing: -0.02em;
  line-height: 1.2;
  color: var(--text-primary);
}
.login-view__logo-sub {
  font-size: clamp(15px, 1.3vw, 20px);
  font-weight: 500;
  text-align: center;
  margin-bottom: 28px;
  color: var(--text-secondary);
}

/* 表单：label-position="top" 的字段标签样式（对齐旧版 13.5px 粗体标签） */
.login-view__form :deep(.el-form-item__label) {
  font-size: 13.5px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.4;
  padding-bottom: 6px;
}
.login-view__form :deep(.el-form-item) {
  margin-bottom: 16px;
}
/* 校验错误占位渲染：EP 默认绝对定位于项底部，本页表单间距压缩到 16px，
   错误文字会覆盖下一个字段标签；改回流内布局，出现错误时表单自然撑高 */
.login-view__form :deep(.el-form-item__error) {
  position: static;
  padding-top: 4px;
  line-height: 1.2;
}
.login-view__form :deep(.el-input__wrapper) {
  border-radius: 10px;
  min-height: 46px;
  padding: 0 14px;
}
.login-view__form :deep(.el-input__inner) {
  height: 44px;
  font-size: 14px;
}
/* 记住我 + 忘记密码 行 */
.login-view__form-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 2px 0 14px;
}
.login-view__form-row :deep(.el-checkbox__label) {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}
.login-view__forgot {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
}
.login-view__submit {
  width: 100%;
  height: 48px;
  margin-top: 4px;
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 6px;
  border-radius: 10px;
  box-shadow:
    0 4px 14px rgba(154, 109, 18, 0.4),
    0 1px 3px rgba(107, 78, 15, 0.25);
  transition: all 0.15s;
}
.login-view__submit:hover {
  transform: translateY(-1px);
  box-shadow:
    0 7px 20px rgba(154, 109, 18, 0.5),
    0 2px 4px rgba(107, 78, 15, 0.3);
}

/* 安全提示 */
.login-view__tip {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding: 10px 14px;
  border-radius: 8px;
  background: var(--brand-light);
  border: 1px solid var(--warning-border);
  font-size: 12px;
  font-weight: 500;
  color: var(--brand-text);
}
.login-view__copy {
  margin-top: 16px;
  font-size: 11px;
  font-weight: 500;
  text-align: center;
  color: var(--text-muted);
}

/* ── 右侧展示区 ── */
.login-view__right {
  display: none;
  flex: 1;
  min-width: 0;
  align-items: center;
  justify-content: center;
  padding: 24px 32px;
  position: relative;
  overflow: hidden;
  background: var(--bg-page);
}
.login-view__glow {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
}
.login-view__glow--top {
  top: -128px;
  right: 6%;
  width: 560px;
  height: 460px;
  background: radial-gradient(closest-side, rgba(154, 109, 18, 0.28) 0%, rgba(154, 109, 18, 0.1) 55%, transparent 100%);
}
.login-view__glow--bottom {
  bottom: -160px;
  left: -96px;
  width: 520px;
  height: 420px;
  background: radial-gradient(closest-side, rgba(154, 109, 18, 0.14) 0%, transparent 70%);
}
.login-view__right-inner {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 760px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  align-self: stretch;
}

/* 打字机标题：顶部居中，衬线体（Noto Serif SC）大字，字号随右侧区宽度增长 */
.login-view__headline {
  margin: 0 0 20px;
  font-family: 'Noto Serif SC', 'Noto Sans SC', serif;
  font-size: clamp(20px, 2.4vw, 34px);
  font-weight: 900;
  letter-spacing: 0.01em;
  line-height: 1.35;
  white-space: nowrap;
  overflow: hidden;
  text-align: center;
  flex-shrink: 0;
}
.login-view__type {
  color: #d97706;
}
.login-view__caret {
  display: inline-block;
  width: 3px;
  height: 0.85em;
  margin: 0 2px;
  vertical-align: -0.08em;
  background: #d97706;
  animation: login-caret-blink 0.85s steps(1) infinite;
}
.login-view__caret--done {
  animation: login-caret-blink 1.1s steps(1) infinite;
}
@keyframes login-caret-blink {
  0%,
  55% {
    opacity: 1;
  }
  56%,
  100% {
    opacity: 0;
  }
}
.login-view__navy {
  color: #1f3a5f;
  opacity: 0;
  transition: opacity 0.5s ease;
}
.login-view__navy--visible {
  opacity: 1;
}

/* 产品展示图：等比完整显示，不裁剪，随窗口缩放保持宽高比 */
.login-view__shot {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: none;
  box-shadow: none;
  background: transparent;
}
.login-view__shot-img {
  display: block;
  max-width: 100%;
  max-height: 100%;
  width: auto;
  height: auto;
  object-fit: contain;
}

/* 功能亮点 */
.login-view__features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 200px), 1fr));
  gap: 12px;
  margin-top: 16px;
  width: 100%;
  flex-shrink: 0;
}
.login-view__feature-card {
  border-radius: 10px;
  padding: 14px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border-light);
}
.login-view__feature-head {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}
.login-view__feature-icon {
  color: var(--brand);
  flex-shrink: 0;
}
.login-view__feature-title {
  font-size: 12.5px;
  font-weight: 800;
  color: var(--text-primary);
}
.login-view__feature-desc {
  margin: 0;
  font-size: 11.5px;
  font-weight: 500;
  line-height: 17px;
  color: var(--text-muted);
}

/* ≥1024px：左侧收窄固定 34%（最小 400px），右侧展示区分隔线出现 */
@media (min-width: 1024px) {
  .login-view__left {
    width: 34%;
    min-width: 400px;
  }
  .login-view__divider {
    display: block;
  }
  .login-view__right {
    display: flex;
  }
}

/* 极矮窗口兜底：压缩标题与功能卡高度，保证一屏不滚动 */
@media (max-height: 680px) {
  .login-view__shot-img {
    max-height: 46vh;
  }
}
/* 窄右侧（1024-1199px）：右侧区宽度不足，标题改用固定小字号避免被裁切 */
@media (min-width: 1024px) and (max-width: 1199px) {
  .login-view__right {
    align-items: flex-start;
  }
  .login-view__headline {
    font-size: 14px;
  }
}
/* 窄高度窗口（700px 以下）：缩小标题字号，腾出纵向空间 */
@media (max-height: 700px) {
  .login-view__logo {
    margin-bottom: 6px;
  }
  .login-view__logo-img {
    width: 40px;
    height: 40px;
  }
  .login-view__logo-title {
    font-size: 20px;
    line-height: 26px;
  }
  .login-view__logo-sub {
    font-size: 12px;
    margin-bottom: 16px;
  }
  .login-view__form :deep(.el-form-item) {
    margin-bottom: 12px;
  }
  .login-view__form-row {
    margin-bottom: 10px;
  }
  .login-view__tip {
    margin-top: 12px;
    padding-top: 8px;
    padding-bottom: 8px;
  }
  .login-view__copy {
    margin-top: 12px;
  }
}
</style>
