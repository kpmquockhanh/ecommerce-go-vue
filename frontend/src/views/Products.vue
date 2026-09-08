<template>
  <div class="page-container">
    <div class="mb-32">
      <h1 class="page-heading">All Products</h1>
      <p class="page-subtitle">
        {{ productStore.pagination.total || 0 }} products found
      </p>
    </div>

    <div class="flex gap-32">
      <div class="filter-panel">
          <div class="filter-header">
            <h3 class="font-semibold text-primary-color m-0">Filters</h3>
            <a-button
              v-if="hasActiveFilters"
              type="link"
              size="small"
              @click="clearFilters"
            >
              Clear all
            </a-button>
          </div>

          <div class="filter-section">
            <h4 class="filter-label">Category</h4>
            <a-space direction="vertical" :size="8" class="w-full">
              <a-checkbox
                v-for="cat in categories"
                :key="cat.id"
                :checked="filters.categories.includes(cat.id)"
                @change="(e) => toggleCategory(cat.id, e.target.checked)"
              >
                <span class="filter-category-text">{{ cat.name }}</span>
              </a-checkbox>
            </a-space>
          </div>

          <div class="filter-section">
            <h4 class="filter-label">Price Range</h4>
            <div class="filter-price-row">
              <a-input-number
                v-model:value="filters.minPrice"
                placeholder="Min"
                :min="0"
                class="w-half"
              />
              <a-input-number
                v-model:value="filters.maxPrice"
                placeholder="Max"
                :min="0"
                class="w-half"
              />
            </div>
          </div>

          <a-button type="primary" block @click="applyFilters">
            Apply Filters
          </a-button>
        </div>

      <div class="flex-1">
        <div class="toolbar">
          <a-input-search
            v-model:value="searchQuery"
            placeholder="Search products..."
            class="w-288"
            @search="applyFilters"
            @keyup.enter="applyFilters"
            allowClear
          />
          <a-select
            v-model:value="sortBy"
            @change="applyFilters"
            class="w-200"
          >
            <a-select-option value="newest">Newest</a-select-option>
            <a-select-option value="price_asc">Price: Low to High</a-select-option>
            <a-select-option value="price_desc">Price: High to Low</a-select-option>
            <a-select-option value="name">Name: A to Z</a-select-option>
          </a-select>
        </div>

        <div v-if="hasActiveFilters" class="active-filters">
          <a-tag
            v-for="catId in filters.categories"
            :key="catId"
            color="green"
            closable
            @close="removeCategory(catId)"
          >
            {{ categories.find(c => c.id === catId)?.name || catId }}
          </a-tag>
          <a-tag
            v-if="filters.minPrice || filters.maxPrice"
            color="green"
            closable
            @close="clearPriceFilter"
          >
            ${{ filters.minPrice || 0 }} - ${{ filters.maxPrice || '\u221E' }}
          </a-tag>
        </div>

        <div v-if="productStore.loading" class="skeleton-grid">
          <a-card v-for="i in 6" :key="i" :bordered="true">
            <a-skeleton :paragraph="false" active>
              <a-skeleton-image class="w-full skeleton-image-sq" />
            </a-skeleton>
            <a-skeleton :paragraph="{ rows: 2 }" active />
          </a-card>
        </div>

        <div v-else-if="productStore.products.length === 0" class="empty-state">
          <a-empty description="No products found">
            <template #image>
              <SearchOutlined class="text-4xl text-placeholder" />
            </template>
            <p class="empty-state-text">Try adjusting your filters or search terms</p>
            <a-button @click="clearFilters">Clear Filters</a-button>
          </a-empty>
        </div>

        <div v-else class="grid-3">
          <ProductCard
            v-for="product in productStore.products"
            :key="product.id"
            :product="product"
          />
        </div>

        <div class="mt-40 text-center">
          <a-pagination
            v-model:current="currentPage"
            :total="productStore.pagination.total_pages * 10"
            :pageSize="10"
            show-less-items
            @change="goToPage"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useProductStore } from '../stores/products'
import { SearchOutlined } from '@ant-design/icons-vue'
import ProductCard from '../components/ProductCard.vue'
import api from '../lib/api'

const route = useRoute()
const router = useRouter()
const productStore = useProductStore()

const searchQuery = ref(route.query.search || '')
const sortBy = ref('newest')
const currentPage = ref(1)

const categories = ref([])

const filters = reactive({
  categories: route.query.categories ? route.query.categories.split(',').map(Number) : [],
  minPrice: null,
  maxPrice: null,
})

const hasActiveFilters = computed(() => {
  return filters.categories.length > 0 || filters.minPrice || filters.maxPrice
})

function toggleCategory(cat, checked) {
  if (checked) {
    filters.categories.push(cat)
  } else {
    filters.categories = filters.categories.filter((c) => c !== cat)
  }
}

function applyFilters() {
  const params = {
    page: currentPage.value,
    sort: sortBy.value,
  }

  if (searchQuery.value) params.search = searchQuery.value
  if (filters.categories.length > 0) params.categories = filters.categories.join(',')
  if (filters.minPrice) params.min_price = filters.minPrice * 100
  if (filters.maxPrice) params.max_price = filters.maxPrice * 100

  productStore.fetchProducts(params)

  const query = {}
  if (searchQuery.value) query.search = searchQuery.value
  if (filters.categories.length > 0) query.categories = filters.categories.join(',')
  router.replace({ query })
}

function goToPage(page) {
  currentPage.value = page
  applyFilters()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function removeCategory(cat) {
  filters.categories = filters.categories.filter((c) => c !== cat)
  applyFilters()
}

function clearPriceFilter() {
  filters.minPrice = null
  filters.maxPrice = null
  applyFilters()
}

function clearFilters() {
  searchQuery.value = ''
  filters.categories = []
  filters.minPrice = null
  filters.maxPrice = null
  currentPage.value = 1
  applyFilters()
}

onMounted(async () => {
  try {
    const { data } = await api.get('/categories')
    categories.value = data.categories
  } catch {
    categories.value = []
  }
  applyFilters()
})
</script>
