<script setup lang="ts">
import { ref } from "vue";
import type { ModelMetadata, ModelRef, SessionSummary, UserInfo } from "../api/types.ts";
import type { ProjectWithSessions } from "../state/projects.ts";
import AppIcon from "./AppIcon.vue";
import PromptBox from "./PromptBox.vue";

const QUICK_PROMPTS = [
	"给我调研一下股票情况",
	"分析一下公司基本面",
	"解读一下最近财报",
	"研究一下某个行业",
];

const props = defineProps<{
	projects: ProjectWithSessions[];
	totalProjects: number;
	sessions: SessionSummary[];
	totalSessions: number;
	skillCount: number;
	datasetCount: number;
	user: UserInfo | null;
	connected: boolean;
	models: ModelMetadata[];
	currentModel: ModelRef | null;
}>();

const emit = defineEmits<{
	send: [text: string];
	"open-session": [sessionId: string];
	"new-chat-in-project": [projectId: string];
	"create-project": [];
}>();

const draft = ref("");
const composerRef = ref<{ focus(): void } | null>(null);

function useQuickPrompt(text: string): void {
	draft.value = text;
	composerRef.value?.focus();
}

function openProject(group: ProjectWithSessions): void {
	const session = group.sessions[0];
	if (session) emit("open-session", session.id);
	else emit("new-chat-in-project", group.project.id);
}

function formatTime(timestamp: number): string {
	const date = new Date(timestamp);
	const now = new Date();
	return date.toDateString() === now.toDateString()
		? date.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" })
		: date.toLocaleDateString("zh-CN", { month: "2-digit", day: "2-digit" });
}
</script>

<template>
	<section class="home" aria-label="金融研究工作台">
		<div class="cockpit">
			<div class="cockpit-main">
				<div class="hero">
					<img class="hero-logo" src="/logo.png" alt="金融科技 Agent" />
					<p class="hero-kicker">Research Cockpit</p>
					<h1 class="hero-title">今天要研究什么？</h1>
					<p class="hero-desc">输入一个公司、行业、财报或市场问题，Agent 会调用工具完成研究。</p>
				</div>

				<div class="composer-shell">
					<PromptBox
						ref="composerRef"
						:busy="false"
						:models="props.models"
						:current="props.currentModel"
						v-model="draft"
						placeholder="输入研究问题，例如：对比宁德时代与比亚迪最近一季的经营质量"
						autofocus
						@send="emit('send', $event)"
					/>
					<div class="quick-prompts" aria-label="常用研究指令">
						<button
							v-for="prompt in QUICK_PROMPTS"
							:key="prompt"
							class="quick-prompt"
							type="button"
							@click="useQuickPrompt(prompt)"
						>
							{{ prompt }}
						</button>
					</div>
				</div>

				<section class="continue-panel" aria-label="继续研究">
					<div class="panel-head">
						<h2>继续研究</h2>
						<span>{{ totalSessions }} 个会话</span>
					</div>
					<div v-if="sessions.length === 0" class="empty-row">
						还没有研究会话。在上方输入第一个研究问题即可开始。
					</div>
					<button
						v-for="session in sessions.slice(0, 3)"
						:key="session.id"
						class="continue-row"
						type="button"
						@click="emit('open-session', session.id)"
					>
						<span class="continue-name">{{ session.name || session.id.slice(0, 8) }}</span>
						<span class="continue-time">{{ formatTime(session.updatedAt) }}</span>
						<AppIcon name="chevron-right" :size="14" />
					</button>
				</section>
			</div>

			<aside class="cockpit-side" aria-label="工作状态">
				<article class="side-card identity">
					<span class="identity-avatar">{{ user?.username.slice(0, 1).toUpperCase() || "A" }}</span>
					<div>
						<strong>{{ user?.username || "未登录" }}</strong>
						<small>{{ user?.role?.toUpperCase() || "GUEST" }}</small>
					</div>
					<span class="connection" :class="connected ? 'online' : 'offline'">
						{{ connected ? "已连接" : "离线" }}
					</span>
				</article>

				<article class="side-card">
					<h2>资源状态</h2>
					<div class="stat-grid">
						<div><small>项目</small><strong>{{ totalProjects }}</strong></div>
						<div><small>会话</small><strong>{{ totalSessions }}</strong></div>
						<div><small>技能</small><strong>{{ skillCount }}</strong></div>
						<div><small>数据集</small><strong>{{ datasetCount }}</strong></div>
					</div>
				</article>

				<article class="side-card">
					<h2>下一步</h2>
					<button class="action-row primary" type="button" @click="emit('create-project')">
						<AppIcon name="plus" :size="14" />
						创建研究项目
					</button>
					<button class="action-row" type="button" disabled>
						<AppIcon name="database" :size="14" />
						上传数据集
					</button>
					<button class="action-row" type="button" disabled>
						<AppIcon name="layers" :size="14" />
						安装技能
					</button>
				</article>
			</aside>
		</div>

		<section class="recent-strip" aria-label="最近项目">
			<div class="section-head">
				<h2>最近项目</h2>
				<span>{{ totalProjects }} 个项目</span>
				<button class="create-btn" type="button" @click="emit('create-project')">
					<AppIcon name="plus" :size="14" />
					创建项目
				</button>
			</div>
			<div v-if="projects.length === 0" class="empty-projects">
				还没有项目。创建项目后，可按公司、行业或课题组织会话和文件。
			</div>
			<div v-else class="project-grid">
				<button
					v-for="group in projects"
					:key="group.project.id"
					class="project-card"
					type="button"
					@click="openProject(group)"
				>
					<span class="project-top">
						<i class="project-icon" aria-hidden="true"><AppIcon name="folder" :size="15" /></i>
						<strong>{{ group.project.name }}</strong>
					</span>
					<span class="project-path mono">{{ group.project.folder || "未绑定文件夹" }}</span>
					<span class="project-meta">
						{{ group.project.sessionIds.length }} 个对话
						<span>{{ formatTime(group.project.updatedAt) }}</span>
					</span>
					<span
						class="project-add"
						role="button"
						tabindex="0"
						aria-label="在该项目新建对话"
						@click.stop="emit('new-chat-in-project', group.project.id)"
						@keydown.enter.prevent="emit('new-chat-in-project', group.project.id)"
					>
						<AppIcon name="plus" :size="14" />
					</span>
				</button>
			</div>
		</section>
	</section>
