import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";
export default defineConfig(({ mode }) => {
  const base = loadEnv(mode, ".", "").VITE_BASE_PATH || "/";
  if (!new RegExp("^/(?:[a-zA-Z0-9_-]+/)*$").test(base))
    throw new Error("Invalid VITE_BASE_PATH");
  return {
    base,
    plugins: [vue()],
    server: {
      proxy: {
        [base + "api"]: {
          target: "http://127.0.0.1:8081",
          rewrite: (p: string) => "/" + p.slice(base.length),
        },
      },
    },
  };
});
