# 部署手册

本文面向运维人员，说明 school-api-v1 的部署要求、步骤与生产注意事项。开发者本地运行请参考 [README.md](../README.md)。

## 一、部署要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Docker / Docker Compose | 兼容 compose 文件格式即可 | 推荐部署方式 |
| PostgreSQL | 15+ | 也可使用外部实例 |
| Redis | 7+ | 用于登录限流等 |
| 反向代理 | 任意 | 生产环境建议位于服务前层，负责 TLS |

服务自身为单个无状态容器，监听 `8080`，水平扩展时多个副本共享同一 PostgreSQL 与 Redis。

## 二、部署架构

`docker-compose.yml` 包含四个服务：

| 服务 | 说明 |
|------|------|
| `postgres` | PostgreSQL 15，数据卷 `postgres_data` |
| `redis` | Redis 7，数据卷 `redis_data` |
| `migrate` | 一次性任务，数据库就绪后执行 `migrate up` 应用全部迁移 |
| `backend` | 后端服务，依赖 `migrate` 成功完成后启动 |

迁移文件已嵌入镜像，迁移容器无需挂载宿主机目录。

## 三、环境变量

所有配置通过 `APP_` 前缀环境变量注入（详见 README「配置说明」）。生产部署必须显式设置以下变量：

| 变量 | 说明 |
|------|------|
| `APP_SERVER__ENV` | 必须设为 `production` |
| `APP_SERVER__LOG_LEVEL` | 建议 `info` |
| `APP_DATABASE__URL` | PostgreSQL 连接串 |
| `APP_REDIS__URL` | Redis 连接串 |
| `APP_JWT__SECRET` | 管理员 JWT 密钥，强随机值 |
| `APP_ENCRYPTION__KEY` | 上游账号 API Key 加密密钥，base64url 编码的 32 字节 |

密钥生成：

```bash
openssl rand -base64 32
```

注意：`APP_ENCRYPTION__KEY` 一旦投入使用即与数据库中已加密的上游账号绑定，丢失后无法解密存量 API Key；泄露则等于泄露全部上游账号，须按最高级别机密管理。

### 3.1 可选配置：调用日志保留期（需求 #7）

| 变量 | 默认 | 说明 |
| ---- | ---- | ---- |
| `APP_RETENTION__CALL_LOG_DAYS` | `30` | 调用日志保留天数，到期后由后台任务删除；设为 `0` 或负数表示不清理 |
| `APP_RETENTION__SWEEP_INTERVAL` | `24h` | 清理任务的执行间隔（Go duration，如 `12h`、`30m`） |

后台任务在服务启动时立即执行一次，随后按间隔循环；也可通过 `POST /api/v1/admin/maintenance/call-logs/sweep` 手动触发一次清理（用于部署后立即验收）。

## 四、首次部署步骤

1. 准备 `.env` 文件，设置上述环境变量。
2. 构建并启动：

```bash
docker compose up --build -d
```

3. 确认迁移成功（`migrate` 容器退出码为 0）：

```bash
docker compose ps -a migrate
```

4. 健康检查：

```bash
curl http://localhost:8080/health
# 期望返回 200，且 postgres / redis 均为健康
```

5. 创建首个管理员（`cmd/admin` 已随镜像构建，也可在宿主机执行 `go run ./cmd/admin`）：

```bash
docker compose exec backend /app/admin create-admin -username=<管理员账号> -password=<强密码> -role=super_admin
```

角色取值为 `super_admin` 或 `admin`。三个参数也可通过环境变量 `APP_ADMIN__SEED_USERNAME`、`APP_ADMIN__SEED_PASSWORD`、`APP_ADMIN__SEED_ROLE` 提供。在宿主机直接运行时，`config.Validate()` 会强制校验完整配置，至少需要 `APP_DATABASE__URL`、`APP_REDIS__URL`、`APP_JWT__SECRET`、`APP_ENCRYPTION__KEY` 均有效。

6. 登录管理端，依次配置：平台 → 上游账号（填写上游 API Key，服务端加密存储）→ AI 模型 → 分组绑定。

## 五、迁移与回滚

镜像内自带 `migrate` 命令，用法与 `cmd/migrate` 一致：

```bash
docker compose run --rm migrate up        # 应用全部迁移
docker compose run --rm migrate down      # 回滚一个版本
docker compose run --rm migrate down 3    # 回滚三个版本
docker compose run --rm migrate version   # 当前版本
```

升级流程：拉取新镜像 → `docker compose up --build -d`（migrate 自动执行新迁移）→ 观察日志确认迁移成功 → 确认 `/health` 正常。

回滚流程：先 `migrate down` 回滚数据库，再切回旧镜像；每个迁移均有对应 `.down.sql`。

## 六、生产检查清单

- [ ] `APP_SERVER__ENV=production`
- [ ] `APP_JWT__SECRET` 与 `APP_ENCRYPTION__KEY` 为强随机值，且未提交到仓库
- [ ] `.env` 文件权限收紧（如 `chmod 600`）
- [ ] 反向代理已启用 TLS 并转发 `X-Forwarded-For`（服务在 production 下仅信任 RFC1918 代理段，防止 ClientIP 伪造）
- [ ] PostgreSQL 与 Redis 不暴露公网；若使用外部实例，确认网络隔离与认证
- [ ] 数据库已配置定期备份（见下节）
- [ ] 日志采集：服务输出 JSON 结构化日志（stdout），由容器平台采集
- [ ] 确认登录限流生效（管理端登录接口按 IP 限流，依赖 Redis）

## 七、备份

需要备份的数据：

| 数据 | 方式 |
|------|------|
| PostgreSQL | `pg_dump` 定时任务；`users`（余额）、`balance_records`、`quota_records`、`call_logs` 属于资金与审计数据，不可丢失 |
| 加密密钥 | `APP_ENCRYPTION__KEY` 须离线备份；无此密钥无法解密上游账号 |
| 环境配置 | `.env` 纳入受控备份 |

## 八、故障排查

| 现象 | 排查方向 |
|------|----------|
| `backend` 反复重启 | `docker compose logs migrate` 查看迁移是否失败；`docker compose logs backend` 查看连接数据库/Redis 是否就绪 |
| `/health` 显示数据库异常 | 检查 `APP_DATABASE__URL`、网络连通性、数据库实例状态 |
| 管理端登录 429 | 登录限流触发，确认 Redis 正常；同一出口 IP 的频繁失败尝试会被限流 |
| 转发 402 | 用户余额或配额不足，属业务预期；检查 `balance_records` / `quota_records` 流水 |
| 上游全部失败 | 检查上游账号状态（连续错误会被标记 `error`）与平台配置；查看 `call_logs` 的 `error_msg` |
| 迁移版本不一致 | `docker compose run --rm migrate version` 对比镜像内迁移文件数量 |

## 九、常用运维命令

```bash
docker compose logs -f backend    # 跟踪服务日志
docker compose up -d --build      # 重建并滚动重启
docker compose down               # 停止（保留数据卷）
docker compose down -v            # 停止并删除数据卷（危险，仅测试环境）
```
