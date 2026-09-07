/**
 * 金融数据源抽象(占位,Phase 6 细化;具体数据源后期实现)。
 */

export interface Quote {
	symbol: string;
	name?: string;
	price: number;
	currency: string;
	timestamp: number;
}

export interface FinancialReport {
	symbol: string;
	period: string;
	/** 报表结构占位:后期按需扩充 */
	fields: Record<string, unknown>;
}

export interface MarketDataSource {
	readonly name: string;
	getQuote(symbol: string, opts?: { signal?: AbortSignal }): Promise<Quote>;
	getQuotes(symbols: string[], opts?: { signal?: AbortSignal }): Promise<Quote[]>;
	getFinancialReport(symbol: string, period?: string, opts?: { signal?: AbortSignal }): Promise<FinancialReport>;
}

/** 数据源不可用/数据缺失时的标准失败面,供工具执行时抛出。 */
export class DataSourceUnavailableError extends Error {
	constructor(message: string) {
		super(message);
		this.name = "DataSourceUnavailableError";
	}
}
