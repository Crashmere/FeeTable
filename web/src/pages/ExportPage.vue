<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { api, download, type Report } from "../api";
import { money } from "../money";
import Icon from "../components/Icon.vue";
const route = useRoute();
const id = String(route.params.id);
const filter = ref(
  typeof route.query.tag === "string" ? route.query.tag : null,
);
const data = ref<Report>();
const loading = ref(true);
const error = ref("");
const exportError = ref("");
const busy = ref("");
const done = ref("");
const shareAvailable = !!navigator.share && window.isSecureContext;
async function load() {
  loading.value = true;
  error.value = "";
  try {
    data.value = await api.report(id, filter.value, 1, true);
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
onMounted(load);
async function run(format: string, share = false) {
  if (!data.value || busy.value) return;
  busy.value = format;
  exportError.value = "";
  done.value = "";
  try {
    await download(data.value, format, share);
    done.value = share ? "图片已准备好" : "文件已生成，请查看浏览器下载";
  } catch (e) {
    if ((e as Error).name !== "AbortError")
      exportError.value = (e as Error).message || "导出失败，请检查网络后重试";
  } finally {
    busy.value = "";
  }
}
function change(e: Event) {
  filter.value = (e.target as HTMLSelectElement).value || null;
  load();
}
</script>
<template>
  <RouterLink :to="'/tables/' + id" class="back-link"
    ><Icon name="back" :size="17" />返回运费明细</RouterLink
  >
  <section class="page-heading">
    <div>
      <h1>导出预览</h1>
    </div>
    <label class="filter-select"
      ><Icon name="tag" :size="17" /><select
        :value="filter || ''"
        @change="change"
        aria-label="导出标签范围"
        :disabled="loading || !!busy"
      >
        <option value="">整张运费表</option>
        <option v-for="t in data?.tags || []" :key="t">{{ t }}</option>
      </select></label
    >
  </section>
  <div v-if="error" class="notice error" role="alert">
    {{ error }} <button @click="load">重试</button>
  </div>
  <div v-else-if="loading" class="skeleton skeleton-table" />
  <template v-else-if="data"
    ><div class="export-layout">
      <section class="paper-surround">
        <div class="preview-caption">
          <span>{{ data.table.year }} 年 {{ data.table.month }} 月</span
          ><span>{{ data.count }} 条</span>
        </div>
        <div class="paper-scroll">
          <table class="report-table">
            <colgroup>
              <col style="width: 6%" />
              <col style="width: 6%" />
              <col style="width: 21%" />
              <col style="width: 21%" />
              <col style="width: 15%" />
              <col style="width: 12%" />
              <col style="width: 19%" />
            </colgroup>
            <thead>
              <tr>
                <th colspan="7" class="report-title">运费明细表</th>
              </tr>
              <tr>
                <th colspan="2">{{ data.table.year }}年</th>
                <th colspan="2">摘要</th>
                <th rowspan="2">运输数量</th>
                <th rowspan="2">单价</th>
                <th rowspan="2">运输金额</th>
              </tr>
              <tr>
                <th>月</th>
                <th>日</th>
                <th></th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in data.records" :key="r.id">
                <td>{{ data.table.month }}</td>
                <td>{{ r.day }}</td>
                <td class="text-left">{{ r.location1 }}</td>
                <td class="text-left">{{ r.location2 }}</td>
                <td class="numeric">{{ r.quantity }}</td>
                <td>{{ r.unitPrice ?? "" }}</td>
                <td class="numeric">{{ r.amount }}</td>
              </tr>
              <tr>
                <td colspan="6"></td>
                <td class="numeric">{{ data.total }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
      <aside class="export-options">
        <div class="export-total">
          <small>合计金额</small><strong>¥ {{ money(data.total) }}</strong>
        </div>
        <button
          class="primary"
          :disabled="!!busy || !data.count"
          @click="run('png')"
        >
          <Icon name="download" />下载 PNG 图片</button
        ><button :disabled="!!busy || !data.count" @click="run('pdf')">
          <Icon name="download" />下载 PDF 文档</button
        ><button :disabled="!!busy || !data.count" @click="run('xlsx')">
          <Icon name="download" />下载 Excel 表格</button
        ><button
          v-if="shareAvailable"
          :disabled="!!busy || !data.count"
          @click="run('png', true)"
        >
          分享 PNG 图片
        </button>
        <p v-if="busy" class="notice" role="status">
          <span class="spinner" />正在生成 {{ busy.toUpperCase() }}…
        </p>
        <p v-if="done" class="notice success" role="status">{{ done }}</p>
        <p v-if="exportError" class="notice error" role="alert">
          {{ exportError }} <button @click="load">刷新预览</button>
        </p>
      </aside>
    </div></template
  >
</template>
