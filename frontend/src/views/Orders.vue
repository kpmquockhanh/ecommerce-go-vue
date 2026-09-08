<template>
  <div class="page-container--narrow">
    <h1 class="page-heading--mb">My Orders</h1>

    <a-spin v-if="orderStore.loading" size="large" class="orders-spinner" />

    <a-empty v-else-if="orderStore.orders.length === 0">
      <template #image>
        <ShoppingCartOutlined class="text-4xl text-placeholder" />
      </template>
      <p class="text-secondary">No orders yet</p>
      <router-link to="/products">
        <a-button type="primary">Start shopping</a-button>
      </router-link>
    </a-empty>

    <a-table
      v-else
      :dataSource="orderStore.orders"
      :columns="orderColumns"
      :pagination="false"
      rowKey="id"
      class="table-bg"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'id'">
          <span class="font-medium">#{{ record.id }}</span>
        </template>
        <template v-if="column.key === 'date'">
          <span class="text-secondary">{{ formatDate(record.created_at) }}</span>
        </template>
        <template v-if="column.key === 'status'">
          <a-tag :color="statusClass(record.status)">{{ record.status }}</a-tag>
        </template>
        <template v-if="column.key === 'items_count'">
          <span class="text-secondary">{{ record.items_count }}</span>
        </template>
        <template v-if="column.key === 'total'">
          <span class="font-medium">${{ (record.total / 100).toFixed(2) }}</span>
        </template>
        <template v-if="column.key === 'action'">
          <router-link :to="`/orders/${record.id}`">
            <a-button type="link">View</a-button>
          </router-link>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useOrderStore } from '../stores/orders'
import { statusClass } from '../lib/utils'
import { ShoppingCartOutlined } from '@ant-design/icons-vue'

const orderStore = useOrderStore()

const orderColumns = [
  { title: 'Order', key: 'id' },
  { title: 'Date', key: 'date' },
  { title: 'Status', key: 'status' },
  { title: 'Items', key: 'items_count' },
  { title: 'Total', key: 'total' },
  { title: '', key: 'action' },
]

function formatDate(dateStr) {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(() => {
  orderStore.fetchOrders()
})
</script>
