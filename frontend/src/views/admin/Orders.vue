<template>
  <div class="admin-page">
    <h1 class="admin-page-header">Orders</h1>

    <div class="admin-filter-bar">
      <a-select
        v-model:value="statusFilter"
        @change="loadOrders"
        class="admin-filter-select"
        placeholder="All Status"
      >
        <a-select-option value="">All Status</a-select-option>
        <a-select-option value="pending">Pending</a-select-option>
        <a-select-option value="paid">Paid</a-select-option>
        <a-select-option value="shipped">Shipped</a-select-option>
        <a-select-option value="delivered">Delivered</a-select-option>
        <a-select-option value="cancelled">Cancelled</a-select-option>
      </a-select>
    </div>

    <a-table
      :dataSource="orders"
      :columns="orderColumns"
      :pagination="false"
      rowKey="id"
      class="admin-table"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'id'">
          <span class="admin-cell-bold">#{{ record.id }}</span>
        </template>
        <template v-if="column.key === 'status'">
          <a-select
            :value="record.status"
            @change="(val) => updateStatus(record.id, val)"
            size="small"
            style="width: 120px;"
          >
            <a-select-option value="pending">Pending</a-select-option>
            <a-select-option value="paid">Paid</a-select-option>
            <a-select-option value="shipped">Shipped</a-select-option>
            <a-select-option value="delivered">Delivered</a-select-option>
            <a-select-option value="cancelled">Cancelled</a-select-option>
          </a-select>
        </template>
        <template v-if="column.key === 'total'">
          ${{ (record.total / 100).toFixed(2) }}
        </template>
        <template v-if="column.key === 'date'">
          <span class="admin-cell-muted">{{ formatDate(record.created_at) }}</span>
        </template>
        <template v-if="column.key === 'tag'">
          <a-tag :color="statusClass(record.status)">{{ record.status }}</a-tag>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../../lib/api'
import { statusClass } from '../../lib/utils'

const orders = ref([])
const statusFilter = ref('')

const orderColumns = [
  { title: 'Order', key: 'id' },
  { title: 'Customer', dataIndex: 'customer_name' },
  { title: 'Status', key: 'status' },
  { title: 'Total', key: 'total' },
  { title: 'Date', key: 'date' },
  { title: 'Tag', key: 'tag' },
]

function formatDate(dateStr) {
  return new Date(dateStr).toLocaleDateString()
}

async function loadOrders() {
  const params = { limit: 50 }
  if (statusFilter.value) params.status = statusFilter.value
  const { data } = await api.get('/admin/orders', { params })
  orders.value = data.orders
}

async function updateStatus(orderId, status) {
  await api.put(`/admin/orders/${orderId}`, { status })
  await loadOrders()
}

onMounted(loadOrders)
</script>
