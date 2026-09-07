import { defineTool } from "@earendil-works/tfa-coding-agent";
import Type from "typebox";

/**
 * 示例金融工具:行情回显。
 *
 * 用途:验证"注册 → 注入会话 → CLI/Web 工具卡片展示"的完整链路。
 * 后期实现真实数据源后,以同类结构注册行情/财报/分析工具即可。
 */
export const echoMarketTool = defineTool({
	name: "echo_market",
	label: "行情回显(示例)",
	description: "示例金融工具:回显指定标的的行情占位数据,用于验证金融工具注册链路",
	promptSnippet: "echo_market: 回显指定标的的行情占位数据(示例工具)",
	parameters: Type.Object({
		symbol: Type.String({ minLength: 1, description: "标的代码,如 600519、AAPL" }),
		market: Type.Optional(Type.String({ description: "市场,如 sh/sz/us/hk" })),
	}),
	execute: async (_toolCallId, params) => {
		const market = params.market ?? "未知";
		return {
			content: [{ type: "text", text: `[echo_market] ${params.symbol}(${market}) 的行情数据待接入,当前为占位结果` }],
			details: { symbol: params.symbol, market, placeholder: true },
		};
	},
});
