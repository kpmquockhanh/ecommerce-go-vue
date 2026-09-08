<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <router-link to="/" class="inline-flex mb-24">
          <span class="logo-mark--auth">E</span>
          <span class="logo-text--auth">Shop</span>
        </router-link>
        <h1 class="auth-title">Create your account</h1>
        <p class="auth-subtitle">
          Already have an account?
          <router-link to="/login" class="auth-link">
            Sign in
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
          <div class="form-grid-2">
            <a-form-item label="First name" name="firstName">
              <a-input
                v-model:value="form.firstName"
                placeholder="John"
                size="large"
              />
            </a-form-item>
            <a-form-item label="Last name" name="lastName">
              <a-input
                v-model:value="form.lastName"
                placeholder="Doe"
                size="large"
              />
            </a-form-item>
          </div>
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
              placeholder="Create a password"
              size="large"
            />
            <p class="auth-hint">Must be at least 6 characters</p>
          </a-form-item>

          <a-button
            type="primary"
            html-type="submit"
            size="large"
            block
            :loading="auth.loading"
          >
            {{ auth.loading ? 'Creating account...' : 'Create account' }}
          </a-button>
        </a-form>
      </a-card>
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useCartStore } from '../stores/cart'

const router = useRouter()
const auth = useAuthStore()
const cart = useCartStore()

const form = reactive({
  firstName: '',
  lastName: '',
  email: '',
  password: '',
})

async function handleSubmit() {
  try {
    await auth.register({
      first_name: form.firstName,
      last_name: form.lastName,
      email: form.email,
      password: form.password,
    })
    await cart.mergeGuestCart()
    router.push('/')
  } catch {}
}
</script>
