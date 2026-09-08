<template>
  <a-layout class="min-h-100vh">
    <a-layout-sider
      :trigger="null"
      collapsible
      class="admin-sider"
    >
      <div class="admin-sider-title">
        Admin Panel
      </div>
      <a-menu
        theme="dark"
        class="admin-menu"
      >
        <a-menu-item v-for="item in menuItems" :key="item.path">
          <router-link :to="item.path">
            <component :is="item.icon" />
            <span>{{ item.label }}</span>
          </router-link>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <a-layout>
      <a-layout-content class="admin-content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup>
import {
  DashboardOutlined,
  ShoppingOutlined,
  FileTextOutlined,
  TeamOutlined,
  AppstoreOutlined,
  WarningOutlined,
} from '@ant-design/icons-vue'

const menuItems = [
  { path: '/admin', label: 'Dashboard', icon: DashboardOutlined },
  { path: '/admin/products', label: 'Products', icon: ShoppingOutlined },
  { path: '/admin/orders', label: 'Orders', icon: FileTextOutlined },
  { path: '/admin/users', label: 'Users', icon: TeamOutlined },
  { path: '/admin/categories', label: 'Categories', icon: AppstoreOutlined },
  { path: '/admin/dead-letters', label: 'Dead Letters', icon: WarningOutlined },
]
</script>
