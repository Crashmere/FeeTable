<script setup lang="ts">
import { ref, computed } from "vue";
import {
  api,
  ApiError,
  type Table,
  type FeeRecord,
  type Suggestion,
} from "../api";
import { calculate } from "../money";
import Modal from "./Modal.vue";
import SuggestInput from "./SuggestInput.vue";
const props = defineProps<{
  table: Table;
  record?: FeeRecord;
  locations: Suggestion[];
  tags: Suggestion[];
}>();
const emit = defineEmits<{ close: []; saved: []; suggestions: [] }>();
const day = ref(props.record?.day.toString() || "");
const location1 = ref(props.record?.location1 || "");
const location2 = ref(props.record?.location2 || "");
const quantity = ref(props.record?.quantity || "");
const unitPrice = ref(props.record?.unitPrice || "");
const amount = ref(props.record?.amount || "");
const tag = ref(props.record?.tag || "");
const busy = ref(false);
const error = ref("");
const uncertain = ref(false);
const computedAmount = computed(() =>
  unitPrice.value !== "" ? calculate(quantity.value, unitPrice.value) : null,
);
const days = computed(() => {
  const d = new Date(0);
  d.setUTCFullYear(props.table.year, props.table.month, 0);
  return d.getUTCDate();
});
async function add(kind: "locations" | "tags", name: string) {
  try {
    await api.addSuggestion(kind, name);
    emit("suggestions");
  } catch (e) {
    error.value = (e as Error).message;
  }
}
async function save() {
  if (busy.value || uncertain.value) return;
  busy.value = true;
  error.value = "";
  try {
    await api.saveRecord(props.table, props.record?.id || null, {
      day: Number(day.value),
      location1: location1.value,
      location2: location2.value,
      quantity: quantity.value,
      unitPrice: unitPrice.value === "" ? null : unitPrice.value,
      amount:
        unitPrice.value === "" ? amount.value : computedAmount.value || "",
      tag: tag.value.trim() || null,
      revision: props.table.revision,
    });
    emit("saved");
  } catch (e) {
    error.value = (e as Error).message;
    uncertain.value =
      e instanceof ApiError && (e.uncertain || e.code === "CONFLICT");
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <Modal
    :title="record ? '编辑运输记录' : '添加运输记录'"
    :busy="busy"
    @close="emit('close')"
    ><form @submit.prevent="save" class="record-form">
      <p class="form-context">{{ table.year }} 年 {{ table.month }} 月</p>
      <label
        >日期 · 日<input
          v-model="day"
          type="number"
          min="1"
          :max="days"
          inputmode="numeric"
          placeholder="输入日期"
          required
      /></label>
      <div class="form-row">
        <SuggestInput
          v-model="location1"
          :items="locations"
          label="地点 1"
          kind="locations"
          @add="add('locations', $event)"
        /><SuggestInput
          v-model="location2"
          :items="locations"
          label="地点 2"
          kind="locations"
          @add="add('locations', $event)"
        />
      </div>
      <div class="form-row">
        <label
          >运输数量<input
            v-model="quantity"
            inputmode="decimal"
            placeholder="0.000"
            required /></label
        ><label
          >单价（元）<input
            v-model="unitPrice"
            inputmode="numeric"
            placeholder="留空为固定费用"
        /></label>
      </div>
      <div class="amount-field">
        <label
          >{{
            unitPrice === "" ? "运输金额 · 固定费用" : "运输金额 · 自动计算"
          }}
          <div class="currency-input">
            <span>¥</span
            ><input
              v-if="unitPrice === ''"
              v-model="amount"
              aria-label="运输金额"
              inputmode="decimal"
              placeholder="0.00"
              required
            /><output v-else>{{ computedAmount ?? "—" }}</output>
          </div></label
        ><small>{{
          unitPrice === ""
            ? "输入本次运输的固定金额"
            : "数量 × 单价，四舍五入到分"
        }}</small>
      </div>
      <SuggestInput
        v-model="tag"
        :items="tags"
        label="标签（可选）"
        kind="tags"
        @add="add('tags', $event)"
      />
      <p v-if="error" class="notice error" role="alert">{{ error }}</p>
      <div class="modal-actions">
        <button type="button" @click="emit('close')" :disabled="busy">
          {{ uncertain ? "返回核对" : "取消" }}</button
        ><button class="primary" :disabled="busy || uncertain">保存记录</button>
      </div>
    </form></Modal
  >
</template>
