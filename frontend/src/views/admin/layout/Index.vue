<template>
  <nav class="navbar navbar-expand-lg border-bottom border-body fixed-top" style="background-color: #e3f2fd"
    data-bs-theme="light">
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
        <li class="nav-item">
          <a class="nav-link text-nowrap" href="/approval/index"> <i class="bi bi-house-door"></i>
            {{ $t('Workbench') }}
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
            </ul>
          </div>
        </li>
        <div class="vr bg-black"></div>
        <li class="nav-item dropdown">
          <a class="nav-link dropdown-toggle" href="#" role="button" data-bs-toggle="dropdown" aria-expanded="false">
            <i class="bi bi-person"></i>
            {{ userStore.username }}
          </a>
          <ul class="dropdown-menu dropdown-menu-end">
            <li>
              <a class="dropdown-item" href="javascript:;" @click="onLogout">
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
      <div class="sidebar border-end col-md-3 col-lg-2 p-0 bg-body-tertiary">
        <div class="offcanvas-md offcanvas-end bg-body-tertiary" tabindex="-1" id="sidebarMenu"
          aria-labelledby="sidebarMenuLabel">
          <div class="offcanvas-body d-md-flex flex-column p-0 pt-lg-3 overflow-y-auto">
            <ul class="nav flex-column">
              <li class="nav-item">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/dashboard/index"
                  :class="{ active: route.path.startsWith('/admin/dashboard') }">
                  <i class="bi bi-house-fill"></i>
                  {{ $t('Dashboard') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('table:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/table/index"
                  :class="{ active: route.path.startsWith('/admin/table') }">
                  <i class="bi bi-table"></i>
                  {{ $t('Table') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('approval_def:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/approval_def/index"
                  :class="{ active: route.path.startsWith('/admin/approval_def') }">
                  <i class="bi bi-diagram-3"></i>
                  {{ $t('Approval Define') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('approval:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/approval/index"
                  :class="{ active: route.path.startsWith('/admin/approval') && !route.path.startsWith('/admin/approval_def') }">
                  <i class="bi bi-clipboard-check"></i>
                  {{ $t('Approvel') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('application:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/application/index"
                  :class="{ active: route.path.startsWith('/admin/application') }">
                  <i class="bi bi-app-indicator"></i>
                  {{ $t('Application') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('webhook:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/webhook/index"
                  :class="{ active: route.path.startsWith('/admin/webhook') && !route.path.startsWith('/admin/webhook_delivery') }">
                  <i class="bi bi-link-45deg"></i>
                  {{ $t('Webhook') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('webhook_delivery:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/webhook_delivery/index"
                  :class="{ active: route.path.startsWith('/admin/webhook_delivery') }">
                  <i class="bi bi-mailbox"></i>
                  {{ $t('Webhook Delivery') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('notification_template:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/notification_template/index"
                  :class="{ active: route.path.startsWith('/admin/notification_template') }">
                  <i class="bi bi-bell"></i>
                  {{ $t('Notification Templates') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('notification_log:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/notification_log/index"
                  :class="{ active: route.path.startsWith('/admin/notification_log') }">
                  <i class="bi bi-file-earmark-text"></i>
                  {{ $t('Notification Logs') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('cron:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/cron/index"
                  :class="{ active: route.path.startsWith('/admin/cron') && !route.path.startsWith('/admin/cron_log') }">
                  <i class="bi bi-clock-history"></i>
                  {{ $t('Cron') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('cron_log:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/cron_log/index"
                  :class="{ active: route.path.startsWith('/admin/cron_log') }">
                  <i class="bi bi-journal-text"></i>
                  {{ $t('Cron Log') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('permission:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/permission/index"
                  :class="{ active: route.path.startsWith('/admin/permission') }">
                  <i class="bi bi-shield-check"></i>
                  {{ $t('Permissions') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('role:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/role/index"
                  :class="{ active: route.path.startsWith('/admin/role') }">
                  <i class="bi bi-person-badge"></i>
                  {{ $t('Roles') }}
                </a>
              </li>
              <li class="nav-item" v-if="hasPermission('user:list')">
                <a class="nav-link d-flex align-items-center gap-2" href="/admin/user/index"
                  :class="{ active: route.path.startsWith('/admin/user') }">
                  <i class="bi bi-people"></i>
                  {{ $t('Users') }}
                </a>
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
</template>

<script setup>
import { getLanguageLabel } from '@/utils/language';
import { localStorage } from '@/utils/local-storage';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { useUserStore } from '@/pinia/modules/user';

const userStore = useUserStore();
const hasPermission = (p) => userStore.hasPermission(p);

const route = useRoute();
const { locale } = useI18n();
const active = ref(locale.value || localStorage.get('lang'));
const getLanguageLabelFn = code => getLanguageLabel(code);
const changeActive = lang => {
  locale.value = lang;
  active.value = lang;
  localStorage.set('lang', lang);
};

const onLogout = () => {
  userStore.logout();
};
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

.sidebar .nav-link {
  font-size: 0.875rem;
}

.sidebar .nav-link.active {
  background-color: rgb(227, 242, 253);
}

.sidebar .nav-link:hover {
  background-color: #e0e0e0;
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

.bi-trash:hover::before {
  content: '\F5DD';
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
</style>
