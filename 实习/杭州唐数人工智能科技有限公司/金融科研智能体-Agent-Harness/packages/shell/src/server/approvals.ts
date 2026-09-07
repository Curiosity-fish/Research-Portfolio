import { randomUUID } from "node:crypto";
import { TfaServerError } from "@earendil-works/tfa-server";

export type ApprovalStatus = "pending" | "approved" | "rejected" | "expired";

export interface ApprovalRequest {
	id: string;
	sessionId: string;
	/** 待批准的操作,如 write/delete/bash/skill。 */
	tool: string;
	/** 目标路径/命令。 */
	target: string;
	risk: string;
	status: ApprovalStatus;
	createdAt: number;
	expiresAt: number;
	decidedAt: number | null;
}

export interface CreateApprovalInput {
	tool: string;
	target: string;
	risk: string;
}

/**
 * 请求批准状态机(T14):request_approval 模式下高风险工具调用在执行前登记请求,
 * 由用户批准/拒绝;超时或断线默认拒绝。执行面拦截待 T8b 桥接接入。
 */
export class ApprovalService {
	private readonly requests = new Map<string, ApprovalRequest>();
	private readonly timeoutMs: number;

	constructor(options: { timeoutMs?: number } = {}) {
		this.timeoutMs = options.timeoutMs ?? 5 * 60 * 1000;
	}

	createRequest(sessionId: string, input: CreateApprovalInput): ApprovalRequest {
		const now = Date.now();
		const request: ApprovalRequest = {
			id: randomUUID(),
			sessionId,
			tool: input.tool,
			target: input.target,
			risk: input.risk,
			status: "pending",
			createdAt: now,
			expiresAt: now + this.timeoutMs,
			decidedAt: null,
		};
		this.requests.set(request.id, request);
		return request;
	}

	get(id: string): ApprovalRequest | undefined {
		return this.requests.get(id);
	}

	listPending(sessionId: string): ApprovalRequest[] {
		return [...this.requests.values()].filter(
			(request) => request.sessionId === sessionId && request.status === "pending",
		);
	}

	/** 批准:仅 pending 可批准;重复批准幂等;已拒绝/超时返回 409。 */
	approve(id: string): ApprovalRequest {
		const request = this.get(id);
		if (!request) throw new TfaServerError("not_found", `批准请求不存在: ${id}`);
		if (request.status === "approved") return request;
		if (request.status === "pending") {
			this.setDecided(request, "approved");
			return request;
		}
		throw new TfaServerError("invalid_request", `该请求已${request.status === "rejected" ? "拒绝" : "超时"}`);
	}

	/** 拒绝:仅 pending 可拒绝;重复拒绝幂等。 */
	reject(id: string): ApprovalRequest {
		const request = this.get(id);
		if (!request) throw new TfaServerError("not_found", `批准请求不存在: ${id}`);
		if (request.status === "rejected" || request.status === "expired") return request;
		if (request.status === "pending") {
			this.setDecided(request, "rejected");
			return request;
		}
		throw new TfaServerError("invalid_request", "该请求已批准");
	}

	/** 超时清理:所有超时 pending 置为 expired(默认拒绝语义)。 */
	expireOverdue(now = Date.now()): number {
		let count = 0;
		for (const request of this.requests.values()) {
			if (request.status === "pending" && request.expiresAt <= now) {
				this.setDecided(request, "expired");
				count += 1;
			}
		}
		return count;
	}

	private setDecided(request: ApprovalRequest, status: ApprovalStatus): void {
		request.status = status;
		request.decidedAt = Date.now();
	}
}
