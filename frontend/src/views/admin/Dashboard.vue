<template>
  <div class="admin-page">
    <h1 class="admin-page-header">Admin Dashboard</h1>

    <a-alert
      v-if="error"
      type="error"
      :message="error"
      show-icon
      class="mb-24"
    />

    <div class="admin-stats-grid">
      <a-card v-for="stat in statCards" :key="stat.title" :bordered="true" class="admin-card">
        <a-statistic :title="stat.title" :value="stat.value" :value-style="stat.color ? { color: stat.color } : {}" />
      </a-card>
    </div>

    <a-card :bordered="true" class="admin-card">
      <template #title>
        <h2 class="admin-card-title">Recent Orders</h2>
      </template>
      <a-table
        :dataSource="recentOrders"
        :columns="orderColumns"
        :pagination="false"
        :loading="loading"
        rowKey="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'id'">
            <span class="admin-cell-bold">#{{ record.id }}</span>
          </template>
          <template v-if="column.key === 'status'">
            <a-tag :class="statusClass(record.status)">{{ record.status }}</a-tag>
          </template>
          <template v-if="column.key === 'total'">
            ${{ (record.total / 100).toFixed(2) }}
          </template>
          <template v-if="column.key === 'date'">
            <span class="admin-cell-muted">{{ formatDate(record.created_at) }}</span>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../../lib/api'
import { statusClass } from '../../lib/utils'

const stats = ref({ totalOrders: 0, totalRevenue: 0, totalUsers: 0, pendingOrders: 0 })
const recentOrders = ref([])
const loading = ref(true)
const error = ref(null)

const statCards = computed(() => [
  { title: 'Total Revenue', value: `$${(stats.value.totalRevenue / 100).toFixed(2)}` },
  { title: 'Total Orders', value: stats.value.totalOrders },
  { title: 'Pending Orders', value: stats.value.pendingOrders, color: stats.value.pendingOrders > 0 ? '#faad14' : undefined },
  { title: 'Total Users', value: stats.value.totalUsers },
])

const orderColumns = [
  { title: 'Order', key: 'id' },
  { title: 'Customer', dataIndex: 'customer_name' },
  { title: 'Status', key: 'status' },
  { title: 'Total', key: 'total' },
  { title: 'Date', key: 'date' },
]

function formatDate(dateStr) {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(async () => {
  try {
    const [statsRes, ordersRes] = await Promise.all([
      api.get('/admin/stats'),
      api.get('/admin/orders', { params: { limit: 10 } }),
    ])

    const s = statsRes.data
    stats.value.totalOrders = s.total_orders
    stats.value.totalRevenue = s.total_revenue
    stats.value.totalUsers = s.total_users
    stats.value.pendingOrders = s.pending_orders

    recentOrders.value = ordersRes.data.orders || []
  } catch (e) {
    error.value = 'Failed to load dashboard data. Please try again.'
  } finally {
    loading.value = false
  }
})
</script>
