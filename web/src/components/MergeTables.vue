<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { api, ApiError, type Table, type TableList } from "../api";
import { money } from "../money";
import Modal from "./Modal.vue";
const props = defineProps<{ target: Table }>();
const emit = defineEmits<{ close: []; merged: [table: Table] }>();
const data = ref<TableList>();
const page = ref(1);
const loading = ref(true);
const error = ref("");
const selected = ref<Table[]>([]);
const confirming = ref(false);
const busy = ref(false);
const uncertain = ref(false);
const count = computed(() =>
  selected.value.reduce((n, t) => n + t.recordCount, props.target.recordCount),
);
const total = computed(() => {
  const cents = [props.target, ...selected.value].reduce(
    (sum, t) => sum + BigInt(t.total.replace(".", "")),
    0n,
  );
  const digits = (cents < 0n ? -cents : cents).toString().padStart(3, "0");
  return (cents < 0n ? "-" : "") + digits.slice(0, -2) + "." + digits.slice(-2);
});
async function load() {
  loading.value = true;
  error.value = "";
  try {
    data.value = await api.tables(
      page.value,
      "updated_desc",
      props.target.year,
      props.target.month,
    );
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
onMounted(load);
function toggle(table: Table) {
  const index = selected.value.findIndex((t) => t.id === table.id);
  if (index >= 0) selected.value.splice(index, 1);
  else selected.value.push(table);
}
function turn(n: number) {
  page.value = n;
  load();
}
async function merge() {
  if (
    busy.value ||
    uncertain.value ||
    !selected.value.length ||
    count.value > 5000
  )
    return;
  busy.value = true;
  error.value = "";
  try {
    const table = await api.mergeTables(props.target, selected.value);
    busy.value = false;
    emit("merged", table);
  } catch (e) {
    error.value = (e as Error).message;
    uncertain.value =
      e instanceof ApiError &&
      (e.uncertain || e.code === "CONFLICT" || e.code === "NOT_FOUND");
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <Modal
    :title="confirming ? '确认合并运费表' : '合并同月表格'"
    :busy="busy"
    @close="emit('close')"
  >
    <p>{{ target.year }} 年 {{ target.month }} 月 · 保留表 #{{ target.id }}</p>
    <template v-if="!confirming">
      <div v-if="loading" class="notice" role="status">正在加载表格…</div>
      <template v-else-if="data">
        <div class="merge-list">
          <label
            v-for="t in data.items.filter((t) => t.id !== target.id)"
            :key="t.id"
            class="merge-choice"
          >
            <input
              type="checkbox"
              :checked="selected.some((s) => s.id === t.id)"
              :disabled="
                selected.length >= 99 && !selected.some((s) => s.id === t.id)
              "
              @change="toggle(t)"
            />
            <span
              ><strong>表 #{{ t.id }}</strong
              ><small>{{ t.recordCount }} 条 · ¥ {{ money(t.total) }}</small
              ><small
                >更新于
                {{ new Date(t.updatedAt).toLocaleString("zh-CN") }}</small
              ></span
            >
          </label>
        </div>
        <p v-if="data.total <= 1" class="subtle">没有其他同年月表格</p>
        <div v-if="data.total > 30" class="pagination">
          <button :disabled="page === 1" @click="turn(page - 1)">上一页</button
          ><span>{{ page }} / {{ Math.ceil(data.total / 30) }}</span
          ><button :disabled="page * 30 >= data.total" @click="turn(page + 1)">
            下一页
          </button>
        </div>
      </template>
    </template>
    <template v-else>
      <p class="wrap-text">
        将表 {{ selected.map((t) => "#" + t.id).join("、") }} 的全部记录移入表
        #{{
          target.id
        }}，随后移除这些来源表。重复内容也会保留，合并后无法直接拆回。
      </p>
    </template>
    <p v-if="selected.length" class="merge-summary">
      {{ selected.length + 1 }} 张表 · {{ count }} 条记录 · ¥ {{ money(total) }}
    </p>
    <p v-if="count > 5000" class="notice error" role="alert">
      合并后超过 5000 条记录，请减少所选表格。
    </p>
    <p v-if="error" class="notice error" role="alert">
      {{ error }} <button v-if="!confirming" @click="load">重试</button>
    </p>
    <div class="modal-actions">
      <button
        :disabled="busy"
        @click="confirming && !uncertain ? (confirming = false) : emit('close')"
      >
        {{ uncertain ? "返回核对" : confirming ? "返回选择" : "取消" }}
      </button>
      <button
        v-if="!confirming"
        class="primary"
        :disabled="loading || !!error || !selected.length || count > 5000"
        @click="confirming = true"
      >
        下一步
      </button>
      <button
        v-else
        class="primary"
        :disabled="busy || uncertain"
        @click="merge"
      >
        确认合并
      </button>
    </div>
  </Modal>
</template>
