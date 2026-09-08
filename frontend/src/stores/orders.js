import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../lib/api'

export const useOrderStore = defineStore('orders', () => {
  const orders = ref([])
  const currentOrder = ref(null)
  const pagination = ref({ page: 1, limit: 20, total: 0, total_pages: 0 })
  const loading = ref(false)
  const error = ref(null)

  async function initCheckout(idempotencyKey) {
    loading.value = true
    error.value = null
    try {
      const payload = {}
      if (idempotencyKey) {
        payload.idempotency_key = idempotencyKey
      }
      const { data } = await api.post('/checkout/init', payload)
      return data
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to initialize checkout'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function confirmCheckout(shippingAddress, idempotencyKey) {
    loading.value = true
    error.value = null
    try {
      const { data } = await api.post('/checkout/confirm', {
        shipping_address: shippingAddress,
        idempotency_key: idempotencyKey,
      })
      return data
    } catch (err) {
      error.value = err.response?.data?.error || 'Checkout failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchOrders(params = {}) {
    loading.value = true
    error.value = null
    try {
      const { data } = await api.get('/orders', { params })
      orders.value = data.orders
      pagination.value = data.pagination
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to fetch orders'
    } finally {
      loading.value = false
    }
  }

  async function fetchOrder(id) {
    loading.value = true
    error.value = null
    try {
      const { data } = await api.get(`/orders/${id}`)
      currentOrder.value = data
    } catch (err) {
      error.value = err.response?.data?.error || 'Order not found'
    } finally {
      loading.value = false
    }
  }

  return { orders, currentOrder, pagination, loading, error, initCheckout, confirmCheckout, fetchOrders, fetchOrder }
})
