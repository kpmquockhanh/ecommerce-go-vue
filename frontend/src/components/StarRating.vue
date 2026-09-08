<template>
  <div class="star-rating">
    <template v-for="i in 5" :key="i">
      <button
        v-if="interactive"
        type="button"
        @click="$emit('update:modelValue', i)"
        @mouseenter="hovered = i"
        @mouseleave="hovered = 0"
        :class="['star-btn', { 'star-btn--readonly': !interactive }]"
        :style="{ fontSize: sizeValue + 'px', color: getStarColor(i, hovered || modelValue) }"
      >
        {{ getStarIcon(i, hovered || modelValue) }}
      </button>
      <span
        v-else
        :style="{ fontSize: sizeValue + 'px', color: getStarColor(i, modelValue) }"
      >
        {{ getStarIcon(i, modelValue) }}
      </span>
    </template>
    <span v-if="count !== null" class="star-count">
      ({{ count }} {{ count === 1 ? 'review' : 'reviews' }})
    </span>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: Number,
    default: 0,
  },
  count: {
    type: Number,
    default: null,
  },
  interactive: Boolean,
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['sm', 'md', 'lg'].includes(v),
  },
})

defineEmits(['update:modelValue'])

const hovered = ref(0)

const sizeMap = {
  sm: 14,
  md: 18,
  lg: 24,
}

const sizeValue = computed(() => sizeMap[props.size])

function getStarIcon(index, value) {
  return index <= value ? '\u2605' : '\u2606'
}

function getStarColor(index, value) {
  return index <= value ? '#EAB308' : '#D1D5DB'
}
</script>
