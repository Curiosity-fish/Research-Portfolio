/** 数据集超限(单文件/每用户累计配额)。由 app.onError 映射为 413。 */
export class QuotaExceededError extends Error {
	readonly code = "quota_exceeded" as const;

	constructor(message: string) {
		super(message);
		this.name = "QuotaExceededError";
	}
}
