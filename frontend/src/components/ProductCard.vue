<template>
  <router-link
    :to="`/products/${product.slug}`"
    class="product-card"
  >
    <div class="product-card-image-wrap">
      <img
        v-if="product.images?.length"
        :src="product.images[0]"
        :alt="product.name"
        class="product-card-image"
      />
      <div v-else class="image-placeholder">
        <span class="image-placeholder-text--sm">No image</span>
      </div>

      <div v-if="badge" class="product-card-badge">
        <a-tag :class="badgeClass">{{ badge }}</a-tag>
      </div>

      <div class="product-card-overlay">
        <a-button
          block
          @click.prevent="handleQuickAdd"
          :disabled="!hasStock"
        >
          {{ !hasStock ? 'Out of Stock' : 'Add to Cart' }}
        </a-button>
      </div>
    </div>

    <div class="product-card-body">
      <p v-if="product.categories?.length" class="product-card-category">
        {{ product.categories.map(c => c.name).join(', ') }}
      </p>
      <h3 class="product-card-name">
        {{ product.name }}
      </h3>
      <div class="flex-center gap-8 mb-8">
        <StarRating
          v-if="product.avg_rating"
          :modelValue="Math.round(product.avg_rating)"
          :count="product.review_count"
          size="sm"
        />
      </div>
      <PriceDisplay
        :price="product.price"
        size="md"
      />
    </div>
  </router-link>
</template>

<script setup>
import { computed } from 'vue'
import { message } from 'ant-design-vue'
import StarRating from './StarRating.vue'
import PriceDisplay from './PriceDisplay.vue'
import { useCartStore } from '../stores/cart'

const props = defineProps({
  product: {
    type: Object,
    required: true,
  },
})

const cart = useCartStore()

const hasStock = computed(() => {
  if (!props.product.variants?.length) return true
  return props.product.variants.some((v) => v.stock > 0)
})

const badge = computed(() => {
  if (props.product.is_new) return 'New'
  if (props.product.on_sale) return 'Sale'
  if (hasStock.value && props.product.variants?.some((v) => v.stock > 0 && v.stock <= 5)) return 'Low Stock'
  return null
})

const badgeClass = computed(() => {
  if (props.product.is_new) return 'tag-photo-badge tag-photo-badge--new'
  if (props.product.on_sale) return 'tag-photo-badge tag-photo-badge--sale'
  return 'tag-photo-badge tag-photo-badge--low-stock'
})

async function handleQuickAdd() {
  if (!hasStock.value) return
  const defaultVariant = props.product.variants?.find((v) => v.stock > 0) || null
  try {
    await cart.addItem(props.product.id, defaultVariant?.id || null)
    message.success('Added to cart!')
  } catch {
    // Error handled by store
  }
}
</script>
