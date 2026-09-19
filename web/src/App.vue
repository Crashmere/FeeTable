<script setup lang="ts">
import brandIcon from "./assets/icon.svg?no-inline";
import SiteMenu from "./components/SiteMenu.vue";
</script>
<template>
  <header class="site-header">
    <RouterLink to="/" class="brand">
      <img class="brand-icon" :src="brandIcon" width="43" height="43" alt="" />
      <span>运费明细表</span>
    </RouterLink>
    <SiteMenu />
  </header>
  <main>
    <RouterView v-slot="{ Component, route }"
      ><component :is="Component" :key="route.path"
    /></RouterView>
  </main>
</template>
