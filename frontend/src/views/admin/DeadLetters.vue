<template>
  <div class="admin-page">
    <h1 class="admin-page-header">Dead Letters</h1>

    <a-empty v-if="deadLetters.length === 0" description="No failed jobs" />

    <a-table
      v-else
      :dataSource="deadLetters"
      :columns="dlColumns"
      :pagination="false"
      rowKey="id"
      class="admin-table"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'queue_name'">
          <span class="admin-cell-bold">{{ record.queue_name }}</span>
        </template>
        <template v-if="column.key === 'job_type'">
          <span class="admin-cell-muted">{{ record.job_type }}</span>
        </template>
        <template v-if="column.key === 'error'">
          <span class="admin-cell-ellipsis">
            {{ record.error_message }}
          </span>
        </template>
        <template v-if="column.key === 'created'">
          <span class="admin-cell-muted">{{ formatDate(record.created_at) }}</span>
        </template>
        <template v-if="column.key === 'actions'">
          <a-space>
            <a-button type="link" size="small" @click="retryDeadLetter(record)">Retry</a-button>
            <a-button type="link" danger size="small" @click="deleteDeadLetter(record)">Delete</a-button>
          </a-space>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import api from '../../lib/api'

const deadLetters = ref([])

const dlColumns = [
  { title: 'Queue', key: 'queue_name' },
  { title: 'Job Type', key: 'job_type' },
  { title: 'Error', key: 'error' },
  { title: 'Created', key: 'created' },
  { title: 'Actions', key: 'actions' },
]

function formatDate(dateStr) {
  return new Date(dateStr).toLocaleString()
}

async function loadDeadLetters() {
  const { data } = await api.get('/admin/dead-letters')
  deadLetters.value = data.dead_letters || []
}

async function retryDeadLetter(dl) {
  Modal.confirm({
    title: `Retry job from ${dl.queue_name}?`,
    onOk: async () => {
      try {
        await api.post(`/admin/dead-letters/${dl.id}`)
        await loadDeadLetters()
      } catch (err) {
        message.error(err.response?.data?.error || 'Failed to retry job')
      }
    },
  })
}

async function deleteDeadLetter(dl) {
  Modal.confirm({
    title: 'Delete this failed job?',
    okText: 'Delete',
    okType: 'danger',
    onOk: async () => {
      try {
        await api.delete(`/admin/dead-letters/${dl.id}`)
        await loadDeadLetters()
      } catch (err) {
        message.error(err.response?.data?.error || 'Failed to delete job')
      }
    },
  })
}

onMounted(loadDeadLetters)
</script>