</template>

<style scoped>
.home {
	flex: 1;
	min-height: 0;
	display: flex;
	flex-direction: column;
	overflow-y: auto;
	background:
		radial-gradient(circle at 88% 8%, color-mix(in srgb, var(--accent) 8%, transparent), transparent 28rem),
		var(--bg-app);
}

.cockpit {
	display: grid;
	grid-template-columns: minmax(0, 1fr) 300px;
	gap: 32px;
	width: min(1480px, 100%);
	margin: 0 auto;
	padding: clamp(28px, 5vh, 56px) clamp(24px, 4vw, 56px) 16px;
}

.cockpit-main { min-width: 0; display: flex; flex-direction: column; gap: 28px; }
.hero { max-width: 720px; }
.hero-logo { width: 48px; height: 48px; object-fit: contain; margin-bottom: 16px; }
.hero-kicker { margin: 0 0 8px; font-size: 10px; font-weight: 700; letter-spacing: 0.18em; text-transform: uppercase; color: var(--accent); }
.hero-title { margin: 0; font-size: clamp(30px, 4vw, 42px); line-height: 1.08; font-weight: 750; letter-spacing: -0.055em; color: var(--text-primary); text-wrap: balance; }
.hero-desc { margin: 12px 0 0; max-width: 620px; font-size: 15px; line-height: 1.65; color: var(--text-secondary); }
.composer-shell { width: 100%; }
.quick-prompts { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 12px; }
.quick-prompt { height: 32px; padding: 0 11px; border: 1px solid var(--border); border-radius: var(--radius-sm); background: color-mix(in srgb, var(--bg-surface) 88%, var(--accent-soft)); color: var(--text-secondary); font-size: 12px; font-weight: 600; cursor: pointer; transition: all 150ms ease; }
.quick-prompt:hover { border-color: color-mix(in srgb, var(--accent) 38%, var(--border)); color: var(--accent-strong); transform: translateY(-1px); box-shadow: var(--shadow-float); }
.continue-panel { border-top: 1px solid var(--border); padding-top: 18px; }
.panel-head { display: flex; align-items: baseline; gap: 8px; margin-bottom: 10px; }
.panel-head h2 { margin: 0; font-size: 15px; font-weight: 700; color: var(--text-primary); }
.panel-head span { font-size: 11px; color: var(--text-muted); }
.continue-row { display: grid; grid-template-columns: minmax(0, 1fr) auto auto; align-items: center; gap: 10px; width: 100%; min-height: 44px; padding: 8px 10px; border: 0; border-bottom: 1px solid var(--border); background: transparent; color: var(--text-primary); text-align: left; cursor: pointer; }
.continue-row:last-of-type { border-bottom: 0; }
.continue-row:hover { background: color-mix(in srgb, var(--accent-soft) 42%, transparent); }
.continue-name { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; font-weight: 600; }
.continue-time { font-size: 11px; color: var(--text-muted); }
.empty-row { min-height: 58px; display: flex; align-items: center; color: var(--text-muted); font-size: 13px; }

