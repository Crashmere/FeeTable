<script setup lang="ts">
import Icon from "./components/Icon.vue";
</script>
<template>
  <header class="site-header">
    <RouterLink to="/" class="brand"
      ><span class="brand-icon"><Icon name="table" :size="23" /></span
      ><span>运费明细表<small>FEE TABLE</small></span></RouterLink
    ><span class="header-note">每一程，清晰有据</span>
  </header>
  <main>
    <RouterView v-slot="{ Component, route }"
      ><component :is="Component" :key="route.path"
    /></RouterView>
  </main>
  <footer class="site-footer">
    <span>运费明细表</span><span>记录 · 汇总 · 导出</span>
  </footer>
</template>
