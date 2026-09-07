import { randomUUID } from "node:crypto";
import { mkdir, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { TfaServerError } from "@earendil-works/tfa-server";
import { AuthError, type AuthUserView } from "./auth.ts";
import type { PlatformRepos } from "./db/repos.ts";
import type { DatasetRecord, ResourceVisibility, SkillReviewState } from "./db/types.ts";
import { QuotaExceededError } from "./errors.ts";
import { isVisibility } from "./skills.ts";

/** A5 默认值:单文件 500MB,每用户累计 20GB。 */
const DEFAULT_MAX_FILE_BYTES = 500 * 1024 * 1024;
const DEFAULT_MAX_USER_BYTES = 20 * 1024 * 1024 * 1024;

function sanitizeFileName(name: string): string {
	const clean = name.replace(/[\\/:*?"<>|\r\n]+/g, "_").trim();
	return clean || "dataset";
}

export interface DatasetView {
	id: string;
	name: string;
	sizeBytes: number;
	visibility: ResourceVisibility;
	groupId: string | null;
	groupName: string | undefined;
	ownerId: string | null;
	status: string;
	reviewState: SkillReviewState;
	createdAt: number;
}

export interface DatasetServiceOptions {
	repos: PlatformRepos;
	/** 数据集存储根目录: {agentDir}/platform/datasets,每个数据集一个 {id}/ 目录。 */
	storageRoot: string;
	/** 单文件上限字节,默认 500MB。 */
	maxFileBytes?: number;
	/** 每用户累计上限字节,默认 20GB。 */
	maxUserBytes?: number;
}

/**
 * 数据集库服务:方法以当前用户为第一参数,服务层强制 ACL。
 * 组库对归档组成员仍只读可见;编辑仅 owner/admin。
 */
export class DatasetService {
	private readonly repos: PlatformRepos;
	private readonly storageRoot: string;
	private readonly maxFileBytes: number;
	private readonly maxUserBytes: number;

	constructor(options: DatasetServiceOptions) {
		this.repos = options.repos;
		this.storageRoot = options.storageRoot;
		this.maxFileBytes = options.maxFileBytes ?? DEFAULT_MAX_FILE_BYTES;
		this.maxUserBytes = options.maxUserBytes ?? DEFAULT_MAX_USER_BYTES;
		void mkdir(this.storageRoot, { recursive: true }).catch(() => {
			// 目录创建失败延迟到实际写入时报错
		});
	}

	private getDatasetOrThrow(id: string): DatasetRecord {
		const dataset = this.repos.datasetGet(id);
		if (!dataset) throw new TfaServerError("not_found", `数据集不存在: ${id}`);
		return dataset;
	}

	private toDatasetView(record: DatasetRecord): DatasetView {
		return {
			id: record.id,
			name: record.name,
			sizeBytes: record.sizeBytes,
			visibility: record.visibility,
			groupId: record.groupId,
			groupName: record.groupId ? (this.repos.groupGet(record.groupId)?.name ?? "未知组") : undefined,
			ownerId: record.ownerId,
			status: record.status,
			reviewState: record.reviewState,
			createdAt: record.createdAt,
		};
	}

	private canSee(user: AuthUserView, record: DatasetRecord): boolean {
		if (record.status !== "active") return false;
		if (record.ownerId === user.id) return true;
		if (record.visibility === "public") return true;
		if (record.visibility === "group" && record.groupId) {
			return this.repos.groupMemberRole(record.groupId, user.id) !== undefined;
		}
		return false;
	}

	/** 当前用户可见数据集:本人私有 ∪ 所在组库 ∪ 公开(admin 另见全部含已下架)。 */
	listDatasets(user: AuthUserView): DatasetView[] {
		const records = user.role === "admin" ? this.repos.datasetList() : this.repos.datasetListVisible(user.id);
		return records.map((record) => this.toDatasetView(record));
	}

	/** 某项目作用域下可见数据集:本人私有 ∪ (挂组则仅该组库) ∪ 公开(admin 同规则)。 */
	listForProject(user: AuthUserView, scope: { type: "private" } | { type: "group"; groupId: string }): DatasetView[] {
		const records = user.role === "admin" ? this.repos.datasetList() : this.repos.datasetListVisible(user.id);
		const groupId = scope.type === "group" ? scope.groupId : null;
		return records
			.filter((record) => {
				if (record.status !== "active") return false;
				if (record.visibility === "public") return true;
				if (record.visibility === "private" && record.ownerId === user.id) return true;
				return record.visibility === "group" && groupId !== null && record.groupId === groupId;
			})
			.map((record) => this.toDatasetView(record));
	}
	/** 上传数据集:校验可见性与配额(A5),文件落盘 {storageRoot}/{id}/{安全文件名}。 */
	async createDataset(
		user: AuthUserView,
		input: { file: File; name?: string; visibility: ResourceVisibility; groupId?: string },
	): Promise<DatasetView> {
		if (!isVisibility(input.visibility)) {
			throw new TfaServerError("invalid_request", "visibility 必须是 private/group/public");
		}
		if (input.file.size <= 0) throw new TfaServerError("invalid_request", "文件不能为空");
		if (input.file.size > this.maxFileBytes) {
			throw new QuotaExceededError(`单文件超过上限 ${Math.floor(this.maxFileBytes / (1024 * 1024))}MB`);
		}
		const currentTotal = this.repos.datasetTotalSizeByOwner(user.id);
		if (currentTotal + input.file.size > this.maxUserBytes) {
			throw new QuotaExceededError("每用户数据集累计超过上限");
		}
		let groupId: string | null = null;
		if (input.visibility === "group") {
			if (!input.groupId) throw new TfaServerError("invalid_request", "组数据集需要 groupId");
			const group = this.repos.groupGet(input.groupId);
			if (!group) throw new TfaServerError("not_found", `组不存在: ${input.groupId}`);
			if (group.status === "archived") throw new TfaServerError("invalid_request", "归档组不可新增数据集");
			if (user.role !== "admin" && this.repos.groupMemberRole(group.id, user.id) === undefined) {
				throw new AuthError("forbidden", "不是该组成员,不能在该组共享数据集");
			}
			groupId = group.id;
		}
		const name = input.name?.trim() || input.file.name || "未命名数据集";

		const id = randomUUID();
		const dir = join(this.storageRoot, id);
		await mkdir(dir, { recursive: true });
		const safeName = sanitizeFileName(name);
		const path = join(dir, safeName);
		await writeFile(path, Buffer.from(await input.file.arrayBuffer()));

		const record = this.repos.datasetCreate({
			id,
			name,
			sizeBytes: input.file.size,
			visibility: input.visibility,
			groupId,
			ownerId: user.id,
			path,
			...(input.visibility === "public" ? { reviewState: "pending" as const } : {}),
		});
		return this.toDatasetView(record);
	}

	/** 改名/改可见性;owner 或 admin。组可见性切换须同时给 groupId。 */
	updateDataset(
		user: AuthUserView,
		id: string,
		input: { name?: string; visibility?: ResourceVisibility; groupId?: string },
	): DatasetView {
		const dataset = this.getDatasetOrThrow(id);
		if (dataset.ownerId !== user.id && user.role !== "admin") {
			throw new AuthError("forbidden", "只有数据集 owner 或管理员可以修改");
		}
		let visibility = dataset.visibility;
		let groupId = dataset.groupId;
		if (input.visibility !== undefined) {
			if (!isVisibility(input.visibility)) {
				throw new TfaServerError("invalid_request", "visibility 必须是 private/group/public");
			}
			visibility = input.visibility;
			if (visibility === "group") {
				if (!input.groupId) throw new TfaServerError("invalid_request", "组数据集需要 groupId");
				const group = this.repos.groupGet(input.groupId);
				if (!group) throw new TfaServerError("not_found", `组不存在: ${input.groupId}`);
				if (group.status === "archived") throw new TfaServerError("invalid_request", "归档组不可新增数据集");
				if (user.role !== "admin" && this.repos.groupMemberRole(group.id, user.id) === undefined) {
					throw new AuthError("forbidden", "不是该组成员,不能在该组共享数据集");
				}
				groupId = group.id;
			} else {
				groupId = null;
			}
		} else if (input.groupId !== undefined) {
			throw new TfaServerError("invalid_request", "仅当设置 visibility=group 时可指定 groupId");
		}
		const name = input.name?.trim() ?? dataset.name;
		if (!name) throw new TfaServerError("invalid_request", "数据集名称不能为空");

		this.repos.datasetUpdate(id, { name, visibility, groupId });
		if (visibility === "public" && dataset.visibility !== "public") {
			this.repos.datasetSetReview(id, "pending");
		}
		return this.toDatasetView(this.repos.datasetGet(id)!);
	}

	/** 下架(软删);owner 或 admin。 */
	deleteDataset(user: AuthUserView, id: string): void {
		const dataset = this.getDatasetOrThrow(id);
		if (dataset.ownerId !== user.id && user.role !== "admin") {
			throw new AuthError("forbidden", "只有数据集 owner 或管理员可以下架");
		}
		this.repos.datasetSetStatus(id, "removed");
	}

	/** 审核公开数据集(admin):pending → approved/rejected。 */
	reviewDataset(user: AuthUserView, id: string, reviewState: SkillReviewState): DatasetView {
		if (user.role !== "admin") throw new AuthError("forbidden", "需要管理员权限");
		if (reviewState !== "approved" && reviewState !== "rejected") {
			throw new TfaServerError("invalid_request", "reviewState 必须是 approved 或 rejected");
		}
		const dataset = this.getDatasetOrThrow(id);
		this.repos.datasetSetReview(dataset.id, reviewState);
		const updated = this.repos.datasetGet(id);
		return this.toDatasetView(updated!);
	}
	/** 下载信息(可见用户只读);不可见返回 undefined。 */
	getDownload(user: AuthUserView, id: string): { path: string; name: string } | undefined {
		const dataset = this.getDatasetOrThrow(id);
		if (!this.canSee(user, dataset)) return undefined;
		if (!dataset.path) return undefined;
		return { path: dataset.path, name: dataset.name };
	}
}
