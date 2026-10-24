import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite-plus";

export default defineConfig({
  plugins: [tailwindcss()],
  server: {
    strictPort: true,
    proxy: {
      "/api/": "http://localhost:8080",
      "/dex": "http://dex:5556",
    },
  },
});
