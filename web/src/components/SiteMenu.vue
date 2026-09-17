<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import Icon from "./Icon.vue";
const route = useRoute();
const open = ref(false);
const container = ref<HTMLElement>();
const trigger = ref<HTMLButtonElement>();
function outside(event: PointerEvent) {
  if (!container.value?.contains(event.target as Node)) open.value = false;
}
function escape(event: KeyboardEvent) {
  if (event.key === "Escape" && open.value) {
    open.value = false;
    trigger.value?.focus();
  }
}
function leave(event: FocusEvent) {
  if (!container.value?.contains(event.relatedTarget as Node))
    open.value = false;
}
watch(
  () => route.fullPath,
  () => {
    open.value = false;
  },
);
onMounted(() => {
  document.addEventListener("pointerdown", outside);
  document.addEventListener("keydown", escape);
});
onBeforeUnmount(() => {
  document.removeEventListener("pointerdown", outside);
  document.removeEventListener("keydown", escape);
});
</script>
<template>
  <div ref="container" class="site-menu" @focusout="leave">
    <button
      ref="trigger"
      type="button"
      class="icon-button menu-trigger"
      aria-label="打开导航菜单"
      aria-controls="site-navigation"
      :aria-expanded="open"
      @click="open = !open"
    >
      <Icon name="menu" :size="23" />
    </button>
    <nav
      v-if="open"
      id="site-navigation"
      class="menu-panel"
      aria-label="常用功能"
    >
      <RouterLink to="/locations" @click="open = false"
        ><Icon name="location" :size="18" />地点管理</RouterLink
      >
    </nav>
  </div>
</template>
