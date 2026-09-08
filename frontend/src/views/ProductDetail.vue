<template>
  <div class="page-container">
    <a-breadcrumb class="breadcrumb-spacer">
      <a-breadcrumb-item><router-link to="/">Home</router-link></a-breadcrumb-item>
      <a-breadcrumb-item><router-link to="/products">Products</router-link></a-breadcrumb-item>
      <a-breadcrumb-item>{{ product?.name || 'Loading...' }}</a-breadcrumb-item>
    </a-breadcrumb>

    <div v-if="loading" class="product-detail-skeleton">
      <a-skeleton :paragraph="false" active>
        <a-skeleton-image class="w-full skeleton-image-square" />
      </a-skeleton>
      <div class="skeleton-info-col">
        <a-skeleton :paragraph="{ rows: 1 }" active class="skeleton-w-33" />
        <a-skeleton :paragraph="{ rows: 1 }" active class="skeleton-w-75" />
        <a-skeleton :paragraph="{ rows: 1 }" active class="skeleton-w-25-h32" />
        <a-skeleton :paragraph="{ rows: 3 }" active />
      </div>
    </div>

    <div v-else-if="error" class="empty-state--flush">
      <a-alert :message="error" type="error" show-icon />
      <router-link to="/products" class="inline-block mt-16 text-primary">
        Back to Products
      </router-link>
    </div>

    <div v-else-if="product" class="product-detail-grid">
      <div>
        <div class="image-container">
          <img
            v-if="product.images?.length"
            :src="product.images[activeImage]"
            :alt="product.name"
            class="product-card-image"
          />
          <div v-else class="image-placeholder">
            <span class="image-placeholder-text">No image</span>
          </div>
        </div>

        <div v-if="product.images?.length > 1" class="thumbnail-strip">
          <button
            v-for="(_, idx) in product.images"
            :key="idx"
            @click="activeImage = idx"
            :class="['thumbnail-btn', { active: activeImage === idx }]"
          >
            <img
              :src="product.images[idx]"
              :alt="`${product.name} view ${idx + 1}`"
            />
          </button>
        </div>
      </div>

      <div class="product-info-col">
        <div class="mb-8">
          <template v-if="product.categories?.length">
            <a-tag v-for="cat in product.categories" :key="cat.id" color="green" class="product-category-tag">
              {{ cat.name }}
            </a-tag>
          </template>
        </div>

        <h1 class="text-3xl font-bold text-primary-color mb-16">{{ product.name }}</h1>

        <div v-if="product.avg_rating" class="product-rating-row">
          <StarRating :modelValue="Math.round(product.avg_rating)" size="md" />
          <span class="text-sm text-secondary">
            {{ product.avg_rating.toFixed(1) }} ({{ product.review_count }} {{ product.review_count === 1 ? 'review' : 'reviews' }})
          </span>
        </div>

        <div class="mb-24">
          <PriceDisplay :price="displayPrice" size="lg" />
          <span v-if="priceModifier !== 0" class="text-sm text-secondary ml-8">
            ({{ priceModifier > 0 ? '+' : '' }}{{ formatPrice(priceModifier) }})
          </span>
        </div>

        <p class="text-secondary lh-normal mb-32">
          {{ product.description }}
        </p>

        <div v-if="product.option_groups?.length" class="mb-32">
          <div v-for="group in product.option_groups" :key="group.id" class="mb-16">
            <h3 class="product-variant-label mb-8">{{ group.name }}</h3>
            <div class="flex flex-wrap gap-8">
              <a-button
                v-for="value in group.values"
                :key="value.id"
                @click="selectOptionValue(group.id, value.id)"
                :type="selectedOptionValues[group.id] === value.id ? 'primary' : 'default'"
                :disabled="isOptionValueDisabled(group.id, value.id)"
              >
                {{ value.value }}
                <span v-if="value.price_modifier !== 0" class="text-xs">
                  ({{ value.price_modifier > 0 ? '+' : '' }}{{ formatPrice(value.price_modifier) }})
                </span>
              </a-button>
            </div>
          </div>
          <p v-if="selectedVariant" class="mt-8 text-sm text-secondary">
            <span v-if="selectedVariant.stock > 5" class="stock-in">In Stock</span>
            <span v-else-if="selectedVariant.stock > 0" class="stock-low">Only {{ selectedVariant.stock }} left</span>
            <span v-else class="stock-out">Out of Stock</span>
          </p>
          <p v-else-if="hasAllSelections" class="mt-8 text-sm text-secondary stock-out">
            This combination is unavailable
          </p>
        </div>

        <div v-else-if="product.variants?.length" class="mb-32">
          <h3 class="product-variant-label">Select Option</h3>
          <div class="flex flex-wrap gap-8">
            <a-button
              v-for="variant in product.variants"
              :key="variant.id"
              @click="selectedVariant = variant"
              :disabled="variant.stock === 0"
              :type="selectedVariant?.id === variant.id ? 'primary' : 'default'"
              :danger="variant.stock === 0"
            >
              {{ variant.label || variant.id }}
            </a-button>
          </div>
          <p v-if="selectedVariant" class="mt-8 text-sm text-secondary">
            <span v-if="selectedVariant.stock > 5" class="stock-in">In Stock</span>
            <span v-else-if="selectedVariant.stock > 0" class="stock-low">Only {{ selectedVariant.stock }} left</span>
            <span v-else class="stock-out">Out of Stock</span>
          </p>
        </div>

        <div class="product-actions">
          <a-button
            type="primary"
            size="large"
            :disabled="!canAddToCart"
            :loading="adding"
            class="product-add-btn"
            @click="handleAddToCart"
          >
            {{ adding ? 'Adding...' : 'Add to Cart' }}
          </a-button>
          <a-button size="large">
            <template #icon><HeartOutlined /></template>
          </a-button>
        </div>

        <div class="product-features">
          <div class="product-feature-item">
            <CarOutlined class="product-feature-icon" />
            <p class="product-feature-text">Free Shipping</p>
          </div>
          <div class="product-feature-item">
            <SwapOutlined class="product-feature-icon" />
            <p class="product-feature-text">30-Day Returns</p>
          </div>
          <div class="product-feature-item">
            <SafetyCertificateOutlined class="product-feature-icon" />
            <p class="product-feature-text">Secure Payment</p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="product" class="reviews-section">
      <div class="reviews-section-inner">
        <h2 class="text-2xl font-bold text-primary-color mb-32">Customer Reviews</h2>

        <div v-if="reviews.length === 0" class="empty-state--compact">
          <a-empty description="No reviews yet. Be the first to share your thoughts!" />
        </div>

        <div v-else class="reviews-list">
          <a-card
            v-for="review in reviews"
            :key="review.id"
            :bordered="true"
          >
            <div class="review-header">
              <div class="review-user">
                <a-avatar class="bg-primary-light text-primary">
                  {{ review.user_name?.charAt(0)?.toUpperCase() || '?' }}
                </a-avatar>
                <div>
                  <p class="review-user-name">{{ review.user_name }}</p>
                  <StarRating :modelValue="review.rating" size="sm" />
                </div>
              </div>
            </div>
            <p class="text-secondary lh-normal">{{ review.comment }}</p>
          </a-card>
        </div>

        <a-card v-if="auth.isAuthenticated" :bordered="true">
          <h3 class="font-semibold text-primary-color mb-16">Write a Review</h3>
          <div class="mb-16">
            <label class="review-form-label">Rating</label>
            <StarRating
              v-model="newReview.rating"
              :interactive="true"
              size="lg"
            />
          </div>
          <div class="mb-16">
            <label class="review-form-label">Your Review</label>
            <a-textarea
              v-model:value="newReview.comment"
              :rows="4"
              placeholder="Share your experience with this product..."
            />
          </div>
          <a-button
            type="primary"
            :disabled="!newReview.rating || submittingReview"
            :loading="submittingReview"
            @click="submitReview"
          >
            {{ submittingReview ? 'Submitting...' : 'Submit Review' }}
          </a-button>
        </a-card>
        <div v-else class="empty-state--login">
          <p class="text-secondary mb-12">Have this product? Share your experience.</p>
          <router-link to="/login" class="text-primary font-medium">
            Sign in to write a review
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useProductStore } from '../stores/products'
import { useCartStore } from '../stores/cart'
import { useAuthStore } from '../stores/auth'
import { message } from 'ant-design-vue'
import { HeartOutlined, CarOutlined, SwapOutlined, SafetyCertificateOutlined } from '@ant-design/icons-vue'
import api from '../lib/api'
import StarRating from '../components/StarRating.vue'
import PriceDisplay from '../components/PriceDisplay.vue'

