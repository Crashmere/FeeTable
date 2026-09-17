<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { api, ApiError, type Suggestion } from "../api";
import Modal from "../components/Modal.vue";
import Icon from "../components/Icon.vue";
const items = ref<Suggestion[]>([]);
const search = ref("");
const loading = ref(true);
const error = ref("");
const editing = ref<Suggestion>();
const deleting = ref<Suggestion>();
const form = ref(false);
const name = ref("");
const busy = ref(false);
const formError = ref("");
const uncertain = ref(false);
const filtered = computed(() =>
  items.value.filter((item) =>
    item.name
      .toLocaleLowerCase()
      .includes(search.value.trim().toLocaleLowerCase()),
  ),
);
async function load() {
  loading.value = true;
  error.value = "";
  try {
    items.value = await api.suggestions("locations");
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
onMounted(load);
function edit(item?: Suggestion) {
  editing.value = item;
  name.value = item?.name || "";
  formError.value = "";
  uncertain.value = false;
  form.value = true;
}
function askDelete(item: Suggestion) {
  formError.value = "";
  uncertain.value = false;
  deleting.value = item;
}
function failed(e: unknown) {
  formError.value = (e as Error).message;
  uncertain.value =
    e instanceof ApiError &&
    (e.uncertain || e.code === "CONFLICT" || e.code === "NOT_FOUND");
}
async function save() {
  if (busy.value || uncertain.value) return;
  busy.value = true;
  formError.value = "";
  try {
    if (editing.value) await api.updateLocation(editing.value, name.value);
    else await api.addSuggestion("locations", name.value);
    form.value = false;
    await load();
  } catch (e) {
    failed(e);
  } finally {
    busy.value = false;
  }
}
async function remove() {
  if (!deleting.value || busy.value || uncertain.value) return;
  busy.value = true;
  formError.value = "";
  try {
    await api.deleteLocation(deleting.value);
    deleting.value = undefined;
    await load();
  } catch (e) {
    failed(e);
  } finally {
    busy.value = false;
  }
}
function close() {
  form.value = false;
  deleting.value = undefined;
  if (uncertain.value) load();
}
</script>
<template>
  <RouterLink to="/" class="back-link"
    ><Icon name="back" :size="17" />返回运费表</RouterLink
  >
  <section class="page-heading">
    <h1>地点管理</h1>
    <button class="primary" @click="edit()">
      <Icon name="plus" />新增地点
    </button>
  </section>
  <label class="location-search"
    >搜索地点<input v-model="search" type="search" placeholder="输入地点名称"
  /></label>
  <div v-if="error" class="notice error" role="alert">
    {{ error }} <button @click="load">重试</button>
  </div>
  <div v-else-if="loading" class="skeleton skeleton-table" />
  <template v-else>
    <div class="section-label">
      <span>常用地点</span><span>{{ filtered.length }} 个</span>
    </div>
    <ul v-if="filtered.length" class="location-list">
      <li v-for="item in filtered" :key="item.id">
        <span class="location-name">{{ item.name }}</span>
        <div class="row-actions">
          <button
            class="icon-button"
            :aria-label="'改名 ' + item.name"
            @click="edit(item)"
          >
            <Icon name="edit" />
          </button>
          <button
            class="icon-button"
            :aria-label="'删除 ' + item.name"
            @click="askDelete(item)"
          >
            <Icon name="trash" />
          </button>
        </div>
      </li>
    </ul>
    <div v-else class="empty-state">
      <h2>{{ search ? "未找到地点" : "暂无地点" }}</h2>
    </div>
  </template>
  <Modal
    v-if="form"
    :title="editing ? '地点改名' : '新增地点'"
    :busy="busy"
    @close="close"
  >
    <form @submit.prevent="save">
      <label
        >地点名称<input
          v-model="name"
          required
          maxlength="80"
          autocomplete="off"
      /></label>
      <p v-if="editing" class="form-note">
        仅修改常用地点，已有运输记录保持原名称。
      </p>
      <p v-if="formError" class="notice error" role="alert">{{ formError }}</p>
      <div class="modal-actions">
        <button type="button" :disabled="busy" @click="close">
          {{ uncertain ? "返回核对" : "取消" }}</button
        ><button class="primary" :disabled="busy || uncertain">保存地点</button>
      </div>
    </form>
  </Modal>
  <Modal v-if="deleting" title="删除地点" :busy="busy" @close="close">
    <p class="wrap-text">
      确定从常用地点中删除「{{ deleting.name }}」？已有运输记录不受影响。
    </p>
    <p v-if="formError" class="notice error" role="alert">{{ formError }}</p>
    <div class="modal-actions">
      <button :disabled="busy" @click="close">
        {{ uncertain ? "返回核对" : "取消" }}</button
      ><button class="danger" :disabled="busy || uncertain" @click="remove">
        确认删除
      </button>
    </div>
  </Modal>
</template>
