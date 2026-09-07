<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const form = reactive({ account: '', password: '' })
const loading = ref(false)
const fieldErrors = reactive({ account: '', password: '' })

const demoAccounts = [
  { label: '学生',     account: '20211001' },
  { label: '班主任',   account: 'T001' },
  { label: '任课老师', account: 'C001' },
  { label: '系主任',   account: 'D20210001' },
  { label: '院领导',   account: 'L20210001' },
]

function fillDemo(account: string) {
  form.account = account
  form.password = '123456'
  fieldErrors.account = ''
  fieldErrors.password = ''
}

function validateForm() {
  fieldErrors.account = form.account.trim() ? '' : '请输入工号或学号'
  fieldErrors.password = form.password ? '' : '请输入密码'
  return !fieldErrors.account && !fieldErrors.password
}

async function handleLogin() {
  if (!validateForm()) return
  // 开发者账号：直接跳转后台平台
  if (form.account === 'admin' && form.password === 'Admin@9000') {
    window.open('http://localhost:3001', '_blank')
    return
  }
  loading.value = true
  const result = await auth.login(form.account.trim(), form.password)
  loading.value = false
  if (!result.ok) {
    ElMessage.error(result.error || '登录失败')
    return
  }
  ElMessage.success('登录成功')
  router.push(result.path!)
}
</script>

<template>
  <div class="login-root">
    <div class="login-brand">
      <div class="login-grid" />
      <div class="login-glow-1" />
      <div class="login-glow-2" />
      <div class="login-motto-watermark">
        <img src="/zjnu-motto.jpg" alt="浙江师范大学校训" />
      </div>

      <div class="login-brand-content">
        <div class="login-emblem">
          <img src="/zjnu-logo.png" alt="浙江师范大学校徽" class="login-emblem-img" />
        </div>
        <h1 class="login-brand-title">学业领航智能体</h1>
        <p class="login-brand-sub">浙江师范大学 · 学生学业导航系统</p>

        <div class="login-brand-features">
          <div class="login-feature"><div class="login-feature-dot" /><span>五维学业画像</span></div>
          <div class="login-feature"><div class="login-feature-dot" /><span>AI 智能对话</span></div>
          <div class="login-feature"><div class="login-feature-dot" /><span>四路径发展引导</span></div>
          <div class="login-feature"><div class="login-feature-dot" /><span>三级预警体系</span></div>
        </div>
      </div>

      <div class="login-brand-footer">ZJNU Academic Navigation v2.0</div>
    </div>

    <!-- Right: Login Form -->
    <div class="login-form-side">
      <div class="login-card">
        <div class="login-card-header">
          <div class="login-card-header-brand">
            <img src="/zjnu-logo.png" alt="浙江师范大学校徽" class="login-card-logo" />
            <div>
              <h2 class="login-card-title">欢迎回来</h2>
              <p class="login-card-desc">登录以访问你的学业导航面板</p>
            </div>
          </div>
        </div>

        <div class="login-fields">
          <div class="login-field" :class="{ 'has-error': fieldErrors.account }">
            <label class="login-label" for="login-account">工号 / 学号</label>
            <el-input id="login-account" v-model="form.account" autocomplete="username" placeholder="请输入工号或学号" size="large" :aria-invalid="!!fieldErrors.account" @input="fieldErrors.account = ''" @blur="validateForm" @keyup.enter="handleLogin" />
            <span v-if="fieldErrors.account" class="field-error" role="alert">{{ fieldErrors.account }}</span>
          </div>
          <div class="login-field" :class="{ 'has-error': fieldErrors.password }">
            <label class="login-label" for="login-password">密码</label>
            <el-input id="login-password" v-model="form.password" autocomplete="current-password" type="password" placeholder="请输入密码" size="large" show-password :aria-invalid="!!fieldErrors.password" @input="fieldErrors.password = ''" @blur="validateForm" @keyup.enter="handleLogin" />
            <span v-if="fieldErrors.password" class="field-error" role="alert">{{ fieldErrors.password }}</span>
          </div>
        </div>

        <button class="login-submit" :disabled="loading" :aria-busy="loading" @click="handleLogin">
          <template v-if="!loading">登 录</template>
          <template v-else><span class="login-dots" aria-label="正在登录"><i /><i /><i /></span></template>
        </button>

        <div class="login-hint">
          <span class="login-hint-tag">DEMO</span>
          <span>系统根据工号前缀自动识别身份，密码统一：123456</span>
        </div>
        <div class="login-accounts">
            <button
            v-for="item in demoAccounts" :key="item.account"
            class="login-account-item"
            type="button"
            @click="fillDemo(item.account)"
          >{{ item.label }}：{{ item.account }}</button>
        </div>

        <div class="login-brand-line">
          <img src="/zjnu-logo.png" alt="浙江师范大学校徽" class="login-brand-line-logo" />
          <span>浙江师范大学 · 学业领航智能体</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-root { min-height: 100vh; display: flex; }

