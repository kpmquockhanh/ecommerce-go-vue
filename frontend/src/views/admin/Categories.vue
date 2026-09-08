<template>
  <div class="admin-page">
    <div class="admin-page-header--row">
      <h1 class="page-heading">Categories</h1>
      <a-button type="primary" @click="showModal = true">
        <template #icon><PlusOutlined /></template>
        Add Category
      </a-button>
    </div>

    <a-table
      :dataSource="categories"
      :columns="categoryColumns"
      :pagination="false"
      rowKey="id"
      class="admin-table"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <span class="admin-cell-bold">{{ record.name }}</span>
        </template>
        <template v-if="column.key === 'product_count'">
          <span class="admin-cell-muted">{{ record.product_count }}</span>
        </template>
        <template v-if="column.key === 'actions'">
          <a-space>
            <a-button type="link" size="small" @click="editCategory(record)">Edit</a-button>
            <a-button type="link" danger size="small" @click="deleteCategory(record)">Delete</a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal
      v-model:open="showModal"
      :title="editingCategory ? 'Edit Category' : 'Add Category'"
      @ok="saveCategory"
      @cancel="closeModal"
    >
      <a-form layout="vertical">
        <a-form-item label="Name">
          <a-input v-model:value="form.name" required />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import api from '../../lib/api'

const categories = ref([])
const showModal = ref(false)
const editingCategory = ref(null)
const form = ref({ name: '' })

const categoryColumns = [
  { title: 'Name', key: 'name' },
  { title: 'Products', key: 'product_count' },
  { title: 'Actions', key: 'actions' },
]

async function loadCategories() {
  const { data } = await api.get('/admin/categories')
  categories.value = data.categories || []
}

function editCategory(category) {
  editingCategory.value = category
  form.value = { name: category.name }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingCategory.value = null
  form.value = { name: '' }
}

async function saveCategory() {
  try {
    if (editingCategory.value) {
      await api.put(`/admin/categories/${editingCategory.value.id}`, form.value)
    } else {
      await api.post('/admin/categories/create', form.value)
    }
    closeModal()
    await loadCategories()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to save category')
  }
}

async function deleteCategory(category) {
  Modal.confirm({
    title: `Delete "${category.name}"?`,
    okText: 'Delete',
    okType: 'danger',
    onOk: async () => {
      try {
        await api.delete(`/admin/categories/${category.id}`)
        await loadCategories()
      } catch (err) {
        message.error(err.response?.data?.error || 'Failed to delete category')
      }
    },
  })
}

onMounted(loadCategories)
</script>
