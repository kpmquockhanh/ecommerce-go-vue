<template>
  <div>
    <section class="hero">
      <div class="hero-image-container bg-hero-gradient" />
      <div class="hero-image-container">
        <img
          v-if="heroImage"
          :src="heroImage"
          alt="Featured collection"
          class="hero-image"
        />
        <div v-else class="hero-image bg-hero-fallback" />
      </div>
      <div class="hero-content">
        <div class="hero-text">
          <a-tag class="tag-success mb-24">New Season Collection</a-tag>
          <h1 class="hero-title">
            Sustainable Style<br />for Modern Living
          </h1>
          <p class="hero-subtitle">
            Discover our curated collection of eco-friendly clothing. Quality craftsmanship meets conscious design.
          </p>
          <div class="hero-actions">
            <router-link to="/products">
              <a-button type="primary" size="large">Shop Collection</a-button>
            </router-link>
            <router-link to="/products?categories=2">
              <a-button size="large" class="hero-btn-ghost">
                View Shirts
              </a-button>
            </router-link>
          </div>
        </div>
      </div>
    </section>

    <section class="features-bar">
      <div class="features-bar-inner">
        <div class="grid-4">
          <div v-for="feature in features" :key="feature.title" class="feature-item">
            <div class="feature-icon-wrap">
              <component :is="feature.icon" class="feature-icon" />
            </div>
            <div>
              <p class="feature-title">{{ feature.title }}</p>
              <p class="feature-desc">{{ feature.desc }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="section">
      <div class="text-center mb-48">
        <h2 class="section-title">Shop by Category</h2>
        <p class="section-subtitle--centered">
          Find exactly what you're looking for in our curated collections
        </p>
      </div>
      <div class="grid-3">
        <router-link
          v-for="category in categories"
          :key="category.id"
          :to="`/products?categories=${category.id}`"
          class="category-card"
        >
          <div class="category-card-bg" :style="{ background: category.bgStyle }" />
          <div class="category-card-overlay" />
          <div class="category-card-content">
            <h3 class="category-card-title">{{ category.name }}</h3>
            <p class="category-card-desc">{{ category.description }}</p>
          </div>
        </router-link>
      </div>
    </section>

    <section class="bg-white">
      <div class="section">
        <div class="section-header">
          <div>
            <h2 class="section-title">Featured Products</h2>
            <p class="section-subtitle">Handpicked favorites just for you</p>
          </div>
          <router-link to="/products" class="section-link">
            View All <RightOutlined />
          </router-link>
        </div>

        <div v-if="loading" class="grid-3">
          <a-card v-for="i in 6" :key="i" :bordered="true">
            <a-skeleton :paragraph="false" active>
              <a-skeleton-image class="w-full skeleton-image-sq" />
            </a-skeleton>
            <a-skeleton :paragraph="{ rows: 2 }" active />
          </a-card>
        </div>

        <div v-else-if="products.length" class="grid-3">
          <ProductCard
            v-for="product in products"
            :key="product.id"
            :product="product"
          />
        </div>

        <div v-else class="empty-state--flush">
          <p class="section-subtitle">No products available yet.</p>
        </div>
      </div>
    </section>

    <section class="section">
      <div class="cta-section">
        <div class="cta-text">
          <h2 class="cta-title">
            Join the Sustainable Fashion Movement
          </h2>
          <p class="cta-description">
            Sign up for exclusive offers, new arrivals, and tips on sustainable living.
          </p>
          <div class="cta-form">
            <a-input
              placeholder="Enter your email"
              class="cta-input"
            />
            <a-button class="cta-btn">Subscribe</a-button>
          </div>
        </div>
        <div class="cta-stats">
          <div class="cta-stat">
            <p class="cta-stat-value">500+</p>
            <p class="cta-stat-label">Happy Customers</p>
          </div>
          <div class="cta-divider" />
          <div class="cta-stat">
            <p class="cta-stat-value">100%</p>
            <p class="cta-stat-label">Eco Friendly</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { RightOutlined, CarOutlined, SwapOutlined, SafetyCertificateOutlined, GlobalOutlined } from '@ant-design/icons-vue'
import api from '../lib/api'
import ProductCard from '../components/ProductCard.vue'

const products = ref([])
const loading = ref(true)
const heroImage = ref(null)

const features = [
  { title: 'Free Shipping', desc: 'On all orders', icon: CarOutlined },
  { title: 'Easy Returns', desc: '30-day policy', icon: SwapOutlined },
  { title: 'Secure Checkout', desc: 'Stripe powered', icon: SafetyCertificateOutlined },
  { title: 'Eco Friendly', desc: 'Sustainable materials', icon: GlobalOutlined },
]

const categories = [
  {
    name: 'Shirts',
    id: 2,
    description: 'Comfortable everyday wear',
    bgStyle: 'linear-gradient(135deg, #52B788, #2D6A4F)',
  },
  {
    name: 'Pants',
    id: 1,
    description: 'Durable and stylish',
    bgStyle: 'linear-gradient(135deg, #40916C, #1B4332)',
  },
  {
    name: 'Shorts',
    id: 3,
    description: 'Perfect for warm days',
    bgStyle: 'linear-gradient(135deg, #74C69D, #52B788)',
  },
]

onMounted(async () => {
  try {
    const { data } = await api.get('/products', { params: { limit: 6 } })
    products.value = data.products
  } catch {
    // Use empty array on error
  } finally {
    loading.value = false
  }
})
</script>
