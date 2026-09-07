import { mkdir, readdir, readFile, rm, stat, writeFile } from "node:fs/promises";
import { basename, dirname, extname, join, resolve, sep } from "node:path";
import { TfaServerError } from "@earendil-works/tfa-server";
import type { AuthUserView } from "./auth.ts";
import type { ProjectStore } from "./project-store.ts";

const TEXT_EXTENSIONS = new Set([
	".txt",
	".md",
	".json",
	".csv",
	".py",
	".ts",
	".tsx",
	".js",
	".mjs",
	".cjs",
	".jsx",
	".vue",
	".html",
	".css",
	".yaml",
	".yml",
	".toml",
	".ini",
	".sh",
	".ps1",
	".bat",
	".sql",
	".log",
	".xml",
]);

function mimeTypeFor(name: string): string {
	switch (extname(name).toLowerCase()) {
		case ".pdf":
			return "application/pdf";
		case ".png":
			return "image/png";
		case ".jpg":
		case ".jpeg":
			return "image/jpeg";
		case ".gif":
			return "image/gif";
		case ".webp":
			return "image/webp";
		case ".svg":
			return "image/svg+xml";
		case ".csv":
			return "text/csv";
		case ".json":
			return "application/json";
		case ".md":
			return "text/markdown";
		case ".html":
			return "text/html";
		case ".css":
			return "text/css";
		case ".js":
			return "text/javascript";
		case ".ts":
			return "text/typescript";
		case ".py":
			return "text/x-python";
		default:
			return "application/octet-stream";
	}
}

function isTextFile(name: string): boolean {
	const ext = extname(name).toLowerCase();
	return TEXT_EXTENSIONS.has(ext) || ext === "";
}

export interface ProjectFileEntry {
	name: string;
	/** 相对项目工作区的路径(正斜杠)。 */
	path: string;
	type: "file" | "dir";
	size: number | null;
	mtime: number | null;
}

export interface ProjectFileServiceOptions {
	projects: ProjectStore;
	/** 平台托管工作空间根目录:{agentDir}/workspaces,每个项目一个 {projectId}/。 */
	workspacesRoot: string;
}

/**
 * 项目文件工作空间服务(FR-6/T12):平台托管目录 {workspacesRoot}/{projectId}。
 * 所有接口校验项目可见性(owner/scope),路径解析后必须落在项目工作区内(防穿越)。
 * 无大小上限(需求确认 2026-08-24)。
 */
export class ProjectFileService {
	private readonly projects: ProjectStore;
	private readonly workspacesRoot: string;

	constructor(options: ProjectFileServiceOptions) {
		this.projects = options.projects;
		this.workspacesRoot = options.workspacesRoot;
	}

	private async assertVisible(user: AuthUserView, projectId: string): Promise<void> {
		const project = await this.projects.getForUser(user, projectId);
		if (!project) throw new TfaServerError("not_found", `项目不存在: ${projectId}`);
	}

	private workspace(projectId: string): string {
		return join(this.workspacesRoot, projectId);
	}

	/** 确保项目工作区目录存在(会话启动/文件操作前调用)。 */
	async ensureWorkspace(projectId: string): Promise<void> {
		await mkdir(this.workspace(projectId), { recursive: true });
	}

	/** 路径必须解析到项目工作区内;越界抛 invalid_request(400)。 */
	private safeResolve(projectId: string, relPath: string | undefined): string {
		const root = this.workspace(projectId);
		const target = resolve(root, relPath ?? "");
		if (target !== root && !target.startsWith(root + sep)) {
			throw new TfaServerError("invalid_request", "路径超出项目工作区");
		}
		return target;
	}

	private async statOrThrow(target: string, what = "路径"): Promise<Awaited<ReturnType<typeof stat>>> {
		const st = await stat(target).catch(() => undefined);
		if (!st) throw new TfaServerError("not_found", `${what}不存在`);
		return st;
	}

