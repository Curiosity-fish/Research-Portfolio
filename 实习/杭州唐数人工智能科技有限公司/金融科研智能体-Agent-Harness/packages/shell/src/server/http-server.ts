import { randomUUID } from "node:crypto";
import { readFile } from "node:fs/promises";
import { createServer, type Server } from "node:http";
import { extname, join, resolve } from "node:path";
import { TfaServerError } from "@earendil-works/tfa-server";
import { Hono } from "hono";
import { AuthError, authenticateRequest } from "./auth.ts";
import type { GroupType, UserRole, UserStatus } from "./db/types.ts";
import { QuotaExceededError } from "./errors.ts";
import type { ShellSessionService } from "./session-service.ts";
import { isVisibility } from "./skills.ts";

export interface HttpServerOptions {
	/** 会话服务 */
	service: ShellSessionService;
	/** 监听端口,默认 8787 */
	port?: number;
	/** 监听地址,默认 127.0.0.1 */
	host?: string;
	/** 前端构建产物目录,默认 shell/web/dist */
	staticDir?: string;
}

export interface HttpServerHandle {
	url: string;
	shutdown(): Promise<void>;
}

const MIME_TYPES: Record<string, string> = {
	".html": "text/html; charset=utf-8",
	".js": "text/javascript; charset=utf-8",
	".css": "text/css; charset=utf-8",
	".json": "application/json; charset=utf-8",
	".svg": "image/svg+xml",
	".png": "image/png",
	".jpg": "image/jpeg",
	".ico": "image/x-icon",
	".woff2": "font/woff2",
	".map": "application/json",
};

const SSE_KEEPALIVE_MS = 15_000;

function isTfaServerError(error: unknown): error is TfaServerError {
	return error instanceof TfaServerError;
}

function toErrorBody(error: unknown): { status: number; body: { error: { code: string; message: string } } } {
	if (isTfaServerError(error)) {
		const status = error.code === "busy" ? 409 : error.code === "not_found" ? 404 : 400;
		return { status, body: { error: { code: error.code, message: error.message } } };
	}
	const message = error instanceof Error ? error.message : String(error);
	return { status: 500, body: { error: { code: "internal_error", message } } };
}

function readJsonBody(c: { req: { json(): Promise<unknown> } }): Promise<Record<string, unknown>> {
	return c.req.json().catch(() => ({})) as Promise<Record<string, unknown>>;
}

/** 把会话事件序列化为 SSE 帧。 */
function serializeSseEvent(event: { type: string; [key: string]: unknown }): string {
	return `event: ${event.type}\ndata: ${JSON.stringify(event)}\n\n`;
}

/**
 * 构建 hono 应用:REST + SSE + 静态托管。
 */
