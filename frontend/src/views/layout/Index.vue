<template>
  <div id="app-layout">
    <nav class="navbar navbar-expand-md navbar-dark fixed-top bg-dark">
      <!-- navbar sticky-top bg-dark flex-md-nowrap p-0 shadow -->
      <div class="container-fluid">
        <div class="d-flex align-items-center">
          <a class="navbar-brand" href="#">
            <img src="@/assets/piemdm.png" alt="Logo" width="24" height="24" class="d-inline-block align-text-top" />
            PieMDM
          </a>
          <div class="bd-navbar-toggle d-md-none ms-2">
            <button class="navbar-toggler p-2" type="button" data-bs-toggle="offcanvas" data-bs-target="#sidebarMenu"
              aria-controls="sidebarMenu" aria-label="Toggle sidebar navigation">
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" class="bi" fill="currentColor"
                viewBox="0 0 16 16">
                <path fill-rule="evenodd"
                  d="M2.5 11.5A.5.5 0 0 1 3 11h10a.5.5 0 0 1 0 1H3a.5.5 0 0 1-.5-.5zm0-4A.5.5 0 0 1 3 7h10a.5.5 0 0 1 0 1H3a.5.5 0 0 1-.5-.5zm0-4A.5.5 0 0 1 3 3h10a.5.5 0 0 1 0 1H3a.5.5 0 0 1-.5-.5z">
                </path>
              </svg>
              <span class="d-none fs-6 pe-1">Browse</span>
            </button>
          </div>
        </div>
        <button class="navbar-toggler ms-auto" type="button" data-bs-toggle="collapse"
          data-bs-target="#navbarNavDropdown" aria-controls="navbarNavDropdown" aria-expanded="false"
          aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
      </div>
      <div class="collapse navbar-collapse" id="navbarNavDropdown">
        <ul class="navbar-nav gap-2">
          <li class="nav-item">
            <a class="nav-link" href="/web/document/index" target="_blank">
              <i class="bi bi-book"></i>
            </a>
          </li>
          <li class="nav-item" v-if="userStore.canAccessAdmin">
            <a class="nav-link text-nowrap" href="/admin/dashboard/index"> <i class="bi bi-gear"></i>
              {{ $t('System Management') }}
            </a>
          </li>
          <li class="nav-item dropdown">
            <div class="dropdown">
              <button class="btn btn-link nav-link py-2 px-0 px-lg-2 dropdown-toggle d-flex align-items-center"
                type="button" data-bs-toggle="dropdown" aria-expanded="false">
                <i class="bi bi-translate"></i>
                &nbsp;
                {{ getLanguageLabelFn(active) }}
              </button>
              <ul class="dropdown-menu dropdown-menu-end">
                <li>
                  <a class="dropdown-item" href="#" @click="changeActive('en-US')">
                    English
                  </a>
                </li>
                <li>
                  <a class="dropdown-item" href="#" @click="changeActive('zh-CN')">
                    简体中文
                  </a>
                </li>
                <li>
                  <a class="dropdown-item" href="#" @click="changeActive('zh-TW')">
                    繁體中文
                  </a>
                </li>
              </ul>
            </div>
          </li>
          <div class="vr bg-white"></div>
          <li class="nav-item dropdown">
            <a class="nav-link dropdown-toggle" href="#" role="button" data-bs-toggle="dropdown" aria-expanded="false">
              <i class="bi bi-person"></i>
              {{ userStore.userInfo.username }}
            </a>
            <ul class="dropdown-menu dropdown-menu-end">
              <li>
                <a class="dropdown-item" href="#" @click.prevent="handleLogout">
                  {{ $t('Logout') }}
                </a>
              </li>
            </ul>
          </li>
        </ul>
      </div>
    </nav>

    <div class="container-fluid pt-5">
      <div class="row">
        <!-- Sidebar -->
        <div class="sidebar border-end col-md-3 col-lg-2 p-1 pt-1 bg-body-tertiary" style="background-color: #80bfff">
          <div class="offcanvas-md offcanvas-end bg-body-tertiary" tabindex="-1" id="sidebarMenu"
            aria-labelledby="sidebarMenuLabel" style="background-color: #80bfff">
            <div class="offcanvas-body d-md-flex flex-column pt-lg-3 overflow-y-auto">
              <ul class="nav flex-column">
                <li class="nav-item">
                  <a class="nav-link d-flex align-items-center gap-2" :class="{ active: activePage === 'approval' }"
                    :href="`/approval/index`">
                    <i class="bi bi-clipboard-check"></i>
                    {{ $t('Approval') }}
                  </a>
                </li>
              </ul>

              <h6
                class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-3 text-body-secondary text-uppercase">
                <span>{{ $t('Entities') }}</span>
                <a class="link-secondary" href="#" aria-label="Add a new report"></a>
              </h6>
              <ul class="nav flex-column mb-auto">
                <!-- 动态表格列表 -->
                <li class="nav-item" v-for="table in tableList" :key="table.id">
                  <a class="nav-link d-flex align-items-center gap-2" :class="{ active: activePage === table.Code }"
                    :href="`/entity/index?table_code=${table.Code}`" @click="activePage = table.Code">
                    <i class="bi bi-file-earmark-text"></i>
                    {{ table.Name || table.Code }}
                  </a>
                </li>
                <!-- 暂无数据提示 -->
                <li class="nav-item" v-if="!loading && tableList.length === 0">
                  <div class="nav-link d-flex align-items-center gap-2 text-muted fst-italic">
                    <i class="bi bi-info-circle"></i>
                    暂无表格数据
                  </div>
                </li>
                <!-- 加载中提示 -->
                <li class="nav-item" v-if="loading">
                  <div class="nav-link d-flex align-items-center gap-2">
                    <div class="spinner-border spinner-border-sm text-primary" role="status">
                      <span class="visually-hidden">加载中...</span>
                    </div>
                    正在加载...
                  </div>
                </li>
              </ul>
            </div>
          </div>
        </div>
        <!-- Main -->
        <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
          <router-view></router-view>
        </main>
      </div>
    </div>
    <div>
      <footer class="d-flex flex-wrap justify-content-between align-items-center p-3 border-top"
        style="font-size: 0.8rem">
        <div class="col-md-4 d-flex align-items-center">
          <span class="mb-3 mb-md-0 text-muted">© 2022 Company, Inc</span>
        </div>

        <ul class="nav col-md-4 justify-content-end list-unstyled d-flex pr-5">
          <li class="ms-3">
            <a class="text-muted" href="javascript:;">
              <i class="bi bi-twitter"></i>
            </a>
          </li>
          <li class="ms-3">
            <a class="text-muted" href="javascript:;">
              <i class="bi bi-instagram"></i>
            </a>
          </li>
          <li class="ms-3">
            <a class="text-muted" href="javascript:;">
              <i class="bi bi-facebook"></i>
            </a>
          </li>
        </ul>
      </footer>
    </div>
  </div>
