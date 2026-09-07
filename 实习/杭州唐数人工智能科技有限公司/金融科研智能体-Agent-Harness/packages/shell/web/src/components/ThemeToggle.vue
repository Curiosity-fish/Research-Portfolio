<script setup lang="ts">
import { onMounted, ref } from "vue";
import AppIcon from "./AppIcon.vue";

const dark = ref(false);

onMounted(() => {
	const saved = localStorage.getItem("tfa-theme");
	dark.value = saved === "dark";
	apply();
});

function toggle(): void {
	dark.value = !dark.value;
	localStorage.setItem("tfa-theme", dark.value ? "dark" : "light");
	apply();
}

function apply(): void {
	document.documentElement.dataset.theme = dark.value ? "dark" : "light";
	const meta = document.querySelector('meta[name="theme-color"]');
	if (meta) meta.setAttribute("content", dark.value ? "#0b141d" : "#f7fafc");
}
</script>

<template>
	<button
		class="icon-btn bordered"
		type="button"
		:title="dark ? '切换到浅色' : '切换到深色'"
		:aria-label="dark ? '切换到浅色' : '切换到深色'"
		@click="toggle"
	>
		<AppIcon :name="dark ? 'sun' : 'moon'" :size="17" />
	</button>
</template>