export function createApp(
	service: ShellSessionService,
	options: Omit<HttpServerOptions, "service"> | undefined = undefined,
): Hono {
	const app = new Hono();
	const staticDir = options?.staticDir ? resolve(options.staticDir) : resolve(import.meta.dirname, "../../web/dist");

	app.get("/api/health", (c) => c.json({ ok: true }));
	app.post("/api/auth/login", async (c) => {
		const body = await readJsonBody(c);
		const username = typeof body.username === "string" ? body.username : "";
		const password = typeof body.password === "string" ? body.password : "";
		if (!username || !password) {
			return c.json({ error: { code: "invalid_request", message: "需要 username 和 password" } }, 400);
		}
		const session = await service.login(username, password);
		return c.json({ token: session.token, user: session.user, expiresAt: session.expiresAt });
	});

	app.post("/api/auth/logout", async (c) => {
		await authenticateRequest(c, service);
		const header = c.req.header("authorization") ?? "";
		await service.logout(header.slice("Bearer ".length));
		return c.json({ ok: true });
	});

	app.get("/api/auth/me", async (c) => {
		const user = await authenticateRequest(c, service);
		return c.json({ user });
	});

	app.get("/api/admin/users", async (c) => {
		const user = await authenticateRequest(c, service);
		if (user.role !== "admin") throw new AuthError("forbidden", "需要管理员权限");
		const users = await service.listUsers();
		return c.json({ users });
	});

	app.post("/api/admin/users", async (c) => {
		const user = await authenticateRequest(c, service);
		if (user.role !== "admin") throw new AuthError("forbidden", "需要管理员权限");
		const body = await readJsonBody(c);
		if (typeof body.username !== "string" || typeof body.password !== "string" || typeof body.role !== "string") {
			return c.json({ error: { code: "invalid_request", message: "需要 username/password/role" } }, 400);
		}
		const created = await service.createUser({
			username: body.username,
			password: body.password,
			role: body.role as UserRole,
		});
		return c.json({ user: created }, 201);
	});

	app.patch("/api/admin/users/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		if (user.role !== "admin") throw new AuthError("forbidden", "需要管理员权限");
		const body = await readJsonBody(c);
		const id = c.req.param("id");
		if (body.status !== undefined) {
			if (body.status !== "active" && body.status !== "disabled") {
				return c.json({ error: { code: "invalid_request", message: "status 需为 active 或 disabled" } }, 400);
			}
			await service.setUserStatus(id, body.status as UserStatus);
		}
		if (body.password !== undefined) {
			if (typeof body.password !== "string") {
				return c.json({ error: { code: "invalid_request", message: "password 需为字符串" } }, 400);
			}
			await service.resetPassword(id, body.password);
		}
		return c.json({ ok: true });
	});

	// ---- 会话监控 (T11, 管理端) ----

	app.get("/api/admin/sessions", async (c) => {
		const user = await authenticateRequest(c, service);
		if (user.role !== "admin") throw new AuthError("forbidden", "需要管理员权限");
		const sessions = await service.listAdminSessions();
		return c.json({ sessions });
	});

	app.post("/api/admin/sessions/:id/terminate", async (c) => {
		const user = await authenticateRequest(c, service);
		if (user.role !== "admin") throw new AuthError("forbidden", "需要管理员权限");
		await service.terminateAdminSession(c.req.param("id"));
		return c.json({ ok: true });
	});

	// ---- 课程/课题组 (FR-2) ----

	app.get("/api/groups", async (c) => {
		const user = await authenticateRequest(c, service);
		return c.json({ groups: service.groups.listGroups(user) });
	});

	app.post("/api/groups", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (typeof body.name !== "string" || typeof body.type !== "string") {
			return c.json({ error: { code: "invalid_request", message: "需要 name 和 type" } }, 400);
		}
		const group = service.groups.createGroup(user, { name: body.name, type: body.type as GroupType });
		return c.json({ group }, 201);
	});

	app.get("/api/groups/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		const group = service.groups.getGroup(user, c.req.param("id"));
		return c.json({ group });
	});

	app.patch("/api/groups/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		const patch: { name?: string } = {};
		if (body.name !== undefined) {
			if (typeof body.name !== "string") {
				return c.json({ error: { code: "invalid_request", message: "name 需为字符串" } }, 400);
			}
			patch.name = body.name;
		}
		const group = service.groups.updateGroup(user, c.req.param("id"), patch);
		return c.json({ group });
	});

	app.delete("/api/groups/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		const group = service.groups.archiveGroup(user, c.req.param("id"));
		return c.json({ group });
	});

	app.post("/api/groups/:id/members", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (typeof body.userId !== "string") {
			return c.json({ error: { code: "invalid_request", message: "需要 userId" } }, 400);
		}
		service.groups.addMember(user, c.req.param("id"), { userId: body.userId });
		return c.json({ ok: true }, 201);
	});

	app.patch("/api/groups/:id/members/:userId", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (body.role !== "owner" && body.role !== "member") {
			return c.json({ error: { code: "invalid_request", message: "role 需为 owner 或 member" } }, 400);
		}
		service.groups.setMemberRole(user, c.req.param("id"), c.req.param("userId"), body.role);
		return c.json({ ok: true });
	});

	app.delete("/api/groups/:id/members/:userId", async (c) => {
		const user = await authenticateRequest(c, service);
		service.groups.removeMember(user, c.req.param("id"), c.req.param("userId"));
		return c.json({ ok: true });
	});

	// ---- 技能库 (FR-4) ----

	app.get("/api/skills", async (c) => {
		const user = await authenticateRequest(c, service);
		return c.json({ skills: service.skills.listSkills(user) });
	});

	app.post("/api/skills", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (typeof body.visibility !== "string" || !isVisibility(body.visibility)) {
			return c.json({ error: { code: "invalid_request", message: "visibility 必须是 private/group/public" } }, 400);
		}
		const skill = await service.skills.createSkill(user, {
			name: typeof body.name === "string" ? body.name : undefined,
			repoUrl: typeof body.repoUrl === "string" ? body.repoUrl : undefined,
			visibility: body.visibility as "private" | "group" | "public",
			groupId: typeof body.groupId === "string" ? body.groupId : undefined,
			...(typeof body.upload === "boolean" ? { upload: body.upload } : {}),
		});
		return c.json({ skill }, 201);
	});

	app.post("/api/skills/:id/update", async (c) => {
		const user = await authenticateRequest(c, service);
		const skill = await service.skills.updateSkillFromRepo(user, c.req.param("id"));
		return c.json({ skill });
	});

	app.patch("/api/skills/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		const patch: { name?: string; visibility?: "private" | "group" | "public"; groupId?: string } = {};
		if (body.name !== undefined) {
			if (typeof body.name !== "string") {
				return c.json({ error: { code: "invalid_request", message: "name 需为字符串" } }, 400);
			}
			patch.name = body.name;
		}
		if (body.visibility !== undefined) {
			if (typeof body.visibility !== "string" || !isVisibility(body.visibility)) {
				return c.json(
					{ error: { code: "invalid_request", message: "visibility 必须是 private/group/public" } },
					400,
				);
			}
			patch.visibility = body.visibility;
		}
		if (body.groupId !== undefined) {
			if (body.groupId !== null && typeof body.groupId !== "string") {
				return c.json({ error: { code: "invalid_request", message: "groupId 需为字符串或 null" } }, 400);
			}
			if (typeof body.groupId === "string") patch.groupId = body.groupId;
		}
		const skill = service.skills.updateSkill(user, c.req.param("id"), patch);
		return c.json({ skill });
	});

	app.delete("/api/skills/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		service.skills.deleteSkill(user, c.req.param("id"));
		return c.json({ ok: true });
	});

	app.post("/api/admin/skills/:id/review", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (body.reviewState !== "approved" && body.reviewState !== "rejected") {
			return c.json({ error: { code: "invalid_request", message: "reviewState 必须是 approved 或 rejected" } }, 400);
		}
		const skill = service.skills.reviewSkill(user, c.req.param("id"), body.reviewState);
		return c.json({ skill });
	});

	// ---- 数据集库 (FR-5) ----

	app.post("/api/admin/datasets/:id/review", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (body.reviewState !== "approved" && body.reviewState !== "rejected") {
			return c.json({ error: { code: "invalid_request", message: "reviewState 必须是 approved 或 rejected" } }, 400);
		}
		const dataset = service.datasets.reviewDataset(user, c.req.param("id"), body.reviewState);
		return c.json({ dataset });
	});

	app.get("/api/datasets", async (c) => {
		const user = await authenticateRequest(c, service);
		return c.json({ datasets: service.datasets.listDatasets(user) });
	});

	app.post("/api/datasets", async (c) => {
		const user = await authenticateRequest(c, service);
		let formData: FormData;
		try {
			formData = await c.req.formData();
		} catch {
			return c.json({ error: { code: "invalid_request", message: "需要 multipart/form-data" } }, 400);
		}
		const file = formData.get("file");
		if (!(file instanceof File)) {
			return c.json({ error: { code: "invalid_request", message: "需要 file 字段" } }, 400);
		}
		const visibility = formData.get("visibility");
		if (typeof visibility !== "string" || !isVisibility(visibility)) {
			return c.json({ error: { code: "invalid_request", message: "visibility 必须是 private/group/public" } }, 400);
		}
		const name = formData.get("name");
		const groupId = formData.get("groupId");
		const dataset = await service.datasets.createDataset(user, {
			file,
			name: typeof name === "string" ? name : undefined,
			visibility,
			groupId: typeof groupId === "string" ? groupId : undefined,
		});
		return c.json({ dataset }, 201);
	});

	app.patch("/api/datasets/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		const patch: { name?: string; visibility?: "private" | "group" | "public"; groupId?: string } = {};
		if (body.name !== undefined) {
			if (typeof body.name !== "string") {
				return c.json({ error: { code: "invalid_request", message: "name 需为字符串" } }, 400);
			}
			patch.name = body.name;
		}
		if (body.visibility !== undefined) {
			if (typeof body.visibility !== "string" || !isVisibility(body.visibility)) {
				return c.json(
					{ error: { code: "invalid_request", message: "visibility 必须是 private/group/public" } },
					400,
				);
			}
			patch.visibility = body.visibility;
		}
		if (body.groupId !== undefined) {
			if (body.groupId !== null && typeof body.groupId !== "string") {
				return c.json({ error: { code: "invalid_request", message: "groupId 需为字符串或 null" } }, 400);
			}
			if (typeof body.groupId === "string") patch.groupId = body.groupId;
		}
		const dataset = service.datasets.updateDataset(user, c.req.param("id"), patch);
		return c.json({ dataset });
	});

	app.delete("/api/datasets/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		service.datasets.deleteDataset(user, c.req.param("id"));
		return c.json({ ok: true });
	});

	app.get("/api/datasets/:id/download", async (c) => {
		const user = await authenticateRequest(c, service);
		const info = service.datasets.getDownload(user, c.req.param("id"));
		if (!info) {
			return c.json({ error: { code: "not_found", message: "数据集不存在或不可见" } }, 404);
		}
		const content = await readFile(info.path);
		const encoded = encodeURIComponent(info.name);
		return c.body(new Uint8Array(content), {
			status: 200,
			headers: {
				"Content-Type": "application/octet-stream",
				"Content-Disposition": `attachment; filename*=UTF-8'${encoded}`,
			},
		});
	});

	app.get("/api/sessions", async (c) => {
		const sessions = await service.listSessions();
		sessions.sort((a, b) => b.updatedAt - a.updatedAt);
		return c.json({ sessions });
	});

	app.post("/api/sessions", async (c) => {
		const body = await readJsonBody(c);
		// 带 token 时记录会话归属(不强制登录,兼容旧流程);权限模式由服务端按项目会话自动决定
		const sessionUser = await authenticateRequest(c, service).catch(() => undefined);
		const runtime = await service.createSession({
			id: randomUUID(),
			cwd: typeof body.cwd === "string" ? body.cwd : undefined,
			name: typeof body.name === "string" ? body.name : undefined,
			projectId: typeof body.projectId === "string" ? body.projectId : undefined,
			...(sessionUser ? { userId: sessionUser.id } : {}),
			...(isModelRef(body.model) ? { model: body.model } : {}),
			...(isThinkingLevel(body.thinkingLevel) ? { thinkingLevel: body.thinkingLevel } : {}),
		});
		return c.json({ session: runtime.snapshot() }, 201);
	});

	// ---- 会话权限模式与请求批准 (T14) ----
	app.get("/api/sessions/:id/permission-mode", async (c) => {
		const user = await authenticateRequest(c, service);
		const mode = service.getSessionPermissionMode(user, c.req.param("id"));
		return c.json({ mode });
	});

	app.post("/api/sessions/:id/permission-mode", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (body.mode !== "request_approval" && body.mode !== "full_access") {
			return c.json(
				{ error: { code: "invalid_request", message: "mode 必须是 request_approval 或 full_access" } },
				400,
			);
		}
		const mode = service.setSessionPermissionMode(user, c.req.param("id"), body.mode);
		return c.json({ mode });
	});

	app.get("/api/sessions/:id/approvals", async (c) => {
		const user = await authenticateRequest(c, service);
		const approvals = service.listSessionApprovals(user, c.req.param("id"));
		return c.json({ approvals });
	});

	app.post("/api/sessions/:id/approvals/:approvalId/approve", async (c) => {
		const user = await authenticateRequest(c, service);
		const approval = await service.decideApproval(user, c.req.param("id"), c.req.param("approvalId"), true);
		return c.json({ approval });
	});

	app.post("/api/sessions/:id/approvals/:approvalId/reject", async (c) => {
		const user = await authenticateRequest(c, service);
		const approval = await service.decideApproval(user, c.req.param("id"), c.req.param("approvalId"), false);
		return c.json({ approval });
	});
	app.get("/api/models", async (c) => {
		const models = await service.listModels();
		return c.json({ models });
	});

	app.get("/api/config", async (c) => {
		const config = await service.getConfig();
		return c.json({ config });
	});

	app.get("/api/tools", async (c) => {
		const tools = await service.listTools();
		return c.json({ tools });
	});

	app.get("/api/projects", async (c) => {
		const user = await authenticateRequest(c, service);
		const projects = await service.listProjects(user);
		return c.json({ projects });
	});

	app.post("/api/projects", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (typeof body.name !== "string" || !body.name.trim()) {
			return c.json({ error: { code: "invalid_request", message: "需要 name" } }, 400);
		}
		let scope: { type: "private" } | { type: "group"; groupId: string } | undefined;
		if (body.scope !== undefined) {
			if (!isProjectScope(body.scope)) {
				return c.json(
					{
						error: {
							code: "invalid_request",
							message: "scope 需为 { type: 'private' } 或 { type: 'group', groupId }",
						},
					},
					400,
				);
			}
			scope = body.scope;
		}
		const project = await service.createProject(user, {
			name: body.name,
			folder: typeof body.folder === "string" ? body.folder : undefined,
			...(scope ? { scope } : {}),
		});
		return c.json({ project }, 201);
	});

	app.patch("/api/projects/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (body.scope !== undefined) {
			return c.json({ error: { code: "invalid_request", message: "项目作用域不可修改" } }, 400);
		}
		const patch: { name?: string; folder?: string } = {};
		if (body.name !== undefined) {
			if (typeof body.name !== "string") {
				return c.json({ error: { code: "invalid_request", message: "name 需为字符串" } }, 400);
			}
			patch.name = body.name;
		}
		if (body.folder !== undefined) {
			if (typeof body.folder !== "string") {
				return c.json({ error: { code: "invalid_request", message: "folder 需为字符串" } }, 400);
			}
			patch.folder = body.folder;
		}
		const project = await service.updateProject(user, c.req.param("id"), patch);
		return c.json({ project });
	});

	app.delete("/api/projects/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		await service.deleteProject(user, c.req.param("id"));
		return c.json({ ok: true });
	});

	app.get("/api/projects/:id/session-context", async (c) => {
		const user = await authenticateRequest(c, service);
		const context = await service.getProjectSessionContext(user, c.req.param("id"));
		return c.json(context);
	});

	// ---- 智能体广场 (FR-9/T16) ----

	app.get("/api/agents", async (c) => {
		const user = await authenticateRequest(c, service);
		return c.json({ agents: service.agents.listAgents(user) });
	});

	app.post("/api/agents", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (typeof body.name !== "string" || typeof body.url !== "string" || typeof body.visibility !== "string") {
			return c.json({ error: { code: "invalid_request", message: "需要 name/url/visibility" } }, 400);
		}
		const agent = service.agents.createAgent(user, {
			name: body.name,
			url: body.url,
			visibility: body.visibility as "private" | "group" | "public",
			description: typeof body.description === "string" ? body.description : undefined,
			groupId: typeof body.groupId === "string" ? body.groupId : undefined,
		});
		return c.json({ agent }, 201);
	});

	app.patch("/api/agents/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		const patch: {
			name?: string;
			description?: string | null;
			url?: string;
			visibility?: "private" | "group" | "public";
			groupId?: string;
		} = {};
		if (body.name !== undefined) {
			if (typeof body.name !== "string")
				return c.json({ error: { code: "invalid_request", message: "name 需为字符串" } }, 400);
			patch.name = body.name;
		}
		if (body.description !== undefined) {
			if (body.description !== null && typeof body.description !== "string")
				return c.json({ error: { code: "invalid_request", message: "description 需为字符串或 null" } }, 400);
			patch.description = body.description === null ? null : (body.description as string);
		}
		if (body.url !== undefined) {
			if (typeof body.url !== "string")
				return c.json({ error: { code: "invalid_request", message: "url 需为字符串" } }, 400);
			patch.url = body.url;
		}
		if (body.visibility !== undefined) {
			if (typeof body.visibility !== "string" || !isVisibility(body.visibility))
				return c.json(
					{ error: { code: "invalid_request", message: "visibility 必须是 private/group/public" } },
					400,
				);
			patch.visibility = body.visibility;
		}
		if (body.groupId !== undefined) {
			if (body.groupId !== null && typeof body.groupId !== "string")
				return c.json({ error: { code: "invalid_request", message: "groupId 需为字符串或 null" } }, 400);
			if (typeof body.groupId === "string") patch.groupId = body.groupId;
		}
		const agent = service.agents.updateAgent(user, c.req.param("id"), patch);
		return c.json({ agent });
	});

	app.delete("/api/agents/:id", async (c) => {
		const user = await authenticateRequest(c, service);
		service.agents.deleteAgent(user, c.req.param("id"));
		return c.json({ ok: true });
	});

	// ---- 项目文件工作空间 (FR-6/T12) ----

	app.get("/api/projects/:id/files", async (c) => {
		const user = await authenticateRequest(c, service);
		const rel = c.req.query("path");
		const files = await service.files.listFiles(user, c.req.param("id"), rel && rel.length > 0 ? rel : undefined);
		return c.json({ files });
	});

	app.get("/api/projects/:id/files/content", async (c) => {
		const user = await authenticateRequest(c, service);
		const rel = c.req.query("path");
		if (!rel || !rel.length)
			return c.json({ error: { code: "invalid_request", message: "需要 path 查询参数" } }, 400);
		const file = await service.files.readText(user, c.req.param("id"), rel);
		return c.json({ file });
	});

	app.put("/api/projects/:id/files/content", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		if (typeof body.path !== "string" || typeof body.content !== "string") {
			return c.json({ error: { code: "invalid_request", message: "需要 path 和 content" } }, 400);
		}
		const file = await service.files.writeText(user, c.req.param("id"), body.path, body.content);
		return c.json({ file });
	});

	app.post("/api/projects/:id/files", async (c) => {
		const user = await authenticateRequest(c, service);
		const form = await c.req.formData();
		const file = form.get("file");
		if (!(file instanceof File))
			return c.json({ error: { code: "invalid_request", message: "需要 file 字段" } }, 400);
		const dir = form.get("path");
		const dirPath = typeof dir === "string" && dir.length > 0 ? dir.replace(/[\\/]+$/, "") : "";
		const rel = dirPath ? `${dirPath}/${file.name}` : file.name;
		const saved = await service.files.uploadFile(user, c.req.param("id"), rel, file);
		return c.json({ file: saved }, 201);
	});

	app.get("/api/projects/:id/files/download", async (c) => {
		const user = await authenticateRequest(c, service);
		const rel = c.req.query("path");
		if (!rel || !rel.length)
			return c.json({ error: { code: "invalid_request", message: "需要 path 查询参数" } }, 400);
		const info = await service.files.downloadInfo(user, c.req.param("id"), rel);
		const content = await readFile(info.absPath);
		const encoded = encodeURIComponent(info.name);
		return c.body(new Uint8Array(content), {
			status: 200,
			headers: {
				"Content-Type": info.mimeType,
				"Content-Disposition": `attachment; filename*=UTF-8'${encoded}`,
			},
		});
	});

	app.delete("/api/projects/:id/files", async (c) => {
		const user = await authenticateRequest(c, service);
		const rel = c.req.query("path");
		if (!rel || !rel.length)
			return c.json({ error: { code: "invalid_request", message: "需要 path 查询参数" } }, 400);
		await service.files.deleteFile(user, c.req.param("id"), rel);
		return c.json({ ok: true });
	});

	app.post("/api/sessions/:id/project", async (c) => {
		const user = await authenticateRequest(c, service);
		const body = await readJsonBody(c);
		const raw = body.projectId;
		if (raw !== null && raw !== undefined && typeof raw !== "string") {
			return c.json({ error: { code: "invalid_request", message: "projectId 需为字符串或 null" } }, 400);
		}
		await service.moveSessionToProject(user, c.req.param("id"), typeof raw === "string" ? raw : null);
		return c.json({ ok: true });
	});

	app.delete("/api/config/api-key", async (c) => {
		const provider = c.req.query("provider");
		if (!provider) {
			return c.json({ error: { code: "invalid_request", message: "需要 provider 查询参数" } }, 400);
		}
		await service.removeApiKey(provider);
		return c.json({ ok: true });
	});

	app.post("/api/config/api-key", async (c) => {
		const body = await readJsonBody(c);
		if (typeof body.provider !== "string" || typeof body.apiKey !== "string") {
			return c.json({ error: { code: "invalid_request", message: "需要 provider 和 apiKey" } }, 400);
		}
		await service.configureApiKey(body.provider, body.apiKey);
		return c.json({ ok: true });
	});

	app.get("/api/sessions/:id", async (c) => {
		const runtime = await service.openSession(c.req.param("id"));
		return c.json({ session: runtime.snapshot() });
	});

	app.delete("/api/sessions/:id", async (c) => {
		await service.deleteSession(c.req.param("id"));
		return c.json({ ok: true });
	});

	app.post("/api/sessions/:id/prompt", async (c) => {
		const runtime = await service.openSession(c.req.param("id"));
		const body = await readJsonBody(c);
		const text = typeof body.text === "string" ? body.text : "";
		if (!text.trim()) {
			return c.json({ error: { code: "invalid_request", message: "text 不能为空" } }, 400);
		}
		await runtime.prompt({ text });
		return c.json({ ok: true });
	});

	app.post("/api/sessions/:id/steer", async (c) => {
		const runtime = await service.openSession(c.req.param("id"));
		const body = await readJsonBody(c);
		const text = typeof body.text === "string" ? body.text : "";
		if (!text.trim()) {
			return c.json({ error: { code: "invalid_request", message: "text 不能为空" } }, 400);
		}
		await runtime.steer({ text });
		return c.json({ ok: true });
	});

	app.post("/api/sessions/:id/abort", async (c) => {
		const runtime = await service.openSession(c.req.param("id"));
		await runtime.abort();
		return c.json({ ok: true });
	});

	app.post("/api/sessions/:id/model", async (c) => {
		const runtime = await service.openSession(c.req.param("id"));
		const body = await readJsonBody(c);
		if (!isModelRef(body.model)) {
			return c.json({ error: { code: "invalid_request", message: "model 格式为 { provider, id }" } }, 400);
		}
		await runtime.setModel(body.model);
		return c.json({ ok: true });
	});

	app.post("/api/sessions/:id/thinking", async (c) => {
		const runtime = await service.openSession(c.req.param("id"));
		const body = await readJsonBody(c);
		if (!isThinkingLevel(body.thinkingLevel)) {
			return c.json({ error: { code: "invalid_request", message: "thinkingLevel 无效" } }, 400);
		}
		await runtime.setThinking(body.thinkingLevel);
		return c.json({ ok: true });
	});

	app.post("/api/sessions/:id/rename", async (c) => {
		const body = await readJsonBody(c);
		if (typeof body.name !== "string") {
			return c.json({ error: { code: "invalid_request", message: "需要 name" } }, 400);
		}
		await service.renameSession(c.req.param("id"), body.name);
		return c.json({ ok: true });
	});

	app.get("/api/sessions/:id/export", async (c) => {
		const content = await service.exportSession(c.req.param("id"));
		return c.text(content, 200, { "Content-Type": "text/plain; charset=utf-8" });
	});

	app.get("/api/sessions/:id/events", async (c) => {
		const runtime = await service.openSession(c.req.param("id"));
		const encoder = new TextEncoder();
		let unsubscribe: (() => void) | undefined;
		let keepAlive: NodeJS.Timeout | undefined;

		const stream = new ReadableStream<Uint8Array>({
			start(controller) {
				// Flush the SSE response immediately. Without an initial frame,
				// Node may buffer the headers/body until the first runtime event,
				// leaving EventSource stuck in the "connecting" state.
				controller.enqueue(encoder.encode(": connected\n\n"));
				unsubscribe = runtime.subscribe((event) => {
					try {
						controller.enqueue(encoder.encode(serializeSseEvent(event)));
					} catch {
						// 连接已关闭
					}
				});
				keepAlive = setInterval(() => {
					try {
						controller.enqueue(encoder.encode(": keep-alive\n\n"));
					} catch {
						// 连接已关闭
					}
				}, SSE_KEEPALIVE_MS);
			},
			cancel() {
				if (keepAlive) clearInterval(keepAlive);
				unsubscribe?.();
			},
		});

		return c.body(stream, {
			status: 200,
			headers: {
				"Content-Type": "text/event-stream",
				"Cache-Control": "no-cache",
				Connection: "keep-alive",
				"X-Accel-Buffering": "no",
			},
		});
	});

	// 静态托管 + SPA 兜底
	app.get("*", async (c) => {
		const urlPath = new URL(c.req.url).pathname;
		const filePath = join(staticDir, urlPath === "/" ? "index.html" : urlPath);
		try {
			const content = await readFile(filePath);
			const mime = MIME_TYPES[extname(filePath).toLowerCase()] ?? "application/octet-stream";
			return c.body(new Uint8Array(content), { status: 200, headers: { "Content-Type": mime } });
		} catch {
			try {
				const index = await readFile(join(staticDir, "index.html"));
				return c.body(new Uint8Array(index), {
					status: 200,
					headers: { "Content-Type": "text/html; charset=utf-8" },
				});
			} catch {
				return c.json({ error: { code: "not_found", message: "前端构建产物不存在,请先构建 web" } }, 404);
			}
		}
	});

	app.onError((error, c) => {
		if (error instanceof QuotaExceededError) {
			return c.json({ error: { code: error.code, message: error.message } }, 413);
		}
		if (error instanceof AuthError) {
			const status = error.code === "forbidden" ? 403 : 401;
			return c.json({ error: { code: error.code, message: error.message } }, status);
		}
		const { status, body } = toErrorBody(error);
		return new Response(JSON.stringify(body), {
			status,
			headers: { "Content-Type": "application/json; charset=utf-8" },
		});
	});

	return app;
}

