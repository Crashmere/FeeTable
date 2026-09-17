<script setup lang="ts">
import { ref, onMounted, nextTick } from "vue";
import { useRouter, useRoute } from "vue-router";
import {
  api,
  ApiError,
  type TableList,
  type Table,
  type TableSort,
} from "../api";
import { money } from "../money";
import Icon from "../components/Icon.vue";
import Modal from "../components/Modal.vue";
import MergeTables from "../components/MergeTables.vue";
const router = useRouter();
const route = useRoute();
const sorts: TableSort[] = [
  "updated_desc",
  "updated_asc",
  "month_desc",
  "month_asc",
];
const sort = ref<TableSort>(
  sorts.includes(route.query.sort as TableSort)
    ? (route.query.sort as TableSort)
    : "updated_desc",
);
const merging = ref<Table>();
const data = ref<TableList>();
const loading = ref(true);
const error = ref("");
const page = ref(1);
const creating = ref(false);
const deleting = ref<Table>();
const busy = ref(false);
const formError = ref("");
const uncertain = ref(false);
const now = new Date();
const year = ref(now.getFullYear());
const month = ref(now.getMonth() + 1);
let sequence = 0;
async function load() {
  const n = ++sequence;
  loading.value = !data.value;
  error.value = "";
  try {
    const result = await api.tables(page.value, sort.value);
    if (n !== sequence) return;
    data.value = result;
    const last = Math.max(1, Math.ceil(result.total / 30));
    if (page.value > last) {
      page.value = last;
      await load();
    }
  } catch (e) {
    if (n === sequence) error.value = (e as Error).message;
  } finally {
    if (n === sequence) loading.value = false;
  }
}
onMounted(load);
function openCreate() {
  formError.value = "";
  uncertain.value = false;
  creating.value = true;
}
function askDelete(t: Table) {
  formError.value = "";
  uncertain.value = false;
  deleting.value = t;
}
async function create() {
  if (busy.value || uncertain.value) return;
  busy.value = true;
  formError.value = "";
  try {
    const t = await api.createTable(Number(year.value), Number(month.value));
    busy.value = false;
    creating.value = false;
    await nextTick();
    await router.push("/tables/" + t.id);
  } catch (e) {
    formError.value = (e as Error).message;
    uncertain.value = e instanceof ApiError && e.uncertain;
  } finally {
    busy.value = false;
  }
}
async function remove() {
  if (!deleting.value || busy.value || uncertain.value) return;
  busy.value = true;
  try {
    await api.deleteTable(deleting.value);
    deleting.value = undefined;
    await load();
  } catch (e) {
    formError.value = (e as Error).message;
    uncertain.value = e instanceof ApiError && e.uncertain;
  } finally {
    busy.value = false;
  }
}
function turn(n: number) {
  page.value = n;
  load();
}
async function changeSort() {
  page.value = 1;
  await router.replace({ query: { ...route.query, sort: sort.value } });
  load();
}
function closeMerge() {
  merging.value = undefined;
  load();
}
async function merged(table: Table) {
  merging.value = undefined;
  await nextTick();
  await router.push("/tables/" + table.id);
}
</script>
<template>
  <section class="page-heading">
    <div>
      <h1>我的运费表</h1>
    </div>
    <button class="primary" @click="openCreate">
      <Icon name="plus" />新建运费表
    </button>
  </section>
  <div class="home-toolbar">
    <RouterLink to="/locations" class="button">地点管理</RouterLink>
    <label class="sort-control"
      >排序<select
        v-model="sort"
        aria-label="排序"
        :disabled="loading"
        @change="changeSort"
      >
        <option value="updated_desc">修改时间：最新在前</option>
        <option value="updated_asc">修改时间：最早在前</option>
        <option value="month_desc">表格年月：最新在前</option>
        <option value="month_asc">表格年月：最早在前</option>
      </select></label
    >
  </div>
  <div v-if="error" class="notice error" role="alert">
    {{ error }} <button @click="load">重试</button>
  </div>
  <div v-else-if="loading" class="card-grid" aria-label="正在加载">
    <div v-for="n in 3" :key="n" class="skeleton skeleton-card" />
  </div>
  <template v-else-if="data"
    ><div class="section-label">
      <span>全部表格</span><span>{{ data.total }} 张</span>
    </div>
    <div v-if="data.items.length" class="card-grid">
      <article v-for="t in data.items" :key="t.id" class="month-card">
        <RouterLink :to="'/tables/' + t.id" class="month-card-main"
          ><div class="card-top">
            <span class="month-badge">{{
              String(t.month).padStart(2, "0")
            }}</span
            ><span class="subtle">{{ t.year }} 年</span
            ><Icon class="card-arrow" name="arrow" />
          </div>
          <h2>{{ t.month }} 月运费明细</h2>
          <p class="amount-large"><small>¥</small>{{ money(t.total) }}</p>
          <div class="card-meta">
            <span>{{ t.recordCount }} 条记录</span>
            <span>#{{ t.id }}</span>
          </div></RouterLink
        >
        <div class="card-bottom">
          <span
            >更新于
            {{ new Date(t.updatedAt).toLocaleDateString("zh-CN") }}</span
          >
          <div class="row-actions">
            <button
              class="card-merge"
              :aria-label="'合并到表 #' + t.id"
              @click="merging = t"
            >
              合并</button
            ><button
              class="icon-button"
              :aria-label="'删除 ' + t.year + '年' + t.month + '月表格'"
              @click="askDelete(t)"
            >
              <Icon name="trash" :size="17" />
            </button>
          </div>
        </div>
      </article>
    </div>
    <div v-else class="empty-state">
      <span class="empty-icon"><Icon name="table" :size="42" /></span>
      <h2>暂无运费表</h2>
    </div>
    <div v-if="data.total > 30" class="pagination">
      <button :disabled="page === 1" @click="turn(page - 1)">上一页</button
      ><span>{{ page }} / {{ Math.ceil(data.total / 30) }}</span
      ><button :disabled="page * 30 >= data.total" @click="turn(page + 1)">
        下一页
      </button>
    </div></template
  >
  <Modal
    v-if="creating"
    title="新建运费表"
    :busy="busy"
    @close="creating = false"
    ><form @submit.prevent="create">
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
        <button type="button" @click="creating = false" :disabled="busy">
          取消</button
        ><button class="primary" :disabled="busy || uncertain">创建表格</button>
      </div>
    </form></Modal
  >
  <Modal
    v-if="deleting"
    title="删除运费表"
    :busy="busy"
    @close="deleting = undefined"
    ><p>
      确定删除 {{ deleting.year }} 年 {{ deleting.month }} 月的这张表格及全部
      {{ deleting.recordCount }} 条记录？此操作无法撤销。
    </p>
    <p v-if="formError" class="notice error">{{ formError }}</p>
    <div class="modal-actions">
      <button @click="deleting = undefined" :disabled="busy">取消</button
      ><button class="danger" @click="remove" :disabled="busy || uncertain">
        确认删除
      </button>
    </div></Modal
  >
  <MergeTables
    v-if="merging"
    :target="merging"
    @close="closeMerge"
    @merged="merged"
  />
</template>
