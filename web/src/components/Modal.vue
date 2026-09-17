<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from "vue";
import { onBeforeRouteLeave } from "vue-router";
import Icon from "./Icon.vue";
const props = defineProps<{ title: string; busy?: boolean }>();
const emit = defineEmits<{ close: [] }>();
const dialog = ref<HTMLDialogElement>();
onMounted(() => dialog.value?.showModal());
onBeforeUnmount(() => dialog.value?.close());
function cancel(e: Event) {
  e.preventDefault();
  if (!props.busy) emit("close");
}
function leaving(e: BeforeUnloadEvent) {
  if (props.busy) {
    e.preventDefault();
    e.returnValue = "";
  }
}
onMounted(() => window.addEventListener("beforeunload", leaving));
onBeforeUnmount(() => window.removeEventListener("beforeunload", leaving));
onBeforeRouteLeave(() => !props.busy);
</script>
<template>
  <dialog ref="dialog" @cancel="cancel" :aria-label="title">
    <div class="modal-head">
      <h2>{{ title }}</h2>
      <button
        class="icon-button"
        type="button"
        aria-label="关闭"
        :disabled="busy"
        @click="emit('close')"
      >
        <Icon name="close" />
      </button>
    </div>
    <slot />
    <div v-if="busy" class="saving-overlay" role="status">
      <span class="spinner" />正在保存，请稍候
    </div>
  </dialog>
</template>
