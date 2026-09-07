<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";

const props = defineProps<{
	modelValue: string;
	language?: string;
	readonly?: boolean;
}>();

const emit = defineEmits<{
	"update:modelValue": [value: string];
}>();

const container = ref<HTMLElement | null>(null);
let disposed = false;
let editor: unknown = null;
let monacoApi: typeof import("monaco-editor/editor/editor.api") | null = null;

onMounted(async () => {
	// 懒加载 Monaco:首次打开代码文件才加载,避免拖慢首屏
	const api = await import("monaco-editor/editor/editor.api");
	const { default: EditorWorker } = await import("monaco-editor/editor/editor.worker?worker");
	(self as unknown as { MonacoEnvironment?: unknown }).MonacoEnvironment = {
		getWorker: () => new EditorWorker(),
	};
	if (disposed || !container.value) return;
	monacoApi = api;
	editor = api.editor.create(container.value, {
		value: props.modelValue,
		language: props.language ?? "plaintext",
		readOnly: props.readonly ?? false,
		automaticLayout: true,
		theme: "vs-dark",
		fontSize: 13,
		scrollBeyondLastLine: false,
		minimap: { enabled: false },
	});
	(editor as { onDidChangeModelContent: (fn: () => void) => void }).onDidChangeModelContent(() => {
		emit("update:modelValue", ((editor as { getValue: () => string })?.getValue() ?? "") as string);
	});
});

watch(
	() => props.modelValue,
	(value) => {
		const current = (editor as { getValue?: () => string })?.getValue?.();
		if (editor && current !== value) (editor as { setValue: (v: string) => void }).setValue(value);
	},
);

watch(
	() => props.language,
	(value) => {
		if (!editor || !monacoApi) return;
		const model = (editor as { getModel: () => { setLanguage?: never } | null }).getModel?.();
		if (model) monacoApi.editor.setModelLanguage(model as never, value ?? "plaintext");
	},
);

onBeforeUnmount(() => {
	disposed = true;
	(editor as { dispose?: () => void })?.dispose?.();
	editor = null;
});
</script>

<template>
	<div ref="container" class="monaco-host" />
</template>

<style scoped>
.monaco-host {
	width: 100%;
	height: 100%;
	min-height: 200px;
}
</style>
