<template>
  <div class="page-container--cart">
    <h1 class="page-heading--mb">Shopping Cart</h1>

    <div v-if="cart.loading" class="cart-page-grid">
      <div class="cart-items-col">
        <a-card v-for="i in 3" :key="i" :bordered="true">
          <a-skeleton :paragraph="false" active avatar :avatar="{ size: 80, shape: 'square' }">
            <a-skeleton :paragraph="{ rows: 1 }" active />
          </a-skeleton>
        </a-card>
      </div>
      <a-skeleton :paragraph="{ rows: 6 }" active :bordered="true" />
    </div>

    <div v-else-if="cart.items.length === 0" class="empty-state">
      <a-empty description="Your cart is empty">
        <template #image>
          <ShoppingCartOutlined class="text-placeholder icon-xl" />
        </template>
        <p class="empty-state-text">Looks like you haven't added anything yet</p>
        <router-link to="/products">
          <a-button type="primary" size="large">
            <template #icon><ShoppingCartOutlined /></template>
            Start Shopping
          </a-button>
        </router-link>
      </a-empty>
    </div>

    <div v-else class="cart-page-grid">
      <div>
        <a-table
          :dataSource="cart.items"
          :columns="cartColumns"
          :pagination="false"
          rowKey="id"
          class="table-bg"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'product'">
              <div class="flex-center gap-16">
                <div class="cart-item-image">
                  <img
                    v-if="record.product.images?.length"
                    :src="record.product.images[0]"
                    :alt="record.product.name"
                  />
                  <div v-else class="no-image-sm">
                    <span class="image-placeholder-text--xs">No img</span>
                  </div>
                </div>
                <div>
                  <router-link
                    :to="`/products/${record.product.slug}`"
                    class="product-name-link"
                  >
                    {{ record.product.name }}
                  </router-link>
                  <p v-if="record.variant" class="product-variant-text">
                    {{ record.variant.label }}
                  </p>
                </div>
              </div>
            </template>
            <template v-if="column.key === 'price'">
              <span class="text-secondary">{{ formatPrice(record.product.price) }}</span>
            </template>
            <template v-if="column.key === 'quantity'">
              <a-input-number
                :value="record.quantity"
                :min="1"
                @change="(val) => updateQuantity(record, val)"
                size="small"
              />
            </template>
            <template v-if="column.key === 'subtotal'">
              <span class="font-semibold text-primary-color">{{ formatPrice(record.subtotal) }}</span>
            </template>
            <template v-if="column.key === 'action'">
              <a-button type="text" danger size="small" @click="removeItem(record.id)">
                <template #icon><DeleteOutlined /></template>
              </a-button>
            </template>
          </template>
        </a-table>

        <div class="mt-24">
          <router-link to="/products" class="text-primary font-medium">
            <LeftOutlined /> Continue Shopping
          </router-link>
        </div>
      </div>

      <div>
        <a-card :bordered="true" class="order-summary">
          <h2 class="text-lg font-semibold text-primary-color mb-24">Order Summary</h2>

          <div class="order-summary-rows">
            <div class="order-summary-row">
              <span class="text-secondary">Subtotal ({{ cart.itemCount }} items)</span>
              <span class="font-medium text-primary-color">{{ cart.formattedTotal }}</span>
            </div>
            <div class="order-summary-row">
              <span class="text-secondary">Shipping</span>
              <span class="font-medium text-primary">Free</span>
            </div>
          </div>

          <a-divider class="divider-spaced" />

          <div class="order-summary-total">
            <span class="font-semibold text-primary-color">Total</span>
            <span class="order-summary-total-value">{{ cart.formattedTotal }}</span>
          </div>

          <router-link to="/checkout" class="block">
            <a-button type="primary" size="large" block>Proceed to Checkout</a-button>
          </router-link>

          <div class="secure-badge">
            <LockOutlined />
            Secure checkout powered by Stripe
          </div>
        </a-card>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useCartStore } from '../stores/cart'
import { message } from 'ant-design-vue'
import { formatPrice } from '../lib/utils'
import { ShoppingCartOutlined, DeleteOutlined, LeftOutlined, LockOutlined } from '@ant-design/icons-vue'

const cart = useCartStore()

const cartColumns = [
  { title: 'Product', key: 'product' },
  { title: 'Price', key: 'price', align: 'center' },
  { title: 'Quantity', key: 'quantity', align: 'center' },
  { title: 'Total', key: 'subtotal', align: 'right' },
  { title: '', key: 'action', width: 48 },
]

onMounted(() => {
  cart.fetchCart()
})

async function updateQuantity(item, newQuantity) {
  if (newQuantity < 1) return
  await cart.updateItem(item.id, newQuantity)
}

async function removeItem(itemId) {
  await cart.removeItem(itemId)
  message.success('Item removed from cart')
}
</script>
