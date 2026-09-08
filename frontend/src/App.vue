<template>
  <a-config-provider :theme="themeConfig">
    <router-view />
  </a-config-provider>
</template>

<script setup>
const themeConfig = {
  token: {
    colorPrimary: '#2D6A4F',
    colorSuccess: '#52B788',
    colorError: '#C1121F',
    colorWarning: '#E09F3E',
    colorBgLayout: '#F8F6F3',
    colorBgContainer: '#FFFFFF',
    colorBorder: '#E8E4DD',
    colorText: '#1A1A1A',
    colorTextSecondary: '#6B6B6B',
    borderRadius: 8,
  },
}
</script>
