import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { applyMigrations, openPlatformDatabase } from "../src/server/db/database.ts";
import { PlatformRepos } from "../src/server/db/repos.ts";
import type { PlatformDatabase } from "../src/server/db/types.ts";
import { ProjectStore } from "../src/server/project-store.ts";

const openDbs: PlatformDatabase[] = [];
const cleanupDirs: string[] = [];

afterEach(() => {
	for (const db of openDbs) {
		try {
			db.close();
		} catch {
			// 已关闭的句柄忽略
		}
	}
	openDbs.length = 0;
	for (const dir of cleanupDirs) {
		rmSync(dir, { recursive: true, force: true });
	}
	cleanupDirs.length = 0;
});

function tempDir(prefix: string): string {
	const dir = mkdtempSync(join(tmpdir(), prefix));
	cleanupDirs.push(dir);
	return dir;
}

function openTestDb(prefix: string): PlatformDatabase {
	const db = openPlatformDatabase(join(tempDir(prefix), "platform.db"));
	openDbs.push(db);
	return db;
}

describe("platform database migrations", () => {
	it("applyMigrations 幂等", () => {
		const db = openTestDb("tfa-db-migrate-");
		applyMigrations(db);
		applyMigrations(db);
		const rows = db.prepare("SELECT id FROM migrations").all<{ id: string }>();
		expect(rows.map((row) => row.id)).toEqual([
			"001_platform.sql",
			"002_auth_sessions.sql",
			"003_datasets_review.sql",
			"004_session_permission.sql",
			"005_agent_listings.sql",
		]);
	});
});

