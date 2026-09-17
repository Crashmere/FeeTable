import { createApp } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import App from "./App.vue";
import Home from "./pages/Home.vue";
import TablePage from "./pages/TablePage.vue";
import ExportPage from "./pages/ExportPage.vue";
import "./style.css";
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: "/", component: Home },
    { path: "/tables/:id", component: TablePage },
    { path: "/tables/:id/export", component: ExportPage },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
  scrollBehavior: () => ({ top: 0 }),
});
createApp(App).use(router).mount("#app");
