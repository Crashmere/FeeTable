<script setup lang="ts">
defineProps<{ name: string; size?: number }>();
const paths: Record<string, string> = {
  plus: "M12 5v14M5 12h14",
  back: "m14 6-6 6 6 6",
  arrow: "M5 12h14m-5-5 5 5-5 5",
  download: "M12 3v12m-5-5 5 5 5-5M5 16v4h14v-4",
  edit: "m14 5 5 5M4 20l4-1L20 7a2 2 0 0 0-4-4L4 15v5Z",
  trash: "M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7m4-7v7",
  close: "m6 6 12 12M6 18 18 6",
  table: "M4 3h16v18H4zM4 9h16M4 15h16M10 9v12",
  tag: "M3 3h8l10 10-8 8L3 11V3Zm4 4h.01",
  check: "m5 12 4 4L19 6",
};
</script>
<template>
  <svg
    :width="size || 20"
    :height="size || 20"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.7"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path :d="paths[name] || paths.table" />
  </svg>
</template>
