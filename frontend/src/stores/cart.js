import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../lib/api'

export const useCartStore = defineStore('cart', () => {
  const items = ref([])
  const total = ref(0)
  const itemCount = ref(0)
  const loading = ref(false)

  const formattedTotal = computed(() => `$${(total.value / 100).toFixed(2)}`)

  async function fetchCart() {
    loading.value = true
    try {
      const { data } = await api.get('/cart')
      items.value = data.items
      total.value = data.total
      itemCount.value = data.item_count
    } catch {
      items.value = []
      total.value = 0
      itemCount.value = 0
    } finally {
      loading.value = false
    }
  }

  async function addItem(productId, variantId = null, quantity = 1) {
    loading.value = true
    try {
      await api.post('/cart/items', {
        product_id: productId,
        variant_id: variantId || 0,
        quantity,
      })
      await fetchCart()
    } finally {
      loading.value = false
    }
  }

  async function updateItem(itemId, quantity) {
    loading.value = true
    try {
      await api.put(`/cart/items/${itemId}`, { quantity })
      await fetchCart()
    } finally {
      loading.value = false
    }
  }

  async function removeItem(itemId) {
    loading.value = true
    try {
      await api.delete(`/cart/items/${itemId}`)
      await fetchCart()
    } finally {
      loading.value = false
    }
  }

  async function mergeGuestCart() {
    const sessionId = localStorage.getItem('session_id')
    if (!sessionId) return
    
    try {
      await api.post('/cart/merge', { session_id: sessionId })
      localStorage.removeItem('session_id')
      await fetchCart()
    } catch {
      // Ignore merge errors
    }
  }

  return {
    items,
    total,
    itemCount,
    loading,
    formattedTotal,
    fetchCart,
    addItem,
    updateItem,
    removeItem,
    mergeGuestCart,
  }
})