const route = useRoute()
const productStore = useProductStore()
const cart = useCartStore()
const auth = useAuthStore()

const product = computed(() => productStore.currentProduct)
const loading = computed(() => productStore.loading)
const error = computed(() => productStore.error)

const activeImage = ref(0)
const selectedVariant = ref(null)
const selectedOptionValues = reactive({})
const adding = ref(false)
const reviews = ref([])
const newReview = reactive({ rating: 0, comment: '' })
const submittingReview = ref(false)

const hasOptionGroups = computed(() => product.value?.option_groups?.length > 0)

const hasAllSelections = computed(() => {
  if (!hasOptionGroups.value) return false
  return product.value.option_groups.every(g => selectedOptionValues[g.id])
})

const matchedVariant = computed(() => {
  if (!hasOptionGroups.value || !hasAllSelections.value) return null
  const selectedIds = Object.values(selectedOptionValues).sort()
  return product.value.variants.find(v => {
    const vIds = v.option_values.map(ov => ov.option_value_id).sort()
    return vIds.length === selectedIds.length && vIds.every((id, i) => id === selectedIds[i])
  }) || null
})

const priceModifier = computed(() => {
  if (!hasOptionGroups.value || !hasAllSelections.value) return 0
  let modifier = 0
  for (const group of product.value.option_groups) {
    const selectedValueId = selectedOptionValues[group.id]
    if (selectedValueId) {
      const value = group.values.find(v => v.id === selectedValueId)
      if (value) modifier += value.price_modifier || 0
    }
  }
  return modifier
})

