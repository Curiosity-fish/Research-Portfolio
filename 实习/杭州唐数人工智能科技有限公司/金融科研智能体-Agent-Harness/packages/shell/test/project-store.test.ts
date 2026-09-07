import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { ProjectStore } from "../src/server/project-store.ts";

const stores: ProjectStore[] = [];
const cleanupDirs: string[] = [];

afterEach(() => {
	for (const store of stores) {
		try {
			store.close();
		} catch {
			// 已关闭的句柄忽略
		}
	}
	stores.length = 0;
	for (const dir of cleanupDirs) {
		rmSync(dir, { recursive: true, force: true });
	}
	cleanupDirs.length = 0;
});

function createStore(): ProjectStore {
	const dir = mkdtempSync(join(tmpdir(), "tfa-project-test-"));
	cleanupDirs.push(dir);
	const store = new ProjectStore(join(dir, "agent"));
	stores.push(store);
	return store;
}

describe("ProjectStore", () => {
	it("create → list → update → delete", async () => {
		const store = createStore();
		const created = await store.create({ name: "新能源行业" });
		expect(created.id).toBeTruthy();
		expect(created.name).toBe("新能源行业");
		expect(created.folder).toBeUndefined();
		expect(created.sessionIds).toEqual([]);

		await expect(store.create({ name: "   " })).rejects.toMatchObject({ code: "invalid_request" });

		const listed = await store.list();
		expect(listed).toHaveLength(1);
		expect(listed[0].name).toBe("新能源行业");

		const updated = await store.update(created.id, { name: "新能源与储能", folder: "D:\\research\\新能源" });
		expect(updated.name).toBe("新能源与储能");
		expect(updated.folder).toBe("D:\\research\\新能源");

		await store.delete(created.id);
		expect(await store.list()).toHaveLength(0);
		await expect(store.delete(created.id)).rejects.toMatchObject({ code: "not_found" });
	});

	it("moveSession 切换项目归属,removeSession 清理", async () => {
		const store = createStore();
		const a = await store.create({ name: "项目A" });
		const b = await store.create({ name: "项目B" });

		await store.moveSession("s-1", a.id);
		await store.moveSession("s-2", a.id);
		let projects = await store.list();
		expect(projects.find((p) => p.id === a.id)?.sessionIds).toEqual(["s-1", "s-2"]);

		// 移动到另一个项目:从旧项目移除
		await store.moveSession("s-1", b.id);
		projects = await store.list();
		expect(projects.find((p) => p.id === a.id)?.sessionIds).toEqual(["s-2"]);
		expect(projects.find((p) => p.id === b.id)?.sessionIds).toEqual(["s-1"]);

		// 移出项目(无项目)
		await store.moveSession("s-2", null);
		projects = await store.list();
		expect(projects.find((p) => p.id === a.id)?.sessionIds).toEqual([]);

		// 目标项目不存在
		await expect(store.moveSession("s-9", "no-such-project")).rejects.toMatchObject({ code: "not_found" });

		// 删除会话:从所有项目清理
		await store.moveSession("s-3", b.id);
		await store.removeSession("s-3");
		projects = await store.list();
		expect(projects.find((p) => p.id === b.id)?.sessionIds).toEqual(["s-1"]);
	});
});
