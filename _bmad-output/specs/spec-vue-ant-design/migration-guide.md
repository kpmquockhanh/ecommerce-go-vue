# Migration Guide: Tailwind CSS → Ant Design Vue v4

## 1. Install Dependencies

```bash
cd frontend
npm install ant-design-vue@4
npm install -D unplugin-vue-components
```

No Less, PostCSS plugins, or babel-plugin-import needed — v4 uses CSS-in-JS (emotion).

## 2. Configure Vite (`vite.config.js`)

```javascript
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'
import Components from 'unplugin-vue-components/vite';
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers';

export default defineConfig({
  plugins: [
    vue(),
    Components({
      resolvers: [
        AntDesignVueResolver({
          importStyle: false, // CSS-in-JS mode — no manual style imports
        }),
      ],
    }),
  ],
  server: {
    port: 3000,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: process.env.DOCKER ? 'http://backend:4242' : 'http://localhost:4242',
        changeOrigin: true,
      },
    },
  },
})
```

## 3. Update Entry Point (`main.js`)

```javascript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue';
import router from './router'
import App from './App.vue'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(Antd)
app.mount('#app')
```

No CSS import needed — v4 injects styles via CSS-in-JS at runtime.

## 4. Theme Configuration (v4 Token API)

Wrap your app in `<a-config-provider>` with theme tokens:

```vue
<script setup>
const themeConfig = {
  token: {
    colorPrimary: '#2D6A4F',
    colorLink: '#2D6A4F',
    colorSuccess: '#52B788',
    borderRadius: 8,
    fontFamily: '-apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
  },
  // Optional: use dark algorithm
  // algorithm: theme.darkAlgorithm,
};
</script>

<template>
  <a-config-provider :theme="themeConfig">
    <router-view />
  </a-config-provider>
</template>
```

### Available Token Keys

| Token                    | Description              | Example Value |
| ------------------------ | ------------------------ | ------------- |
| `colorPrimary`           | Primary brand color      | `'#2D6A4F'`   |
| `colorLink`              | Link color               | `'#2D6A4F'`   |
| `colorSuccess`           | Success state color      | `'#52B788'`   |
| `colorWarning`           | Warning state color      | `'#FAAD14'`   |
| `colorError`             | Error state color        | `'#FF4D4F'`   |
| `colorBgBase`            | Base background color    | `'#FFFFFF'`   |
| `colorTextBase`          | Base text color          | `'#1A1A1A'`   |
| `borderRadius`           | Global border radius     | `8`           |
| `fontSize`               | Base font size           | `14`          |
| `fontFamily`             | Font family              | Custom string |

### Algorithm Options

```javascript
import { theme } from 'ant-design-vue';

const themeConfig = {
  token: { colorPrimary: '#2D6A4F' },
  algorithm: theme.darkAlgorithm,        // dark mode
  // algorithm: theme.compactAlgorithm,  // compact density
  // algorithm: [theme.darkAlgorithm, theme.compactAlgorithm], // both
};
```

## 5. Component Mapping Reference

| Current (Tailwind/HTML)    | Ant Design Vue v4              |
| -------------------------- | ------------------------------ |
| `<input>`                  | `<a-input>` / `<a-input-search>` |
| `<button>`                 | `<a-button>`                     |
| `<nav>` / `<ul>` / `<li>`  | `<a-menu>` / `<a-menu-item>`     |
| Custom header layout       | `<a-layout>` + `<a-layout-header>` |
| Custom footer              | `<a-layout-footer>`              |
| Badge (cart count)         | `<a-badge :count="n">`          |
| Search input               | `<a-input-search>`               |
| Mobile menu toggle         | `<a-button>` with menu icon      |
| Product grid               | `<a-card>` / `<a-row>/<a-col>`  |
| Product table (admin)      | `<a-table>`                      |
| Forms (login, register)    | `<a-form>` + `<a-form-item>`    |
| Modal dialogs              | `<a-modal>`                      |
| Toast notifications        | `message` / `notification` API   |
| Dropdown menus             | `<a-dropdown>`                   |
| Tabs                       | `<a-tabs>`                       |
| Pagination                 | `<a-pagination>`                 |

## 6. Layout Structure for App.vue

```vue
<template>
  <a-config-provider :theme="themeConfig">
    <a-layout class="min-h-screen">
      <a-layout-header>
        <div class="max-w-7xl mx-auto px-4 flex items-center justify-between">
          <router-link to="/" class="logo">EShop</router-link>
          <a-menu mode="horizontal" :selected-keys="selectedKeys">
            <a-menu-item key="products">
              <router-link to="/products">All Products</router-link>
            </a-menu-item>
          </a-menu>
          <a-input-search placeholder="Search products..." style="width: 300px" @search="handleSearch" />
          <a-badge :count="cart.itemCount">
            <a-button shape="circle" @click="$router.push('/cart')">
              <template #icon><ShoppingCartOutlined /></template>
            </a-button>
          </a-badge>
        </div>
      </a-layout-header>
      <a-layout-content>
        <router-view />
      </a-layout-content>
      <a-layout-footer>Footer content</a-layout-footer>
    </a-layout>
  </a-config-provider>
</template>
```

## 7. Utility APIs (No Manual Style Imports in v4)

```javascript
import { message, notification, Modal } from 'ant-design-vue';

// All work without separate style imports in v4 CSS-in-JS mode
message.success('Item added to cart');
notification.open({ message: 'Title', description: 'Content' });
Modal.confirm({ title: 'Confirm?', onOk: () => {} });
```

## 8. Remove Tailwind

After migration is complete:

```bash
npm uninstall tailwindcss @tailwindcss/vite
rm -rf src/style.css  # if only contains Tailwind directives
```

Remove `@tailwindcss/vite` from `vite.config.js` plugins and all `class="..."` Tailwind utility classes from templates.
