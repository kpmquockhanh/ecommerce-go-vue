<template>
  <a-layout class="min-h-100vh">
    <a-layout-header class="site-header">
      <div class="header-inner">
        <div class="header-left">
          <router-link to="/" class="logo">
            <span class="logo-mark">E</span>
            <span class="logo-text">Shop</span>
          </router-link>
          <a-menu mode="horizontal" class="header-nav">
            <a-menu-item key="products">
              <router-link to="/products">All Products</router-link>
            </a-menu-item>
          </a-menu>
        </div>

        <div class="header-search">
          <a-input-search
            v-model:value="searchQuery"
            placeholder="Search products..."
            @search="handleSearch"
            @keyup.enter="handleSearch"
            allowClear
          />
        </div>

        <div class="header-right">
          <router-link to="/cart" class="nav-cart-btn">
            <a-badge :count="cart.itemCount" :offset="[-4, 0]">
              <ShoppingCartOutlined class="nav-cart-icon" />
            </a-badge>
            <span class="nav-cart-text">Cart</span>
          </router-link>

          <template v-if="auth.isAuthenticated">
            <router-link to="/orders" class="nav-link-with-icon">
              <UnorderedListOutlined class="nav-link-icon" />
              <span>Orders</span>
            </router-link>
            <router-link
              v-if="auth.isAdmin"
              to="/admin"
              class="nav-link-with-icon nav-link--admin"
            >
              <SettingOutlined class="nav-link-icon" />
              <span>Admin</span>
            </router-link>
            <a-button type="text" @click="handleLogout" class="nav-logout-btn">Logout</a-button>
          </template>
          <template v-else>
            <router-link to="/login" class="nav-link--auth">
              Login
            </router-link>
            <router-link to="/register">
              <a-button type="primary">Sign Up</a-button>
            </router-link>
          </template>

          <a-button
            type="text"
            class="mobile-menu-toggle"
            @click="mobileMenuOpen = !mobileMenuOpen"
          >
            <MenuOutlined v-if="!mobileMenuOpen" />
            <CloseOutlined v-else />
          </a-button>
        </div>
      </div>
    </a-layout-header>

    <a-layout-content>
      <router-view />
    </a-layout-content>

    <a-layout-footer class="site-footer">
      <div class="footer-inner">
        <div class="footer-grid">
          <div>
            <div class="footer-brand">
              <span class="logo-mark--footer">E</span>
              <span class="logo-text--footer">Shop</span>
            </div>
            <p class="footer-description">
              Sustainable clothing for the conscious consumer. Quality fashion that cares for our planet.
            </p>
          </div>
          <div>
            <h4 class="footer-heading">Shop</h4>
            <ul class="footer-links">
              <li><router-link to="/products">All Products</router-link></li>
              <li><router-link to="/products?categories=2">Shirts</router-link></li>
              <li><router-link to="/products?categories=1">Pants</router-link></li>
              <li><router-link to="/products?categories=3">Shorts</router-link></li>
            </ul>
          </div>
          <div>
            <h4 class="footer-heading">Account</h4>
            <ul class="footer-links">
              <li><router-link to="/orders">Order History</router-link></li>
              <li><router-link to="/cart">Shopping Cart</router-link></li>
              <li><router-link to="/login">Sign In</router-link></li>
            </ul>
          </div>
          <div>
            <h4 class="footer-heading">Support</h4>
            <ul class="footer-links">
              <li><span>Free shipping on all orders</span></li>
              <li><span>30-day return policy</span></li>
              <li><span>Secure checkout</span></li>
            </ul>
          </div>
        </div>
        <div class="footer-bottom">
          <p class="footer-copyright">
            &copy; {{ new Date().getFullYear() }} E-Shop. All rights reserved.
          </p>
          <span class="footer-payment">Payments secure with Stripe</span>
        </div>
      </div>
    </a-layout-footer>
  </a-layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useCartStore } from '../stores/cart'
import {
  ShoppingCartOutlined,
  MenuOutlined,
  CloseOutlined,
  UnorderedListOutlined,
  SettingOutlined
} from '@ant-design/icons-vue'

const router = useRouter()
const auth = useAuthStore()
const cart = useCartStore()
const mobileMenuOpen = ref(false)
const searchQuery = ref('')

function handleSearch() {
  if (searchQuery.value?.trim()) {
    router.push({ path: '/products', query: { search: searchQuery.value.trim() } })
    mobileMenuOpen.value = false
  }
}

function handleLogout() {
  auth.logout()
  cart.items = []
  cart.total = 0
  cart.itemCount = 0
  mobileMenuOpen.value = false
  router.push('/')
}

onMounted(() => {
  if (auth.isAuthenticated) {
    cart.fetchCart()
  }
})
</script>