</template>

<script setup>
import { getTableList } from '@/api/table';
import { useUserStore } from '@/pinia/modules/user';
import { getLanguageLabel } from '@/utils/language';
import { localStorage } from '@/utils/local-storage';
import httpLinkHeader from 'http-link-header';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

// 响应式数据
const { locale } = useI18n();
const active = ref(locale.value || localStorage.get('lang'));
const route = useRoute();
const tableList = ref([]);
const activePage = ref('dashboard');
const loading = ref(false);
const page = ref(1);
const pageSize = ref(150); // 设置较大的pageSize以获取更多表格
const total = ref(0);
const userStore = useUserStore();

// 组件挂载时获取数据
onMounted(() => {
  activePage.value = route.query.table_code || '';
  // 获取当前路由的path，如果包含 approval 则activePage为 approval
  if (route.path.includes('approval')) {
    activePage.value = 'approval';
  } else if (route.path.includes('dashboard')) {
    activePage.value = 'dashboard';
  } else {
    activePage.value = route.query.table_code || '';
  }
  getEntityList();
});

const changeActive = lang => {
  locale.value = lang;
  active.value = lang;
  localStorage.set('lang', lang);
};

// 获取表格列表数据
const getEntityList = async () => {
  try {
    loading.value = true;
    // 参考admin/table/index.vue的实现，使用getTableList API
    const res = await getTableList({
      page: page.value,
      pageSize: pageSize.value,
      table_type: 'Entity',
      // 不指定搜索条件，获取所有表格
    });

    if (res) {
      tableList.value = res.data;
      const links = httpLinkHeader.parse(res.headers.link).refs;
      links.forEach(link => {
        if (['last'].includes(link.rel)) {
          const url = new URL(link.uri);
          total.value = parseInt(url.searchParams.get('page')) || 1;
        }
      });
    } else {
      tableList.value = [];
    }
  } catch (error) {
    console.error('获取表格列表失败:', error);
    tableList.value = []; // 清空数据
  } finally {
    loading.value = false;
  }
};

