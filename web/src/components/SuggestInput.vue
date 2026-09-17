<script setup lang="ts">
import { ref, computed } from "vue";
import type { Suggestion } from "../api";
const props = defineProps<{
  modelValue: string;
  items: Suggestion[];
  label: string;
  kind: "locations" | "tags";
}>();
const emit = defineEmits<{
  "update:modelValue": [value: string];
  add: [value: string];
}>();
const open = ref(false);
const filtered = computed(() =>
  props.items
    .filter((i) =>
      i.name.toLowerCase().includes(props.modelValue.toLowerCase()),
    )
    .slice(0, 6),
);
const canAdd = computed(
  () =>
    props.modelValue.trim() &&
    !props.items.some((i) => i.name === props.modelValue.trim()),
);
function select(value: string) {
  emit("update:modelValue", value);
  open.value = false;
}
function blur() {
  setTimeout(() => {
    open.value = false;
  }, 120);
}
</script>
<template>
  <label class="suggest-field"
    >{{ label
    }}<input
      :value="modelValue"
      maxlength="80"
      autocomplete="off"
      :required="kind === 'locations'"
      @input="
        emit('update:modelValue', ($event.target as HTMLInputElement).value);
        open = true;
      "
      @focus="open = true"
      @blur="blur"
      @keydown.esc="open = false"
    />
    <div v-if="open && (filtered.length || canAdd)" class="suggest-menu">
      <button
        v-for="item in filtered"
        :key="item.id"
        type="button"
        @mousedown.prevent
        @click="select(item.name)"
      >
        {{ item.name }}</button
      ><button
        v-if="canAdd"
        type="button"
        class="suggest-add"
        @mousedown.prevent
        @click="
          emit('add', modelValue.trim());
          select(modelValue.trim());
        "
      >
        ＋ 添加「{{ modelValue.trim() }}」
      </button>
    </div></label
  >
</template>
