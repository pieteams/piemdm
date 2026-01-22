import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue()],
  test: {
    // 测试环境
    environment: "node",

    // 测试文件匹配模式 - 只包含 .test.js 文件
    include: ["tests/**/*.test.js"],
    exclude: ["node_modules", "dist", ".git", "**/*.d.ts"],

    // 全局设置
    globals: true,

    // 测试超时时间（毫秒）
    testTimeout: 10000,

    // 钩子超时时间（毫秒）
    hookTimeout: 10000,

    // 是否在失败时继续运行其他测试
    allowOnly: false,

    // 是否运行被标记为 `todo` 的测试
    includeTaskLocation: true,

    // 报告器
    reporter: ["verbose", "default"],

    // 输出目录
    outputFile: {
      html: "test-results/html/index.html",
      json: "test-results/json/index.json",
      junit: "test-results/junit/junit.xml",
    },

    // 覆盖率配置
    coverage: {
      enabled: false, // 暂时关闭覆盖率，专注于功能测试
      provider: "v8",
      reporter: ["text"],
      reportsDirectory: "test-results/coverage",
      exclude: ["node_modules/", "test/", "dist/", "**/*.config.js", "**/*.config.ts"],
    },
  },
});