// 处理登出
const handleLogout = () => {
  userStore.logout();
};

// 导出 getLanguageLabel 以供模板使用
const getLanguageLabelFn = code => getLanguageLabel(code);
</script>

<style>
.bd-placeholder-img {
  font-size: 1.125rem;
  text-anchor: middle;
  -webkit-user-select: none;
  -moz-user-select: none;
  user-select: none;
}

@media (min-width: 768px) {
  .bd-placeholder-img-lg {
    font-size: 3.5rem;
  }
}

.btn-toolbar .bi {
  display: inline-block;
  width: 1rem;
  height: 1rem;
}

.sidebar-heading {
  font-size: 0.75rem;
}

.nav {
  --bs-nav-link-color: black;
  --bs-nav-link-hover-color: var(--bs-link-hover-color);
  --bs-nav-link-disabled-color: black;
}

.actions {
  /* 固定位置，始终留在可视区域左侧 */
  position: sticky;
  right: -5px;
  z-index: 1000;
  padding-left: 10px !important;
  padding-right: 10px !important;
}

.actions a {
  /* text-decoration: none; */
  margin-right: 5px;
}

.actions a:last-child {
  margin-right: 0;
}

.bi-clipboard2-check:hover::before {
  content: '\f724';
}

.bi-grid-3x2-gap:hover::before {
  content: '\F3F5';
}

.bi-file-text:hover::before {
  content: '\F3B8';
}

.bi-pencil:hover::before {
  content: '\F4C9';
}

.bi-bell:hover::before {
  content: '\F189';
}

.bi-arrow-left-circle:hover::before {
  content: '\F129';
}

.bi-file-ruled:hover::before {
  content: '\F3B2';
}

.bi-send:hover::before {
  content: '\F6B9';
}

.bi-clock:hover::before {
  content: '\F291';
}

.sidebar .nav-link {
  font-size: 0.875rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  border-radius: 0.3rem;
  margin-right: 5px;
}

.sidebar .nav-link.active {
  color: #ffffff;
  background-color: #0d6efd;
}

.sidebar .nav-link:hover {
  background-color: #d0d0d0;
}

.sidebar-heading {
  font-size: 0.75rem;
}

/* 侧边栏滚动样式 */
.sidebar .offcanvas-body {
  max-height: calc(100vh - 100px);
  overflow-y: auto;
}

/* 轻量级滚动条样式 */
.sidebar .offcanvas-body::-webkit-scrollbar {
  width: 6px;
}

.sidebar .offcanvas-body::-webkit-scrollbar-track {
  background: #f8f9fa;
}

.sidebar .offcanvas-body::-webkit-scrollbar-thumb {
  background-color: #dee2e6;
  border-radius: 3px;
}

.sidebar .offcanvas-body::-webkit-scrollbar-thumb:hover {
  background-color: #adb5bd;
}

.nav .nav-link {
  line-height: 1;
}

.dp__input {
  padding: 0.15rem 0.25rem 0.15rem 1.8rem !important;
  font-size: 0.875rem !important;
}

.vs__selected {
  font-size: 0.875rem !important;
  margin-top: 1px !important;
}

.vs__dropdown-toggle {
  padding-bottom: 1px !important;
}

.vs__clear {
  margin-top: -5px !important;
}

.vs__search {
  font-size: 0.875rem !important;
}

/* 确保侧边栏切换按钮在小屏幕上显示 */
.bd-navbar-toggle {
  display: block;
}

.bd-navbar-toggle .navbar-toggler {
  border: none;
  color: rgba(255, 255, 255, 0.85);
}

.bd-navbar-toggle .navbar-toggler:hover {
  color: rgba(255, 255, 255, 1);
}

.bd-navbar-toggle .navbar-toggler:focus {
  box-shadow: none;
  outline: none;
}

@media (min-width: 768px) {
  .bd-navbar-toggle {
    display: none;
  }
}
</style>
