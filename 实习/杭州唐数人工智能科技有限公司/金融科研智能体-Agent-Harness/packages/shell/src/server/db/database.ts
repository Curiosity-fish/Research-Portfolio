import { readFileSync } from "node:fs";
import type { SQLInputValue } from "node:sqlite";
import { DatabaseSync } from "node:sqlite";
import type { PlatformDatabase, PlatformRunResult, PlatformStatement } from "./types.ts";

function isNamedParameters(value: unknown): value is Record<string, SQLInputValue> {
	if (value === null || typeof value !== "object") return false;
	if (Array.isArray(value) || ArrayBuffer.isView(value)) return false;
	return true;
}

function isAsyncResult(value: unknown): boolean {
	return value !== null && (typeof value === "object" || typeof value === "function") && "then" in value;
}

class NodeStatement implements PlatformStatement {
	private readonly statement: ReturnType<DatabaseSync["prepare"]>;

	constructor(statement: ReturnType<DatabaseSync["prepare"]>) {
		this.statement = statement;
	}

	run(...params: unknown[]): PlatformRunResult {
		const [first, ...rest] = params;
		const result = isNamedParameters(first)
			? this.statement.run(first, ...(rest as SQLInputValue[]))
			: this.statement.run(...(params as SQLInputValue[]));
		return {
			changes: Number(result.changes),
			lastInsertRowid: result.lastInsertRowid === undefined ? undefined : Number(result.lastInsertRowid),
		};
	}

	get<TRow extends object>(...params: unknown[]): TRow | undefined {
		const [first, ...rest] = params;
		return (
			isNamedParameters(first)
				? this.statement.get(first, ...(rest as SQLInputValue[]))
				: this.statement.get(...(params as SQLInputValue[]))
		) as TRow | undefined;
	}

	all<TRow extends object>(...params: unknown[]): TRow[] {
		const [first, ...rest] = params;
		return (
			isNamedParameters(first)
				? this.statement.all(first, ...(rest as SQLInputValue[]))
				: this.statement.all(...(params as SQLInputValue[]))
		) as TRow[];
	}
}

class NodePlatformDatabase implements PlatformDatabase {
	private readonly db: DatabaseSync;

	constructor(db: DatabaseSync) {
		this.db = db;
	}

	exec(sql: string): void {
		this.db.exec(sql);
	}

	prepare(sql: string): PlatformStatement {
		return new NodeStatement(this.db.prepare(sql));
	}

	transaction<T>(fn: () => T): T {
		this.db.exec("BEGIN IMMEDIATE");
		try {
			const result = fn();
			if (isAsyncResult(result)) {
				throw new TypeError("SQLite transaction callbacks must be synchronous");
			}
			this.db.exec("COMMIT");
			return result;
		} catch (error) {
			try {
				this.db.exec("ROLLBACK");
			} catch {
				// 忽略回滚失败,重抛原始错误
			}
			throw error;
		}
	}

	close(): void {
		this.db.close();
	}
}

/** 打开（必要时创建）平台数据库：启用外键并幂等应用迁移。 */
export function openPlatformDatabase(path: string): PlatformDatabase {
	const db = new NodePlatformDatabase(new DatabaseSync(path));
	db.exec("PRAGMA foreign_keys = ON");
	applyMigrations(db);
	return db;
}

export interface PlatformMigration {
	id: string;
	order: number;
	sql: string;
}

function loadMigrationsSync(): PlatformMigration[] {
	// 迁移文件随包发布;源码运行(tsx)时相对 src,构建后需复制到 dist(见构建脚本 TODO)。
	const sql001 = readFileSync(new URL("./migrations/001_platform.sql", import.meta.url), "utf8");
	const sql002 = readFileSync(new URL("./migrations/002_auth_sessions.sql", import.meta.url), "utf8");
	const sql003 = readFileSync(new URL("./migrations/003_datasets_review.sql", import.meta.url), "utf8");
	const sql004 = readFileSync(new URL("./migrations/004_session_permission.sql", import.meta.url), "utf8");
	const sql005 = readFileSync(new URL("./migrations/005_agent_listings.sql", import.meta.url), "utf8");
	return [
		{ id: "001_platform.sql", order: 1, sql: sql001 },
		{ id: "002_auth_sessions.sql", order: 2, sql: sql002 },
		{ id: "003_datasets_review.sql", order: 3, sql: sql003 },
		{ id: "004_session_permission.sql", order: 4, sql: sql004 },
		{ id: "005_agent_listings.sql", order: 5, sql: sql005 },
	];
}

function ensureMigrationsTable(db: PlatformDatabase): void {
	db.exec(`
CREATE TABLE IF NOT EXISTS migrations (
	id TEXT PRIMARY KEY,
	applied_at TEXT NOT NULL
);
`);
}

/** 幂等应用未执行的迁移。 */
export function applyMigrations(db: PlatformDatabase): void {
	ensureMigrationsTable(db);
	const migrations = loadMigrationsSync();
	const appliedRows = db.prepare("SELECT id FROM migrations ORDER BY applied_at, id").all<{ id: string }>();
	const applied = new Set(appliedRows.map((row) => row.id));

	for (const migration of migrations) {
		if (applied.has(migration.id)) continue;
		db.transaction(() => {
			db.exec(migration.sql);
			db.prepare("INSERT INTO migrations (id, applied_at) VALUES (?, ?)").run(
				migration.id,
				new Date().toISOString(),
			);
		});
		applied.add(migration.id);
	}
}
