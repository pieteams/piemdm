import js from '@eslint/js';
import vuePlugin from 'eslint-plugin-vue';
import globals from 'globals';

export default [
  // 基础 JavaScript 配置
  js.configs.recommended,

  // Vue 配置
  ...vuePlugin.configs['flat/recommended'],

  // 全局变量配置
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
        ...globals.es2021,
      },
      ecmaVersion: 'latest',
      sourceType: 'module',
    },
  },

  // 自定义规则 - 移除所有自定义规则，使用默认推荐配置
  {
    rules: {
      'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      'no-debugger': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
      'no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    },
  },

  // 测试文件配置
  {
    files: ['**/__tests__/**/*.[jt]s?(x)', '**/?(*.)+(spec|test).[jt]s?(x)'],
    languageOptions: {
      globals: {
        ...globals.jest,
        vitest: true,
        describe: true,
        it: true,
        test: true,
        expect: true,
        beforeEach: true,
        afterEach: true,
      },
    },
  },

  // 忽略文件
  {
    ignores: [
      'node_modules/',
      'dist/',
      'coverage/',
      '.zed/',
      '.vscode/',
      '*.log',
      '*.min.js',
      '*.min.css',
      '*.umd.js',
      '*.es.js',
      '*.cjs.js',
      'package.json.backup',
      '__snapshots__/',
      '.env',
      '.env.local',
      '.env.*.local',
      '.idea/',
      '*.swp',
      '*.swo',
      '*~',
      '.DS_Store',
      'Thumbs.db',
      'pnpm-lock.yaml',
      'package-lock.json',
      'yarn.lock',
      'docs/.vitepress/dist/',
    ],
  },
];
