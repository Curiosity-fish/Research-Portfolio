<script setup lang="ts">
/**
 * 统一线性图标集(参考 Lucide 风格,stroke 1.75,24x24 viewBox)。
 * 零依赖实现,避免引入 npm 包与 lockfile 变更。
 */

export type IconName =
	| "plus"
	| "arrow-up"
	| "stop"
	| "wrench"
	| "sliders"
	| "sun"
	| "moon"
	| "search"
	| "book"
	| "zap"
	| "message"
	| "history"
	| "pen"
	| "file-text"
	| "trending"
	| "chart"
	| "chevron-right"
	| "chevron-down"
	| "arrow-right"
	| "folder"
	| "home"
	| "x"
	| "more"
	| "paperclip"
	| "check"
	| "upload"
	| "download"
	| "trash"
	| "users"
	| "shield"
	| "eye"
	| "refresh"
	| "database"
	| "layers"
	| "terminal"
	| "archive"
	| "globe"
	| "lock"
	| "box";

const PATHS: Record<IconName, string[]> = {
	plus: ["M12 5v14", "M5 12h14"],
	"arrow-up": ["M12 19V5", "m5 12 7-7 7 7"],
	stop: ["M6 6h12v12H6z"],
	wrench: [
		"M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z",
	],
	sliders: [
		"M4 21v-7",
		"M4 10V3",
		"M12 21v-9",
		"M12 8V3",
		"M20 21v-5",
		"M20 12V3",
		"M2 14h4",
		"M10 8h4",
		"M18 16h4",
	],
	sun: [
		"M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8Z",
		"M12 2v2",
		"M12 20v2",
		"M4.93 4.93l1.41 1.41",
		"M17.66 17.66l1.41 1.41",
		"M2 12h2",
		"M20 12h2",
		"M6.34 17.66l-1.41 1.41",
		"M19.07 4.93l-1.41 1.41",
	],
	moon: ["M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"],
	search: ["M11 3a8 8 0 1 0 0 16 8 8 0 0 0 0-16Z", "M21 21l-4.35-4.35"],
	book: [
		"M4 19.5A2.5 2.5 0 0 1 6.5 17H20",
		"M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2Z",
	],
	zap: ["M13 2 3 14h9l-1 8 10-12h-9l1-8z"],
	message: ["M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"],
	history: [
		"M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8",
		"M3 3v5h5",
		"M12 7v5l4 2",
	],
	pen: ["M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"],
	"file-text": [
		"M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z",
		"M14 2v6h6",
		"M16 13H8",
		"M16 17H8",
		"M10 9H8",
	],
	trending: ["M22 7l-8.5 8.5-5-5L2 17", "M16 7h6v6"],
	chart: ["M3 3v16a2 2 0 0 0 2 2h16", "M18 17V9", "M13 17V5", "M8 17v-3"],
	"chevron-right": ["m9 18 6-6-6-6"],
	"chevron-down": ["m6 9 6 6 6-6"],
	"arrow-right": ["M5 12h14", "m12 5 7 7-7 7"],
	folder: [
		"M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z",
	],
	home: ["M3 10.5 12 3l9 7.5", "M5 9.5V21h14V9.5", "M9 21v-6h6v6"],
	x: ["M18 6 6 18", "M6 6l12 12"],
	more: ["M5 12h.01", "M12 12h.01", "M19 12h.01"],
	paperclip: [
		"m21.44 11.05-9.19 9.19a6 6 0 0 1-8.49-8.49l8.57-8.57A4 4 0 1 1 18 8.84l-8.59 8.57a2 2 0 0 1-2.83-2.83l8.49-8.48",
	],
	check: ["M20 6 9 17l-5-5"],
	upload: ["M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4", "m17 8-5-5-5 5", "M12 3v12"],
	download: ["M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4", "m7 10 5 5 5-5", "M12 15V3"],
	trash: ["M3 6h18", "M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6", "M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2", "M10 11v6", "M14 11v6"],
	users: ["M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2", "M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z", "M22 21v-2a4 4 0 0 0-3-3.87", "M16 3.13a4 4 0 0 1 0 7.75"],
	shield: ["M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"],
	eye: ["M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z", "M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z"],
	refresh: ["M3 12a9 9 0 0 1 15-6.7L21 8", "M21 3v5h-5", "M21 12a9 9 0 0 1-15 6.7L3 16", "M3 21v-5h5"],
	database: ["M12 2c4.4 0 8 1.8 8 4s-3.6 4-8 4-8-1.8-8-4 3.6-4 8-4Z", "M4 6v12c0 2.2 3.6 4 8 4s8-1.8 8-4V6", "M4 12c0 2.2 3.6 4 8 4s8-1.8 8-4"],
	layers: ["m12 2 10 6-10 6L2 8l10-6Z", "m2 16 10 6 10-6"],
	terminal: ["m4 17 6-6-6-6", "M12 19h8"],
	archive: ["M21 8v13H3V8", "M1 3h22v5H1z", "M10 12h4"],
	globe: ["M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20Z", "M2 12h20", "M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"],
	lock: ["M19 11H5a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7a2 2 0 0 0-2-2Z", "M7 11V7a5 5 0 0 1 10 0v4"],
	box: ["M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z", "M3.3 7 12 12l8.7-5", "M12 22V12"],
};

withDefaults(defineProps<{ name: IconName; size?: number }>(), { size: 18 });
</script>

<template>
	<svg
		:width="size"
		:height="size"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="1.75"
		stroke-linecap="round"
		stroke-linejoin="round"
		aria-hidden="true"
	>
		<path v-for="(d, index) in PATHS[name]" :key="index" :d="d" />
	</svg>
</template>