function isProjectScope(value: unknown): value is { type: "private" } | { type: "group"; groupId: string } {
	if (typeof value !== "object" || value === null) return false;
	const scope = value as { type?: unknown; groupId?: unknown };
	if (scope.type === "private") return true;
	return scope.type === "group" && typeof scope.groupId === "string";
}
function isModelRef(value: unknown): value is { provider: string; id: string } {
	return (
		typeof value === "object" &&
		value !== null &&
		typeof (value as { provider?: unknown }).provider === "string" &&
		typeof (value as { id?: unknown }).id === "string"
	);
}

const THINKING_LEVELS = new Set(["off", "minimal", "low", "medium", "high", "xhigh", "max"]);

function isThinkingLevel(value: unknown): value is "off" | "minimal" | "low" | "medium" | "high" | "xhigh" | "max" {
	return typeof value === "string" && THINKING_LEVELS.has(value);
}

/**
 * 启动 HTTP 服务:默认监听 127.0.0.1:8787。
 */
export async function startHttpServer(options: HttpServerOptions): Promise<HttpServerHandle> {
	const port = options.port ?? 8787;
	const host = options.host ?? "127.0.0.1";
	const app = createApp(options.service, options);
	const server: Server = createServer((req, res) => {
		const headers = new Headers();
		for (const [name, value] of Object.entries(req.headers)) {
			if (typeof value === "string") {
				headers.set(name, value);
			} else if (Array.isArray(value)) {
				headers.set(name, value.join(", "));
			}
		}
		const request = new Request(`http://${req.headers.host ?? `${host}:${port}`}${req.url}`, {
			method: req.method,
			headers,
			body: req.method === "GET" || req.method === "HEAD" ? undefined : req,
			duplex: "half",
		});
		Promise.resolve(app.fetch(request))
			.then((response) => {
				res.writeHead(response.status, Object.fromEntries(response.headers.entries()));
				if (response.body) {
					const reader = response.body.getReader();
					const pump = (): void => {
						void reader.read().then(({ done, value }) => {
							if (done) {
								res.end();
								return;
							}
							res.write(Buffer.from(value));
							pump();
						});
					};
					pump();
				} else {
					res.end();
				}
			})
			.catch((error: unknown) => {
				res.writeHead(500, { "Content-Type": "application/json" });
				res.end(JSON.stringify({ error: { code: "internal_error", message: String(error) } }));
			});
	});

	await new Promise<void>((resolvePromise) => {
		server.listen(port, host, () => resolvePromise());
	});

	return {
		url: `http://${host}:${port}`,
		async shutdown() {
			await new Promise<void>((resolvePromise) => {
				server.close(() => resolvePromise());
			});
			await options.service.close();
		},
	};
}
