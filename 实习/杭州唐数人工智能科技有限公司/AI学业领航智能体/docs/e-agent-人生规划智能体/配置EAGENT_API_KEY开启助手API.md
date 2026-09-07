# 配置 EAGENT_API_KEY 开启 E-Agent 助手 API

> 目标：让 E-Agent 的 `assistant/chat/completions` 接口可用，从而像 Dify 一样用 API Key 直连助手。

## 重要前提

`EAGENT_API_KEY` 必须配置在远端 E-Agent 后端服务器上，不是配置在 Windows 项目或前端代码里。前端只是填同一个 Key 去调用接口。

## 1. 登录 E-Agent 服务器

Windows 本机执行：

```powershell
ssh -p 6000 -i C:\Users\BenBen\.ssh\id_ed25519 root@120.55.180.232
```

跳到 E-Agent 主机：

```bash
ssh root@172.20.20.35
```

## 2. 确认 E-Agent 运行方式

```bash
docker ps | grep -i eagent
systemctl list-units --type=service | grep -i eagent
ps aux | grep -i eagent | grep -v grep
```

通常只会命中其中一种，决定下一步 Key 写在哪里。

## 3. 生成强随机 Key

```bash
openssl rand -hex 32
```

记下输出，例如：

```text
d23fef00024943cbba8f35e1fa32b9ff60b48a596e9c4801bb9df623e00e5c65
```

## 4. 写入环境变量

### Docker Compose

```bash
docker compose ps
docker inspect <eagent容器名> | grep -i -A5 Env
```

编辑 `docker-compose.yml`，给 E-Agent 服务增加：

```yaml
environment:
  - EAGENT_API_KEY=d23fef...
```

重启：

```bash
docker compose up -d
```

### Docker run

把原命令重跑，并增加：

```bash
-e EAGENT_API_KEY=d23fef...
```

### systemd

```bash
systemctl cat eagent
```

编辑 unit 文件，增加：

```ini
Environment=EAGENT_API_KEY=d23fef...
```

重载并重启：

```bash
systemctl daemon-reload
systemctl restart eagent
```

### 裸进程或启动脚本

在启动脚本里增加：

```bash
export EAGENT_API_KEY=d23fef...
```

然后重启对应进程。

## 5. 验证接口

```bash
curl -sS http://127.0.0.1:3001/api/v2/assistant/chat/completions \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: d23fef...' \
  -d '{"model":"IWhc8a17ZoCo8wzAIxqXEj5kC0_A_Ezf","messages":[{"role":"user","content":"你好"}],"stream":false}'
```

- 返回 `API v2 not configured`：环境变量未被后端读取，回到第 2 步确认运行方式。
- 返回正常 JSON：接口已开启。

## 6. 回填前端

打开 `academic-navigation-demo/src/mock/agents.ts`，将 `AI 人生规划` 的 `externalApi` 改为：

```ts
externalApi: {
  type: 'eagent',
  baseUrl: 'http://localhost:3001',
  assistantId: 'IWhc8a17ZoCo8wzAIxqXEj5kC0_A_Ezf',
  apiKey: 'd23fef...',
}
```

然后重新构建前端。

## 安全提醒

不要把 Key 长期写在浏览器前端代码里。正式环境应由学校后端代理 E-Agent，前端只带学生 JWT，后端负责：

1. 校验学生身份。
2. 读取本人学业数据。
3. 注入学生上下文。
4. 携带 `X-API-Key` 调用 E-Agent。
