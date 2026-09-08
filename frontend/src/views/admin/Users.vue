<template>
  <div class="admin-page">
    <h1 class="admin-page-header">Users</h1>

    <div class="mb-16">
      <a-input-search
        v-model:value="search"
        placeholder="Search by name or email..."
        class="admin-search"
        allowClear
      />
    </div>

    <a-table
      :dataSource="filteredUsers"
      :columns="userColumns"
      :pagination="false"
      rowKey="id"
      class="admin-table"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <span class="admin-cell-bold">{{ record.first_name }} {{ record.last_name }}</span>
        </template>
        <template v-if="column.key === 'email'">
          <span class="admin-cell-muted">{{ record.email }}</span>
        </template>
        <template v-if="column.key === 'role'">
          <a-tag :class="record.role === 'admin' ? 'tag-plum' : 'tag-neutral'">{{ record.role }}</a-tag>
        </template>
        <template v-if="column.key === 'joined'">
          <span class="admin-cell-muted">{{ formatDate(record.created_at) }}</span>
        </template>
        <template v-if="column.key === 'actions'">
          <a-space>
            <a-button type="link" size="small" @click="editUser(record)">Edit</a-button>
            <a-button
              v-if="record.id !== currentUserId"
              type="link"
              danger
              size="small"
              @click="deleteUser(record)"
            >Delete</a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal
      v-model:open="showModal"
      title="Edit User Role"
      @ok="saveUserRole"
      @cancel="closeModal"
    >
      <a-form layout="vertical">
        <a-form-item label="Role">
          <a-select v-model:value="form.role">
            <a-select-option value="user">User</a-select-option>
            <a-select-option value="admin">Admin</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import api from '../../lib/api'
import { useAuthStore } from '../../stores/auth'

const authStore = useAuthStore()
const currentUserId = computed(() => authStore.user?.id)

const users = ref([])
const search = ref('')
const showModal = ref(false)
const editingUser = ref(null)
const form = ref({ role: 'user' })

const userColumns = [
  { title: 'Name', key: 'name' },
  { title: 'Email', key: 'email' },
  { title: 'Role', key: 'role' },
  { title: 'Joined', key: 'joined' },
  { title: 'Actions', key: 'actions' },
]

const filteredUsers = computed(() => {
  if (!search.value) return users.value
  const q = search.value.toLowerCase()
  return users.value.filter(u =>
    (u.first_name && u.first_name.toLowerCase().includes(q)) ||
    (u.last_name && u.last_name.toLowerCase().includes(q)) ||
    (u.email && u.email.toLowerCase().includes(q))
  )
})

function formatDate(dateStr) {
  return new Date(dateStr).toLocaleDateString()
}

async function loadUsers() {
  const { data } = await api.get('/admin/users')
  users.value = data.users || data
}

function editUser(user) {
  editingUser.value = user
  form.value = { role: user.role }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingUser.value = null
}

async function saveUserRole() {
  try {
    await api.put(`/admin/users/${editingUser.value.id}`, form.value)
    closeModal()
    await loadUsers()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to update user')
  }
}

async function deleteUser(user) {
  Modal.confirm({
    title: `Delete ${user.first_name} ${user.last_name}?`,
    okText: 'Delete',
    okType: 'danger',
    onOk: async () => {
      try {
        await api.delete(`/admin/users/${user.id}`)
        await loadUsers()
      } catch (err) {
        message.error(err.response?.data?.error || 'Failed to delete user')
      }
    },
  })
}

onMounted(loadUsers)
</script>
