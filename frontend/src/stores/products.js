import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../lib/api'

export const useProductStore = defineStore('products', () => {
  const products = ref([])
  const currentProduct = ref(null)
  const pagination = ref({ page: 1, limit: 20, total: 0, total_pages: 0 })
  const loading = ref(false)
  const error = ref(null)

  async function fetchProducts(params = {}) {
    loading.value = true
    error.value = null
    try {
      const { data } = await api.get('/products', { params })
      products.value = data.products
      pagination.value = data.pagination
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to fetch products'
    } finally {
      loading.value = false
    }
  }

  async function fetchProduct(slug) {
    loading.value = true
    error.value = null
    try {
      const { data } = await api.get(`/products/${slug}`)
      currentProduct.value = data
    } catch (err) {
      error.value = err.response?.data?.error || 'Product not found'
    } finally {
      loading.value = false
    }
  }

  return { products, currentProduct, pagination, loading, error, fetchProducts, fetchProduct }
})
