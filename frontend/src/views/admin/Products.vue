<template>
  <div class="admin-page">
    <div class="admin-page-header--row">
      <h1 class="page-heading">Products</h1>
      <a-button type="primary" @click="showModal = true">
        <template #icon><PlusOutlined /></template>
        Add Product
      </a-button>
    </div>

    <a-table
      :dataSource="products"
      :columns="productColumns"
      :pagination="false"
      rowKey="id"
      class="admin-table"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <span class="admin-cell-bold">{{ record.name }}</span>
        </template>
        <template v-if="column.key === 'category'">
          <template v-if="record.category_names?.length">
            <a-tag v-for="name in record.category_names" :key="name">{{ name }}</a-tag>
          </template>
          <span v-else class="admin-hint">Uncategorized</span>
        </template>
        <template v-if="column.key === 'price'">
          ${{ (record.price / 100).toFixed(2) }}
        </template>
        <template v-if="column.key === 'status'">
          <a-select
            :value="record.status"
            size="small"
            style="width: 120px"
            @change="(val) => updateStatus(record, val)"
          >
            <a-select-option value="published">Published</a-select-option>
            <a-select-option value="draft">Draft</a-select-option>
          </a-select>
        </template>
        <template v-if="column.key === 'actions'">
          <a-space>
            <a-button type="link" size="small" @click="$router.push(`/admin/products/${record.id}`)">Manage</a-button>
            <a-button type="link" danger size="small" @click="deleteProduct(record)">Delete</a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal
      v-model:open="showModal"
      title="Add Product"
      :confirmLoading="saving"
      @ok="saveProduct"
      @cancel="closeModal"
    >
      <a-form layout="vertical">
        <a-form-item label="Name">
          <a-input v-model:value="form.name" required />
        </a-form-item>
        <a-form-item label="Description">
          <a-textarea v-model:value="form.description" :rows="3" />
        </a-form-item>
        <div class="admin-form-grid">
          <a-form-item label="Price (cents)">
            <a-input-number v-model:value="form.price" :min="0" class="w-full" />
          </a-form-item>
          <a-form-item label="Category">
            <a-select v-model:value="form.category" placeholder="Select category">
              <a-select-option v-for="cat in categories" :key="cat.id" :value="cat.name">
                {{ cat.name }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </div>
        <a-form-item label="Image">
          <a-upload
            :before-upload="handleImageSelect"
            :show-upload-list="false"
            accept="image/jpeg,image/png,image/webp"
          >
            <a-button>
              <template #icon><UploadOutlined /></template>
              Select Image
            </a-button>
          </a-upload>
          <p class="admin-hint">Max 5MB. JPG, PNG, or WebP.</p>
          <div v-if="imagePreview" class="mt-8">
            <img :src="imagePreview" class="admin-thumb-preview" />
            <a-button type="link" danger size="small" @click="removeImage">Remove</a-button>
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined, UploadOutlined } from '@ant-design/icons-vue'
import api from '../../lib/api'

const products = ref([])
const categories = ref([])
const showModal = ref(false)
const saving = ref(false)
const selectedImage = ref(null)
const imagePreview = ref(null)
const form = ref({ name: '', description: '', price: 0, category: '' })

const MAX_FILE_SIZE = 5 * 1024 * 1024
const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp']

const productColumns = [
  { title: 'Name', key: 'name' },
  { title: 'Category', key: 'category' },
  { title: 'Price', key: 'price' },
  { title: 'Status', key: 'status' },
  { title: 'Actions', key: 'actions' },
]

async function loadProducts() {
  const { data } = await api.get('/admin/products', { params: { limit: 100 } })
  products.value = data.products
}

async function loadCategories() {
  const { data } = await api.get('/admin/categories')
  categories.value = data.categories || []
}

function closeModal() {
  showModal.value = false
  selectedImage.value = null
  imagePreview.value = null
  form.value = { name: '', description: '', price: 0, category: '' }
}

function handleImageSelect(file) {
  if (!ALLOWED_TYPES.includes(file.type)) {
    message.error('Only JPG, PNG, and WebP images are allowed.')
    return false
  }
  if (file.size > MAX_FILE_SIZE) {
    message.error('Image must be less than 5MB.')
    return false
  }
  selectedImage.value = file
  imagePreview.value = URL.createObjectURL(file)
  return false
}

function removeImage() {
  selectedImage.value = null
  imagePreview.value = null
}

async function uploadImage(file) {
  const formData = new FormData()
  formData.append('image', file)
  const { data } = await api.post('/images/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
  return data.path
}

async function saveProduct() {
  saving.value = true
  try {
    let imagePath = null
    if (selectedImage.value) {
      imagePath = await uploadImage(selectedImage.value)
    }
    const payload = { ...form.value }
    if (imagePath) payload.images = [imagePath]

    await api.post('/admin/products', payload)
    closeModal()
    await loadProducts()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to save product')
  } finally {
    saving.value = false
  }
}

async function updateStatus(product, status) {
  try {
    await api.put(`/admin/products/${product.id}`, { status })
    product.status = status
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to update status')
  }
}

async function updateCategory(product, categoryId) {
  try {
    const categoryIds = categoryId ? [categoryId] : []
    await api.put(`/admin/products/${product.id}`, { category_ids: categoryIds })
    const cat = categories.value.find(c => c.id === categoryId)
    product.category_names = cat ? [cat.name] : []
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to update category')
  }
}

async function deleteProduct(product) {
  Modal.confirm({
    title: `Delete ${product.name}?`,
    okText: 'Delete',
    okType: 'danger',
    onOk: async () => {
      await api.delete(`/admin/products/${product.id}`)
      await loadProducts()
    },
  })
}

onMounted(() => {
  loadProducts()
  loadCategories()
})
</script>
