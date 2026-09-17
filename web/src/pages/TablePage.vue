<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import {
  api,
  ApiError,
  query,
  type Report,
  type FeeRecord,
  type Suggestion,
} from "../api";
import { money } from "../money";
import Icon from "../components/Icon.vue";
import Modal from "../components/Modal.vue";
import RecordForm from "../components/RecordForm.vue";
const route = useRoute();
const id = String(route.params.id);
const data = ref<Report>();
const locations = ref<Suggestion[]>([]);
const tags = ref<Suggestion[]>([]);
const loading = ref(true);
const error = ref("");
const filter = ref<string | null>(null);
const page = ref(1);
const form = ref(false);
const editing = ref<FeeRecord>();
const deleting = ref<FeeRecord>();
const editMonth = ref(false);
const year = ref(2026);
const month = ref(1);
const busy = ref(false);
const formError = ref("");
const uncertain = ref(false);
let sequence = 0;
async function load() {
  const n = ++sequence;
  loading.value = true;
  error.value = "";
  try {
    const report = await api.report(id, filter.value, page.value);
    if (n === sequence) {
      const last = Math.max(1, Math.ceil(report.count / 50));
      if (page.value > last) {
        page.value = last;
        await load();
        return;
      }
      data.value = report;
    }
  } catch (e) {
    if (n === sequence) error.value = (e as Error).message;
  } finally {
    if (n === sequence) loading.value = false;
  }
}
async function loadSuggestions() {
  try {
    [locations.value, tags.value] = await Promise.all([
      api.suggestions("locations"),
      api.suggestions("tags"),
    ]);
  } catch (e) {
    error.value = (e as Error).message;
  }
}
onMounted(() => {
  load();
  loadSuggestions();
});
function applyFilter(event: Event) {
  filter.value = (event.target as HTMLSelectElement).value || null;
  page.value = 1;
  load();
}
function openForm(r?: FeeRecord) {
  editing.value = r;
  form.value = true;
}
async function saved() {
  form.value = false;
  await Promise.all([load(), loadSuggestions()]);
}
function openMonth() {
  if (!data.value) return;
  year.value = data.value.table.year;
  month.value = data.value.table.month;
  formError.value = "";
  uncertain.value = false;
  editMonth.value = true;
}
function askDelete(r: FeeRecord) {
  deleting.value = r;
  formError.value = "";
  uncertain.value = false;
}
async function changeMonth() {
  if (!data.value || busy.value || uncertain.value) return;
  busy.value = true;
  formError.value = "";
  try {
    await api.updateTable(
      data.value.table,
      Number(year.value),
      Number(month.value),
    );
    editMonth.value = false;
    await load();
  } catch (e) {
    formError.value = (e as Error).message;
    uncertain.value =
      e instanceof ApiError && (e.uncertain || e.code === "CONFLICT");
  } finally {
    busy.value = false;
  }
}
async function remove() {
  if (!data.value || !deleting.value || busy.value || uncertain.value) return;
  busy.value = true;
  try {
    await api.deleteRecord(data.value.table, deleting.value.id);
    deleting.value = undefined;
    await load();
  } catch (e) {
    formError.value = (e as Error).message;
    uncertain.value =
      e instanceof ApiError && (e.uncertain || e.code === "CONFLICT");
  } finally {
    busy.value = false;
  }
}
function turn(n: number) {
  page.value = n;
  load();
}
</script>
<template>
  <RouterLink class="back-link" to="/"
    ><Icon name="back" :size="17" />全部运费表</RouterLink
  >
  <section class="page-heading">
    <div>
      <p class="eyebrow">运输明细</p>
      <div class="title-line">
        <h1>
          {{
            data
              ? data.table.year + " 年 " + data.table.month + " 月"
              : "运费明细"
          }}
        </h1>
        <button
          v-if="data"
          class="icon-button"
          aria-label="修改年月"
          @click="openMonth"
        >
          <Icon name="edit" />
        </button>
      </div>
      <p class="subtle">管理每一笔运输，清晰掌握本月运费。</p>
    </div>
    <div class="heading-actions">
      <RouterLink
        v-if="data && data.table.recordCount"
        class="button"
        :to="'/tables/' + id + '/export' + query({ tag: filter })"
        ><Icon name="download" />导出表格</RouterLink
      ><button
        class="primary"
        @click="openForm()"
        :disabled="!data || loading || !!error"
      >
        <Icon name="plus" />添加记录
      </button>
    </div>
  </section>
  <div v-if="error" class="notice error" role="alert">
    {{ error }} <button @click="load">重试</button>
  </div>
  <template v-else>
    <div v-if="data" class="summary-bar">
      <div>
        <span class="summary-label">{{
          filter ? "当前标签合计" : "本月运费合计"
        }}</span>
        <p class="summary-amount"><small>¥</small>{{ money(data.total) }}</p>
      </div>
      <div class="summary-count">
        <strong>{{ data.count }}</strong
        ><span>条运输记录</span>
      </div>
      <div class="summary-decoration"><Icon name="table" :size="64" /></div>
    </div>
    <div class="records-toolbar">
      <h2>运输记录</h2>
      <label class="filter-select"
        ><Icon name="tag" :size="17" /><select
          :value="filter || ''"
          @change="applyFilter"
          aria-label="按标签筛选"
        >
          <option value="">全部标签</option>
          <option v-for="t in data?.tags || []" :key="t">{{ t }}</option>
        </select></label
      >
    </div>
    <div
      v-if="loading"
      class="skeleton skeleton-table"
      aria-label="正在加载记录"
    />
    <template v-else-if="data"
      ><div v-if="data.records.length" class="records-panel">
        <table class="records-table">
          <thead>
            <tr>
              <th>日期</th>
              <th>运输路线</th>
              <th class="numeric">数量</th>
              <th class="numeric">单价</th>
              <th class="numeric">金额</th>
              <th>标签</th>
              <th><span class="sr-only">操作</span></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in data.records" :key="r.id">
              <td class="record-date">{{ data.table.month }}/{{ r.day }}</td>
              <td class="record-route">
                <span>{{ r.location1 }}</span
                ><Icon name="arrow" :size="14" /><span>{{ r.location2 }}</span>
              </td>
              <td class="numeric record-quantity">
                <span class="mobile-label">数量 </span>{{ r.quantity }}
              </td>
              <td class="numeric record-price">
                <span class="mobile-label">单价 </span
                >{{ r.unitPrice ?? "固定费用" }}
              </td>
              <td class="numeric record-amount">¥ {{ money(r.amount) }}</td>
              <td class="record-tag">
                <span v-if="r.tag" class="tag">{{ r.tag }}</span
                ><span v-else class="subtle desktop-only">—</span>
              </td>
              <td class="record-actions">
                <button
                  class="icon-button"
                  aria-label="编辑记录"
                  @click="openForm(r)"
                >
                  <Icon name="edit" :size="17" /></button
                ><button
                  class="icon-button"
                  aria-label="删除记录"
                  @click="askDelete(r)"
                >
                  <Icon name="trash" :size="17" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty-state compact">
        <span class="empty-icon"><Icon name="table" :size="36" /></span>
        <h2>{{ filter ? "这个标签下暂无记录" : "还没有运输记录" }}</h2>
        <p>添加日期、地点和运费，合计会自动更新。</p>
        <button class="primary" @click="openForm()">
          <Icon name="plus" />添加记录
        </button>
      </div>
      <div v-if="data.count > 50" class="pagination">
        <button :disabled="page === 1" @click="turn(page - 1)">上一页</button
        ><span>{{ page }} / {{ Math.ceil(data.count / 50) }}</span
        ><button :disabled="page * 50 >= data.count" @click="turn(page + 1)">
          下一页
        </button>
      </div></template
    ></template
  >
  <RecordForm
    v-if="form && data"
    :table="data.table"
    :record="editing"
    :locations="locations"
    :tags="tags"
    @close="
      form = false;
      load();
    "
    @saved="saved"
    @suggestions="loadSuggestions"
  />
  <Modal
    v-if="editMonth"
    title="修改表格年月"
    :busy="busy"
    @close="
      editMonth = false;
      load();
    "
    ><form @submit.prevent="changeMonth">
      <p class="subtle">表内全部记录会归入新的年月，日期中的“日”保持不变。</p>
      <div class="form-row">
        <label
          >年份<input
            v-model="year"
            type="number"
            min="1"
            max="9999"
            required /></label
        ><label
          >月份<select v-model="month">
            <option v-for="m in 12" :key="m" :value="m">{{ m }} 月</option>
          </select></label
        >
      </div>
      <p v-if="formError" class="notice error" role="alert">{{ formError }}</p>
      <div class="modal-actions">
        <button
          type="button"
          @click="
            editMonth = false;
            load();
          "
          :disabled="busy"
        >
          取消</button
        ><button class="primary" :disabled="busy || uncertain">保存年月</button>
      </div>
    </form></Modal
  >
  <Modal
    v-if="deleting"
    title="删除运输记录"
    :busy="busy"
    @close="
      deleting = undefined;
      load();
    "
    ><p>
      确定删除 {{ deleting.location1 }} → {{ deleting.location2 }} 的这条记录？
    </p>
    <p v-if="formError" class="notice error" role="alert">{{ formError }}</p>
    <div class="modal-actions">
      <button
        @click="
          deleting = undefined;
          load();
        "
        :disabled="busy"
      >
        取消</button
      ><button class="danger" @click="remove" :disabled="busy || uncertain">
        确认删除
      </button>
    </div></Modal
  >
</template>