.login-brand {
  flex: 0 0 46%; background: #0a1329; position: relative; overflow: hidden;
  display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 48px;
}
.login-grid { position: absolute; inset: 0; background-image: linear-gradient(rgba(255,255,255,0.025) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.025) 1px, transparent 1px); background-size: 44px 44px; }
.login-glow-1 { position: absolute; top: -15%; left: -10%; width: 500px; height: 500px; border-radius: 50%; background: radial-gradient(ellipse at center, rgba(42,77,153,0.3) 0%, transparent 70%); pointer-events: none; }
.login-glow-2 { position: absolute; bottom: -10%; right: -5%; width: 380px; height: 380px; border-radius: 50%; background: radial-gradient(ellipse at center, rgba(61,184,126,0.12) 0%, transparent 70%); pointer-events: none; }
.login-brand-content { position: relative; z-index: 1; text-align: center; }
.login-emblem { width: 80px; height: 80px; margin: 0 auto 24px; border-radius: 50%; display: flex; align-items: center; justify-content: center; background: rgba(255,255,255,0.06); box-shadow: 0 0 0 1px rgba(255,255,255,0.1), 0 8px 32px rgba(0,0,0,0.3); }
.login-emblem-img { width: 68px; height: 68px; object-fit: contain; border-radius: 50%; max-width: 100%; }
.login-brand-title { color: #ffffff; font-size: 26px; font-weight: 800; letter-spacing: 0.04em; margin: 0 0 8px; line-height: 1.2; }
.login-brand-sub { color: rgba(180,200,230,0.5); font-size: 13px; letter-spacing: 0.06em; margin: 0 0 48px; }
.login-brand-features { display: grid; grid-template-columns: 1fr 1fr; gap: 12px 32px; text-align: left; }
.login-feature { display: flex; align-items: center; gap: 10px; font-size: 13px; color: rgba(180,200,230,0.55); letter-spacing: 0.02em; }
.login-feature-dot { width: 6px; height: 6px; border-radius: 50%; background: #5580d4; flex-shrink: 0; box-shadow: 0 0 8px rgba(85,128,212,0.5); }
.login-brand-footer { position: absolute; bottom: 24px; font-size: 11px; color: rgba(140,160,200,0.2); letter-spacing: 0.06em; }

.login-form-side { flex: 1; display: flex; align-items: center; justify-content: center; padding: 40px; background: #f6f8fb; }
.login-card { width: 100%; max-width: 400px; animation: card-in 0.45s cubic-bezier(0.22,1,0.36,1) both; }
.login-card-header { margin-bottom: 28px; }
.login-card-title { font-size: 24px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.login-card-desc { font-size: 13.5px; color: #8a99b4; margin: 6px 0 0; }

.login-fields { display: flex; flex-direction: column; gap: 18px; margin-bottom: 24px; }
.login-field { display: flex; flex-direction: column; gap: 6px; }
.field-error { color: #c93b2f; font-size: 12px; line-height: 1.3; }
.login-field.has-error :deep(.el-input__wrapper) { box-shadow: 0 0 0 1px #e04538 inset !important; }
.login-label { font-size: 12.5px; font-weight: 700; color: #3e4759; letter-spacing: 0.04em; }

.login-submit {
  width: 100%; height: 46px;
  background: linear-gradient(135deg, #2A4D99 0%, #1d3670 100%);
  color: #fff; border: none; border-radius: 12px;
  font-size: 15px; font-weight: 700; letter-spacing: 0.1em;
  cursor: pointer; transition: all 0.2s ease; font-family: inherit;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 4px 16px rgba(42,77,153,0.3); margin-bottom: 20px;
}
.login-submit:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 6px 24px rgba(42,77,153,0.4); background: linear-gradient(135deg, #3360b0 0%, #244389 100%); }
.login-submit:active:not(:disabled) { transform: translateY(0); box-shadow: 0 2px 8px rgba(42,77,153,0.3); }
.login-submit:disabled { opacity: 0.7; cursor: not-allowed; }

.login-dots { display: flex; gap: 6px; align-items: center; }
.login-dots i { width: 6px; height: 6px; border-radius: 50%; background: rgba(255,255,255,0.9); animation: typing-dot 1.2s infinite; display: block; }
.login-dots i:nth-child(2) { animation-delay: 0.2s; }
.login-dots i:nth-child(3) { animation-delay: 0.4s; }

.login-hint { display: flex; align-items: center; gap: 8px; font-size: 11.5px; color: #9facc5; justify-content: center; }
.login-hint-tag { font-size: 10px; font-weight: 800; letter-spacing: 0.08em; color: #b4bed2; background: #edf0f6; padding: 2px 6px; border-radius: 4px; }
.login-accounts { display: flex; gap: 12px; flex-wrap: wrap; justify-content: center; font-size: 11px; color: #b4bed2; margin-top: 6px; margin-bottom: 16px; }
.login-account-item { white-space: nowrap; cursor: pointer; transition: color 0.15s; }
.login-account-item:hover { color: #2A4D99; }
.login-account-item:focus-visible { outline: 2px solid #5580d4; outline-offset: 3px; border-radius: 3px; }

.login-brand-line { display: flex; align-items: center; justify-content: center; gap: 7px; font-size: 11.5px; color: #5aa96b; font-weight: 500; letter-spacing: 0.03em; }
.login-brand-line-logo { width: 16px; height: 16px; object-fit: contain; max-width: 100%; }

.login-card-header-brand { display: flex; align-items: center; gap: 12px; margin-bottom: 24px; }
.login-card-logo { width: 36px; height: 36px; object-fit: contain; flex-shrink: 0; max-width: 100%; }

.login-motto-watermark { position: absolute; bottom: 48px; left: 50%; transform: translateX(-50%); width: 160px; opacity: 0.18; pointer-events: none; z-index: 0; mix-blend-mode: screen; }
.login-motto-watermark img { width: 100%; max-width: 100%; object-fit: contain; filter: invert(1) grayscale(100%) brightness(1.4); }

@media (max-width: 900px) {
  .login-root { flex-direction: column; }
  .login-brand { flex: 0 0 auto; padding: 40px 24px; min-height: 260px; }
  .login-brand-features { display: none; }
  .login-form-side { padding: 32px 20px; }
}

@media (max-width: 560px) {
  .login-brand { min-height: 220px; padding: 28px 20px 32px; }
  .login-emblem { width: 64px; height: 64px; margin-bottom: 16px; }
  .login-emblem-img { width: 54px; height: 54px; }
  .login-brand-title { font-size: 22px; }
  .login-brand-sub { margin-bottom: 0; font-size: 12px; }
  .login-brand-footer { bottom: 12px; font-size: 10px; }
  .login-form-side { align-items: flex-start; padding: 28px 18px 36px; }
  .login-card-header { margin-bottom: 22px; }
  .login-card-title { font-size: 21px; }
  .login-fields { gap: 14px; }
  .login-submit { height: 48px; }
  .login-hint { align-items: flex-start; text-align: left; line-height: 1.5; }
  .login-accounts { gap: 8px 12px; }
  .login-account-item { padding: 2px 0; border: 0; background: transparent; color: #7b879d; font: inherit; font-size: 11px; }
}
</style>