.cockpit-side { display: flex; flex-direction: column; gap: 12px; }
.side-card { padding: 16px; border: 1px solid var(--border); border-radius: var(--radius-lg); background: color-mix(in srgb, var(--bg-surface) 94%, transparent); box-shadow: var(--shadow-card); }
.side-card h2 { margin: 0 0 12px; font-size: 10px; font-weight: 700; letter-spacing: 0.12em; text-transform: uppercase; color: var(--text-muted); }
.identity { display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: center; gap: 10px; }
.identity-avatar { width: 36px; height: 36px; display: inline-flex; align-items: center; justify-content: center; border-radius: 8px; background: linear-gradient(135deg, var(--accent), #6c99ff); color: #fff; font-weight: 700; }
.identity strong { display: block; font-size: 13px; color: var(--text-primary); }
.identity small { font-size: 9px; letter-spacing: 0.1em; color: var(--text-muted); }
.connection { grid-column: 1 / -1; justify-self: start; padding: 3px 7px; border-radius: var(--radius-xs); background: var(--accent-soft); color: var(--accent-strong); font-size: 10px; font-weight: 700; }
.connection.offline { background: var(--danger-bg); color: var(--danger-strong); }
.stat-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; }
.stat-grid div { padding: 10px; border-radius: var(--radius-sm); background: color-mix(in srgb, var(--accent-soft) 35%, var(--bg-surface)); }
.stat-grid small { display: block; font-size: 10px; color: var(--text-muted); }
.stat-grid strong { font-size: 20px; line-height: 1.2; color: var(--text-primary); }
.action-row { display: flex; align-items: center; gap: 8px; width: 100%; min-height: 36px; margin-top: 8px; padding: 0 10px; border: 1px solid var(--border); border-radius: var(--radius-sm); background: var(--bg-surface); color: var(--text-secondary); font-size: 12px; font-weight: 600; cursor: pointer; }
.action-row.primary { background: var(--accent-strong); border-color: var(--accent-strong); color: #fff; }
.action-row.primary:hover { filter: brightness(1.08); }
.action-row:disabled { opacity: 0.5; cursor: not-allowed; }

.recent-strip { width: min(1480px, 100%); margin: 0 auto; padding: 16px clamp(24px, 4vw, 56px) 32px; }
.section-head { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; }
.section-head h2 { margin: 0; font-size: 17px; font-weight: 700; letter-spacing: -0.02em; color: var(--text-primary); }
.section-head > span { font-size: 11px; color: var(--text-muted); }
.create-btn { display: inline-flex; align-items: center; gap: 6px; margin-left: auto; height: 32px; padding: 0 11px; border: 1px solid var(--border); border-radius: var(--radius-sm); background: var(--bg-surface); color: var(--text-secondary); font-size: 12px; font-weight: 600; cursor: pointer; }
.create-btn:hover { border-color: var(--accent); color: var(--accent-strong); }
.empty-projects { min-height: 88px; display: flex; align-items: center; justify-content: center; padding: 20px; border: 1px dashed var(--border-strong); border-radius: var(--radius-lg); color: var(--text-muted); background: color-mix(in srgb, var(--bg-surface) 60%, transparent); }
.project-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; }
.project-card { position: relative; min-height: 118px; display: flex; flex-direction: column; gap: 10px; padding: 15px; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--bg-surface); color: var(--text-primary); text-align: left; cursor: pointer; transition: all 170ms ease; }
.project-card:hover { transform: translateY(-2px); border-color: color-mix(in srgb, var(--accent) 45%, var(--border)); box-shadow: var(--shadow-card); }
.project-top { display: flex; align-items: center; gap: 9px; min-width: 0; padding-right: 28px; }
.project-icon { display: inline-flex; align-items: center; justify-content: center; width: 28px; height: 28px; flex-shrink: 0; border-radius: 7px; background: var(--accent-soft); color: var(--accent-strong); }
.project-top strong { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 14px; }
.project-path { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 10px; color: var(--text-muted); }
.project-meta { display: flex; align-items: center; justify-content: space-between; margin-top: auto; font-size: 11px; color: var(--text-muted); }
.project-add { position: absolute; top: 10px; right: 10px; width: 26px; height: 26px; display: inline-flex; align-items: center; justify-content: center; border-radius: 7px; color: var(--text-muted); transition: all 140ms ease; }
.project-add:hover, .project-add:focus-visible { background: var(--accent-soft); color: var(--accent-strong); }

@media (max-width: 1180px) {
	.cockpit { grid-template-columns: minmax(0, 1fr); gap: 20px; }
	.project-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (max-width: 680px) {
	.cockpit { padding-top: 20px; }
	.hero-title { font-size: 30px; }
	.project-grid { grid-template-columns: minmax(0, 1fr); }
	.section-head { flex-wrap: wrap; }
}
</style>