const displayPrice = computed(() => {
  return (product.value?.price || 0) + priceModifier.value
})

function formatPrice(cents) {
  return '$' + (Math.abs(cents) / 100).toFixed(2)
}

function selectOptionValue(groupId, valueId) {
  if (selectedOptionValues[groupId] === valueId) {
    delete selectedOptionValues[groupId]
  } else {
    selectedOptionValues[groupId] = valueId
  }
  // Match to variant
  selectedVariant.value = matchedVariant.value
}

function isOptionValueDisabled(groupId, valueId) {
  if (!product.value?.variants?.length) return false
  // Check if any variant with this value exists and is in stock
  return !product.value.variants.some(v => {
    const hasValue = v.option_values.some(ov => ov.option_value_id === valueId)
    if (!hasValue) return false
    // Check other selections match
    for (const g of product.value.option_groups) {
      if (g.id === groupId) continue
      const sel = selectedOptionValues[g.id]
      if (sel && !v.option_values.some(ov => ov.option_value_id === sel)) return false
    }
    return v.stock > 0
  })
}

const canAddToCart = computed(() => {
  if (!product.value) return false
  if (hasOptionGroups.value) {
    if (!hasAllSelections.value) return false
    if (matchedVariant.value && matchedVariant.value.stock === 0) return false
    if (!matchedVariant.value) return false
  } else if (product.value.variants?.length) {
    if (!selectedVariant.value) return false
    if (selectedVariant.value.stock === 0) return false
  }
  return true
})

async function handleAddToCart() {
  adding.value = true
  try {
    const variantId = selectedVariant.value?.id || null
    await cart.addItem(product.value.id, variantId)
    message.success('Added to cart!')
  } catch {
    // Error handled by store
  } finally {
    adding.value = false
  }
}

async function loadReviews() {
  try {
    const { data } = await api.get(`/reviews/product/${product.value.id}`)
    reviews.value = data.reviews
  } catch {
    // Silent fail
  }
}

async function submitReview() {
  submittingReview.value = true
  try {
    await api.post('/reviews', {
      product_id: product.value.id,
      rating: newReview.rating,
      comment: newReview.comment,
    })
    newReview.rating = 0
    newReview.comment = ''
    message.success('Review submitted!')
    await loadReviews()
  } catch {
    // Silent fail
  } finally {
    submittingReview.value = false
  }
}

onMounted(async () => {
  await productStore.fetchProduct(route.params.slug)
  if (product.value) {
    await loadReviews()
  }
})
</script>
