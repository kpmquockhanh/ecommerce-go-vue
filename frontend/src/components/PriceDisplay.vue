<template>
  <div :class="['price-display', `price-display--${size}`]">
    <span v-if="originalPrice && originalPrice > price" class="price-original">
      {{ formatPrice(originalPrice) }}
    </span>
    <span :class="{ 'price-sale': sale }">
      {{ formatPrice(price) }}
    </span>
    <a-tag
      v-if="showSavePercent && originalPrice && originalPrice > price"
      color="error"
      class="ml-8"
    >
      Save {{ savePercent }}%
    </a-tag>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { formatPrice } from '../lib/utils'

const props = defineProps({
  price: {
    type: Number,
    required: true,
  },
  originalPrice: {
    type: Number,
    default: null,
  },
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['sm', 'md', 'lg'].includes(v),
  },
  sale: Boolean,
  showSavePercent: Boolean,
})

const savePercent = computed(() => {
  if (!props.originalPrice || props.originalPrice <= props.price) return 0
  return Math.round(((props.originalPrice - props.price) / props.originalPrice) * 100)
})
</script>
