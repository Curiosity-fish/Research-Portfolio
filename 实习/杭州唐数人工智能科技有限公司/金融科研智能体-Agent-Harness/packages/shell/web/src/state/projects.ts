/**
 * 会话归属模型:项目 = 用户创建的分类(可选绑定本地文件夹)。
 * 依据后端 projects[].sessionIds 把会话分为「无项目」与「按项目分组」两组。
 * 不再按 cwd 自动分组。
 */

import type { ProjectInfo, SessionSummary } from "../api/types.ts";

export interface ProjectWithSessions {
	project: ProjectInfo;
	sessions: SessionSummary[];
}

export interface SessionGrouping {
	/** 无项目会话(不在任何项目 sessionIds 中),按 updatedAt 倒序 */
	unassigned: SessionSummary[];
	/** 有归属项目的会话,按项目分组;项目按 updatedAt 倒序,组内会话按 updatedAt 倒序 */
	projects: ProjectWithSessions[];
}

export function groupSessions(sessions: SessionSummary[], projects: ProjectInfo[]): SessionGrouping {
	const sorted = [...sessions].sort((a, b) => b.updatedAt - a.updatedAt);
	const byProject = new Map<string, ProjectWithSessions>();
	for (const project of projects) {
		byProject.set(project.id, { project, sessions: [] });
	}
	const unassigned: SessionSummary[] = [];
	for (const session of sorted) {
		const owner = projects.find((project) => project.sessionIds.includes(session.id));
		const group = owner ? byProject.get(owner.id) : undefined;
		if (owner && group) {
			group.sessions.push(session);
		} else {
			unassigned.push(session);
		}
	}
	const grouped = [...byProject.values()].sort((a, b) => b.project.updatedAt - a.project.updatedAt);
	return { unassigned, projects: grouped };
}

/** 查找会话当前所属项目 id(无项目返回 null)。 */
export function sessionProjectId(sessionId: string, projects: ProjectInfo[]): string | null {
	for (const project of projects) {
		if (project.sessionIds.includes(sessionId)) return project.id;
	}
	return null;
}