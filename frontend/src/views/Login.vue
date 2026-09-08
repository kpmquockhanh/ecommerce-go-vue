<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <router-link to="/" class="inline-flex mb-24">
          <span class="logo-mark--auth">E</span>
          <span class="logo-text--auth">Shop</span>
        </router-link>
        <h1 class="auth-title">Welcome back</h1>
        <p class="auth-subtitle">
          Don't have an account?
          <router-link to="/register" class="auth-link">
            Sign up
          </router-link>
        </p>
      </div>

      <a-card :bordered="true">
        <a-alert
          v-if="auth.error"
          :message="auth.error"
          type="error"
          show-icon
          class="mb-24"
        />

        <a-form layout="vertical" @submit.prevent="handleSubmit">
          <a-form-item label="Email address" name="email">
            <a-input
              v-model:value="form.email"
              type="email"
              placeholder="you@example.com"
              size="large"
            />
          </a-form-item>
          <a-form-item label="Password" name="password">
            <a-input-password
              v-model:value="form.password"
              placeholder="Enter your password"
              size="large"
            />
          </a-form-item>

          <a-button
            type="primary"
            html-type="submit"
            size="large"
            block
            :loading="auth.loading"
          >
            {{ auth.loading ? 'Signing in...' : 'Sign in' }}
          </a-button>
        </a-form>
      </a-card>
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useCartStore } from '../stores/cart'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const cart = useCartStore()

const form = reactive({ email: '', password: '' })

async function handleSubmit() {
  try {
    await auth.login(form.email, form.password)
    await cart.mergeGuestCart()
    await cart.fetchCart()
    router.push(route.query.redirect || '/')
  } catch {}
}
</script>