describe("PlatformRepos", () => {
	it("users CRUD", () => {
		const repos = new PlatformRepos(openTestDb("tfa-db-users-"));

		const created = repos.userCreate({ username: "teacher1", passwordHash: "hash", role: "teacher" });
		expect(created.role).toBe("teacher");
		expect(created.status).toBe("active");
		expect(repos.userGetById(created.id)?.username).toBe("teacher1");
		expect(repos.userGetById(created.id)?.passwordHash).toBe("hash");
		expect(repos.userGetByUsername("teacher1")?.role).toBe("teacher");

		repos.userSetStatus(created.id, "disabled");
		expect(repos.userGetById(created.id)?.status).toBe("disabled");
		expect(repos.userList()).toHaveLength(1);

		repos.userDelete(created.id);
		expect(repos.userList()).toHaveLength(0);
	});

	it("course_groups + members + 一个组多个项目", () => {
		const repos = new PlatformRepos(openTestDb("tfa-db-groups-"));

		const teacher = repos.userCreate({ username: "t", passwordHash: "h", role: "teacher" });
		const s1 = repos.userCreate({ username: "s1", passwordHash: "h", role: "student" });
		const s2 = repos.userCreate({ username: "s2", passwordHash: "h", role: "student" });

		const group = repos.groupCreate({ name: "金融数据分析", type: "course", ownerId: teacher.id });
		repos.groupAddMember(group.id, teacher.id, "owner");
		repos.groupAddMember(group.id, s1.id, "member");
		repos.groupAddMember(group.id, s2.id, "member");

		expect(repos.groupListMembers(group.id)).toHaveLength(3);
		expect(repos.groupMemberRole(group.id, s1.id)).toBe("member");
		expect(repos.groupListByMember(s1.id).map((g) => g.id)).toEqual([group.id]);

		repos.groupRemoveMember(group.id, s1.id);
		expect(repos.groupListByMember(s1.id)).toHaveLength(0);

		// 归档而非删除(有项目引用时删除会被外键拒绝)
		repos.groupSetStatus(group.id, "archived");
		expect(repos.groupGet(group.id)?.status).toBe("archived");

		// 一个课程组挂多个项目
		const p1 = repos.projectCreate({ name: "实验一", scopeType: "group", groupId: group.id });
		const p2 = repos.projectCreate({ name: "实验二", scopeType: "group", groupId: group.id });
		expect(repos.projectList().filter((p) => p.groupId === group.id)).toHaveLength(2);
		expect(repos.projectGet(p1.id)?.scopeType).toBe("group");
		expect(repos.projectGet(p2.id)?.groupId).toBe(group.id);

		// 私有项目
		const priv = repos.projectCreate({ name: "私有项目", scopeType: "private" });
		expect(priv.groupId).toBeNull();
	});

	it("projects + session links", () => {
		const repos = new PlatformRepos(openTestDb("tfa-db-projects-"));

		const a = repos.projectCreate({ name: "项目A" });
		const b = repos.projectCreate({ name: "项目B" });

		repos.projectAddSession(a.id, "s-1");
		repos.projectAddSession(a.id, "s-2");
		expect(repos.projectListSessionIds(a.id)).toEqual(["s-1", "s-2"]);

		repos.projectRemoveSession("s-1");
		expect(repos.projectListSessionIds(a.id)).toEqual(["s-2"]);

		const updated = repos.projectUpdate(a.id, { name: "项目A2", folder: "/data/a" });
		expect(updated?.name).toBe("项目A2");
		expect(updated?.folder).toBe("/data/a");

		repos.projectDelete(b.id);
		expect(repos.projectGet(b.id)).toBeUndefined();
	});

	it("skills CRUD", () => {
		const repos = new PlatformRepos(openTestDb("tfa-db-skills-"));
		const owner = repos.userCreate({ username: "owner", passwordHash: "h", role: "teacher" });
		const group = repos.groupCreate({ name: "组", type: "course", ownerId: owner.id });

		const skill = repos.skillCreate({
			name: "pdf-extract",
			visibility: "group",
			groupId: group.id,
			ownerId: owner.id,
			path: "/libraries/组/pdf-extract",
			repoUrl: "https://github.com/example/pdf-extract",
			version: "v1.0.0",
		});
		expect(skill.reviewState).toBe("none");
		expect(repos.skillGet(skill.id)?.visibility).toBe("group");

		repos.skillSetReview(skill.id, "pending");
		expect(repos.skillGet(skill.id)?.reviewState).toBe("pending");
		repos.skillSetStatus(skill.id, "removed");
		expect(repos.skillGet(skill.id)?.status).toBe("removed");

		repos.skillDelete(skill.id);
		expect(repos.skillList()).toHaveLength(0);
	});

	it("datasets CRUD", () => {
		const repos = new PlatformRepos(openTestDb("tfa-db-datasets-"));

		const ds = repos.datasetCreate({
			name: "沪深300.csv",
			sizeBytes: 1024,
			visibility: "public",
			path: "/libraries/public/沪深300.csv",
		});
		expect(repos.datasetGet(ds.id)?.sizeBytes).toBe(1024);
		expect(repos.datasetGet(ds.id)?.reviewState).toBe("none");
		expect(repos.datasetList()).toHaveLength(1);

		repos.datasetSetReview(ds.id, "pending");
		expect(repos.datasetGet(ds.id)?.reviewState).toBe("pending");

		repos.datasetSetStatus(ds.id, "removed");
		expect(repos.datasetGet(ds.id)?.status).toBe("removed");
		repos.datasetDelete(ds.id);
		expect(repos.datasetList()).toHaveLength(0);
	});

	it("sessions CRUD", () => {
		const repos = new PlatformRepos(openTestDb("tfa-db-sessions-"));
		const user = repos.userCreate({ username: "u", passwordHash: "h", role: "student" });
		const project = repos.projectCreate({ name: "项目" });

		const created = repos.sessionCreate({ projectId: project.id, userId: user.id, cwd: "/work/p-1" });
		expect(created.status).toBe("created");
		expect(repos.sessionGet(created.id)?.projectId).toBe(project.id);

		repos.sessionUpdate(created.id, { status: "running", containerId: "c-1" });
		expect(repos.sessionGet(created.id)?.containerId).toBe("c-1");
		expect(repos.sessionGet(created.id)?.status).toBe("running");

		repos.sessionUpdate(created.id, { status: "ended", endedAt: 123 });
		expect(repos.sessionGet(created.id)?.endedAt).toBe(123);

		repos.sessionDelete(created.id);
		expect(repos.sessionList()).toHaveLength(0);
	});
});

describe("ProjectStore legacy import", () => {
	it("首开时导入 projects.json,不重复导入", async () => {
		const agentDir = join(tempDir("tfa-store-import-"), "agent");
		mkdirSync(agentDir, { recursive: true });
		writeFileSync(
			join(agentDir, "projects.json"),
			JSON.stringify({
				version: 1,
				projects: [{ id: "p-1", name: "旧项目", createdAt: 1, updatedAt: 2, sessionIds: ["s-a", "s-b"] }],
			}),
		);

		const store = new ProjectStore(agentDir);
		const list = await store.list();
		expect(list).toHaveLength(1);
		expect(list[0].name).toBe("旧项目");
		expect(list[0].sessionIds).toEqual(["s-a", "s-b"]);
		store.close();

		// 二次实例化不重复导入
		const store2 = new ProjectStore(agentDir);
		expect(await store2.list()).toHaveLength(1);
		store2.close();
	});
});
