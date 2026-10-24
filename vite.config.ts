import { defineConfig } from "vite-plus";

export default defineConfig({
  fmt: {
    ignorePatterns: ["packages/proto/proto/**", "packages/proto/google/**"],
  },
});
