<script setup lang="ts">
import Icon from "./components/Icon.vue";
import SiteMenu from "./components/SiteMenu.vue";
</script>
<template>
  <header class="site-header">
    <RouterLink to="/" class="brand"
      ><span class="brand-icon"><Icon name="table" :size="23" /></span
      ><span>运费明细表</span></RouterLink
    >
    <SiteMenu />
  </header>
  <main>
    <RouterView v-slot="{ Component, route }"
      ><component :is="Component" :key="route.path"
    /></RouterView>
  </main>
</template>
