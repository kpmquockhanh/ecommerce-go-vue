<template>
  <div class="page-container--narrow">
    <a-spin v-if="orderStore.loading" size="large" class="orders-spinner" />
    <a-alert v-else-if="orderStore.error" :message="orderStore.error" type="error" show-icon />
    <template v-else-if="order">
      <div class="order-detail-header">
        <h1 class="page-heading">Order #{{ order.id }}</h1>
        <a-tag :class="statusClass(order.status)">{{ order.status }}</a-tag>
      </div>

      <div class="order-detail-cards mb-32">
        <a-card :bordered="true">
          <template #title>
            <h2 class="order-card-title">Shipping Address</h2>
          </template>
          <a-descriptions :column="1">
            <a-descriptions-item label="Name">
              {{ order.shipping_address.first_name }} {{ order.shipping_address.last_name }}
            </a-descriptions-item>
            <a-descriptions-item label="Address">
              {{ order.shipping_address.address_1 }}
            </a-descriptions-item>
            <a-descriptions-item v-if="order.shipping_address.address_2" label="Address 2">
              {{ order.shipping_address.address_2 }}
            </a-descriptions-item>
            <a-descriptions-item label="City/State/ZIP">
              {{ order.shipping_address.city }}, {{ order.shipping_address.state }} {{ order.shipping_address.zip }}
            </a-descriptions-item>
            <a-descriptions-item label="Country">
              {{ order.shipping_address.country }}
            </a-descriptions-item>
          </a-descriptions>
        </a-card>

        <a-card :bordered="true">
          <template #title>
            <h2 class="order-card-title">Payment Info</h2>
          </template>
          <a-descriptions :column="1">
            <a-descriptions-item label="Status">{{ order.payment_status }}</a-descriptions-item>
            <a-descriptions-item label="Method">Stripe</a-descriptions-item>
            <a-descriptions-item label="ID">
              <span class="text-xs text-secondary">{{ order.stripe_payment_intent }}</span>
            </a-descriptions-item>
          </a-descriptions>
        </a-card>
      </div>

      <a-card :bordered="true">
        <template #title>
          <h2 class="order-card-title">Items</h2>
        </template>
        <a-table
          :dataSource="order.items"
          :columns="itemColumns"
          :pagination="false"
          rowKey="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'product'">
              <span class="font-medium">{{ record.product_name }}</span>
            </template>
            <template v-if="column.key === 'variant'">
              <span class="text-secondary">{{ record.variant_label || '-' }}</span>
            </template>
            <template v-if="column.key === 'subtotal'">
              <span class="font-medium">${{ (record.subtotal / 100).toFixed(2) }}</span>
            </template>
          </template>
          <template #summary>
            <a-table-summary fixed>
              <a-table-summary-row>
                <a-table-summary-cell :col-span="4" :index="0" class="text-right">
                  <span class="font-bold">Total: ${{ (order.total / 100).toFixed(2) }}</span>
                </a-table-summary-cell>
              </a-table-summary-row>
            </a-table-summary>
          </template>
        </a-table>
      </a-card>

      <div class="order-back-link">
        <router-link to="/orders" class="text-primary">
          <LeftOutlined /> Back to Orders
        </router-link>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useOrderStore } from '../stores/orders'
import { statusClass } from '../lib/utils'
import { LeftOutlined } from '@ant-design/icons-vue'

const route = useRoute()
const orderStore = useOrderStore()

const order = computed(() => orderStore.currentOrder)

const itemColumns = [
  { title: 'Product', key: 'product' },
  { title: 'Variant', key: 'variant' },
  { title: 'Qty', dataIndex: 'quantity' },
  { title: 'Price', customRender: ({ record }) => `$${(record.price / 100).toFixed(2)}` },
  { title: 'Subtotal', key: 'subtotal' },
]

onMounted(() => {
  orderStore.fetchOrder(route.params.id)
})
</script>