	/** 列出目录内容(可选相对路径)。 */
	async listFiles(user: AuthUserView, projectId: string, relPath?: string): Promise<ProjectFileEntry[]> {
		await this.assertVisible(user, projectId);
		const dir = this.safeResolve(projectId, relPath);
		const st = await stat(dir).catch(() => undefined);
		if (!st) {
			// 新项目工作区尚未创建:根目录返回空列表,子路径返回 404
			if (dir === this.workspace(projectId)) return [];
			throw new TfaServerError("not_found", "路径不存在");
		}
		if (!st.isDirectory()) throw new TfaServerError("invalid_request", "目标不是目录");
		const names = await readdir(dir);
		const entries: ProjectFileEntry[] = [];
		for (const name of names) {
			const full = join(dir, name);
			const fileStat = await stat(full);
			entries.push({
				name,
				path: relPath ? `${relPath.replace(/[\\/]+$/, "")}/${name}` : name,
				type: fileStat.isDirectory() ? "dir" : "file",
				size: fileStat.isDirectory() ? null : Number(fileStat.size),
				mtime: fileStat.mtimeMs,
			});
		}
		entries.sort((a, b) => {
			if (a.type !== b.type) return a.type === "dir" ? -1 : 1;
			return a.name.localeCompare(b.name);
		});
		return entries;
	}

	/** 读取文本文件内容与元数据;二进制文件调用方应走 download。 */
	async readText(
		user: AuthUserView,
		projectId: string,
		relPath: string,
	): Promise<{ path: string; name: string; size: number; mimeType: string; content: string }> {
		await this.assertVisible(user, projectId);
		const file = this.safeResolve(projectId, relPath);
		const st = await this.statOrThrow(file, "文件");
		if (!st.isFile()) throw new TfaServerError("invalid_request", "目标不是文件");
		const content = isTextFile(file) ? await readFile(file, "utf8") : "";
		return { path: relPath, name: basename(file), size: Number(st.size), mimeType: mimeTypeFor(file), content };
	}

	/** 保存文本/代码(自动创建父目录);无大小上限。 */
	async writeText(
		user: AuthUserView,
		projectId: string,
		relPath: string,
		content: string,
	): Promise<{ path: string; name: string; size: number; mimeType: string }> {
		await this.assertVisible(user, projectId);
		const file = this.safeResolve(projectId, relPath);
		await mkdir(dirname(file), { recursive: true });
		await writeFile(file, content, "utf8");
		const st = await stat(file);
		return { path: relPath, name: basename(file), size: Number(st.size), mimeType: mimeTypeFor(file) };
	}

	/** 上传文件到项目工作区(relPath 为含文件名的相对路径,自动创建父目录);无大小上限。 */
	async uploadFile(
		user: AuthUserView,
		projectId: string,
		relPath: string,
		file: File,
	): Promise<{ path: string; name: string; size: number; mimeType: string }> {
		await this.assertVisible(user, projectId);
		if (file.size <= 0) throw new TfaServerError("invalid_request", "文件不能为空");
		const target = this.safeResolve(projectId, relPath);
		await mkdir(dirname(target), { recursive: true });
		await writeFile(target, Buffer.from(await file.arrayBuffer()));
		const st = await stat(target);
		return { path: relPath, name: basename(target), size: Number(st.size), mimeType: mimeTypeFor(target) };
	}

	/** 下载信息(供路由流式返回);不可见/不存在抛错。 */
	async downloadInfo(
		user: AuthUserView,
		projectId: string,
		relPath: string,
	): Promise<{ absPath: string; name: string; size: number; mimeType: string }> {
		await this.assertVisible(user, projectId);
		const file = this.safeResolve(projectId, relPath);
		const st = await this.statOrThrow(file, "文件");
		if (!st.isFile()) throw new TfaServerError("invalid_request", "目标不是文件");
		return { absPath: file, name: basename(file), size: Number(st.size), mimeType: mimeTypeFor(file) };
	}

	/** 删除文件或目录(递归,仅限项目工作区内;禁止删除工作区根)。 */
	async deleteFile(user: AuthUserView, projectId: string, relPath: string): Promise<void> {
		await this.assertVisible(user, projectId);
		const root = this.workspace(projectId);
		const target = this.safeResolve(projectId, relPath);
		if (target === root) throw new TfaServerError("invalid_request", "不能删除工作区根目录");
		const st = await this.statOrThrow(target);
		await rm(target, { recursive: st.isDirectory(), force: true });
	}
}
