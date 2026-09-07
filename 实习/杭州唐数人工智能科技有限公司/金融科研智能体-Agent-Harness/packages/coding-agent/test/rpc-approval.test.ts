import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, test } from "vitest";
import { RpcClient } from "../src/modes/rpc/rpc-client.ts";
import type { RpcApprovalRequest } from "../src/modes/rpc/rpc-types.ts";

const tempDirs: string[] = [];

function writeChildScript(contents: string): string {
	const dir = mkdtempSync(join(tmpdir(), "tfa-rpc-approval-"));
	tempDirs.push(dir);
	const path = join(dir, "child.mjs");
	writeFileSync(path, contents);
	return path;
}

afterEach(() => {
	for (const dir of tempDirs.splice(0)) {
		rmSync(dir, { recursive: true, force: true });
	}
});

const childScript = `
import fs from "node:fs";
const recordFile = process.env.FAKE_RPC_APPROVAL_FILE;
process.stdin.on("data", (chunk) => {
  for (const line of chunk.toString().split("\\n")) {
    if (!line.trim()) continue;
    let cmd;
    try { cmd = JSON.parse(line); } catch { continue; }
    if (cmd.type === "prompt") {
      process.stdout.write(JSON.stringify({ type: "approval_request", id: "appr-1", tool: "write", target: "a.txt", risk: "修改项目文件" }) + "\\n");
      process.stdout.write(JSON.stringify({ id: cmd.id, type: "response", command: "prompt", success: true }) + "\\n");
    } else if (cmd.type === "approval_response") {
      if (recordFile) {
        fs.appendFileSync(recordFile, JSON.stringify({ id: cmd.id, approved: cmd.approved }) + "\\n");
      }
    } else {
      process.stdout.write(JSON.stringify({ id: cmd.id, type: "response", command: cmd.type, success: true }) + "\\n");
    }
  }
});
process.stdin.resume();
`;

const sleep = (ms: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, ms));

describe("RPC approval round-trip", () => {
	test("onApproval 收到 approval_request,respondApproval 回传决策", async () => {
		const dir = mkdtempSync(join(tmpdir(), "tfa-rpc-approval-record-"));
		tempDirs.push(dir);
		const recordFile = join(dir, "approval.jsonl");
		const client = new RpcClient({
			cliPath: writeChildScript(childScript),
			cwd: dir,
			env: { FAKE_RPC_APPROVAL_FILE: recordFile },
		});
		await client.start();

		let received: RpcApprovalRequest | undefined;
		client.onApproval((request) => {
			received = request;
		});

		await client.prompt("hi");
		await sleep(50);

		expect(received).toBeDefined();
		expect(received?.tool).toBe("write");
		expect(received?.target).toBe("a.txt");

		await client.respondApproval(received!.id, true);
		await sleep(50);
		const record = JSON.parse(readFileSync(recordFile, "utf8").trim()) as { id: string; approved: boolean };
		expect(record.id).toBe(received!.id);
		expect(record.approved).toBe(true);

		await client.stop();
	});

	test("respondApproval 拒绝决策被回传", async () => {
		const dir = mkdtempSync(join(tmpdir(), "tfa-rpc-approval-record-"));
		tempDirs.push(dir);
		const recordFile = join(dir, "approval.jsonl");
		const client = new RpcClient({
			cliPath: writeChildScript(childScript),
			cwd: dir,
			env: { FAKE_RPC_APPROVAL_FILE: recordFile },
		});
		await client.start();

		await client.prompt("hi");
		await sleep(50);
		await client.respondApproval("appr-1", false);
		await sleep(50);
		const record = JSON.parse(readFileSync(recordFile, "utf8").trim()) as { approved: boolean };
		expect(record.approved).toBe(false);

		await client.stop();
	});
});
