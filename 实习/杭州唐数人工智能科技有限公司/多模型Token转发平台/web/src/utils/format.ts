/* 展示与提交格式化：金额一律整数运算，避免浮点误差 */

/** micro-currency（1 元 = 1,000,000）→ 元字符串，保留两位小数 */
export function formatMicro(micro: number | string | null | undefined): string {
  if (micro === null || micro === undefined || micro === '') return '-'
  const n = Number(micro)
  if (!Number.isFinite(n)) return '-'
  return (n / 1e6).toFixed(2)
}

/** 元（字符串或数字，如 "100.50"）→ micro-currency 整数；非法输入返回 null */
export function toMicro(yuan: string | number): number | null {
  const n = typeof yuan === 'number' ? yuan : Number(yuan)
  if (!Number.isFinite(n) || n < 0) return null
  return Math.round(n * 1e6)
}

/** micro-currency → 元字符串，最多 4 位小数并去尾零（模型单价等小额场景） */
export function formatMicroTrim(micro: number | string | null | undefined): string {
  if (micro === null || micro === undefined || micro === '') return '-'
  const n = Number(micro)
  if (!Number.isFinite(n)) return '-'
  return (n / 1e6).toFixed(4).replace(/0+$/, '').replace(/\.$/, '')
}

/** token 计数等原始大整数展示 */
export function formatCount(n: number | string | null | undefined): string {
  if (n === null || n === undefined || n === '') return '-'
  const num = Number(n)
  if (!Number.isFinite(num)) return '-'
  return num.toLocaleString('zh-CN')
}

const pad = (n: number) => String(n).padStart(2, '0')

/** RFC3339 → 'YYYY-MM-DD HH:mm:ss'（本地时区） */
export function formatTime(iso: string | null | undefined): string {
  if (!iso) return '-'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '-'
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
    `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  )
}
