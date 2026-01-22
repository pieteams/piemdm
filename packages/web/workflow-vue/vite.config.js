// vite.config.js
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import dts from 'vite-plugin-dts';
import path from 'path';

export default defineConfig({
  // 插件配置 - 只保留必要的 Vue 插件
  plugins: [
    vue(),
    dts(),
  ],

  // 构建配置 - 专注于组件库构建
  build: {
    // 输出目录
    outDir: 'dist',

    // 库模式配置
    lib: {
      // 库模式的入口
      entry: path.resolve(__dirname, 'src/lib/index.js'),
      // 包名，在 UMD 构建模式下会作为全局变量名
      name: 'WorkflowVue',
      // 输出文件名的前缀
      fileName: format => `workflow-vue.${format}.js`,
      // 明确指定输出格式
      formats: ['es', 'umd'],
    },

    // 源映射 - 便于调试
    sourcemap: true,

    // 清空输出目录
    emptyOutDir: true,

    // Rollup 配置
    rollupOptions: {
      // 确保外部化处理那些你不想打包进库的依赖
      external: ['vue', 'bootstrap', 'bootstrap-icons', 'uuid', 'vue-select'],

      output: {
        // 在 UMD 构建模式下，为这些外部化的依赖提供一个全局变量
        globals: {
          vue: 'Vue',
          bootstrap: 'bootstrap',
          'bootstrap-icons': 'bootstrapIcons',
          uuid: 'uuidv4',
          'vue-select': 'VueSelect',
        },

        // 导出配置：明确指定使用命名导出
        exports: 'named',

        // 配置 CSS 输出文件名 - 优化逻辑
        assetFileNames: assetInfo => {
          // 如果是 CSS 文件，固定输出为 style.css
          if (assetInfo.names?.some(name => name?.endsWith('.css'))) {
            return 'style.css';
          }
          // 其他资源文件保持默认命名
          return 'assets/[name]-[hash][extname]';
        },
      },
    },

    // 构建报告
    reportCompressedSize: true,

    // chunk 大小警告限制
    chunkSizeWarningLimit: 1000,
  },

  // 解析配置 - 添加路径别名便于开发
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@lib': path.resolve(__dirname, 'src/lib'),
    },
    extensions: ['.js', '.vue', '.json'],
  },

  // CSS 配置 - 简化配置
  css: {
    // 开发环境启用 sourcemap
    devSourcemap: true,
  },

  // 开发服务器配置 - 简化，仅用于组件开发测试
  server: {
    port: 3000,
    open: false,
  },

  // 预览服务器配置
  preview: {
    port: 3001,
    open: false,
  },
});
