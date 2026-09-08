<template>
  <div class="page-container--cart">
    <!-- Confirmation Page -->
    <div v-if="orderComplete" class="confirmation-card">
      <div class="confirmation-success-icon">
        <span class="icon">✓</span>
      </div>
      <h1 class="confirmation-title">Order Confirmed!</h1>
      <div class="text-center mb-16">
        <span class="confirmation-order-number">Order #{{ orderId }}</span>
      </div>
      <div class="confirmation-email-notice">
        <MailOutlined />
        A confirmation email has been sent to your email address
      </div>
      <div class="confirmation-details">
        <div class="confirmation-detail-section">
          <div class="confirmation-detail-label">Shipping to</div>
          <div class="confirmation-detail-value">
            {{ form.firstName }} {{ form.lastName }}<br />
            {{ form.address1 }}<br />
            <template v-if="form.address2">{{ form.address2 }}<br /></template>
            {{ form.city }}, {{ form.state }} {{ form.zip }}<br />
            {{ form.country }}
          </div>
        </div>
        <div class="confirmation-detail-section">
          <div class="confirmation-detail-label">Estimated Delivery</div>
          <div class="confirmation-detail-value">5-7 business days</div>
        </div>
      </div>
      <div class="confirmation-actions">
        <router-link :to="`/orders/${orderId}`">
          <a-button type="primary" size="large">View Order Details</a-button>
        </router-link>
        <router-link to="/products">
          <a-button size="large">Continue Shopping</a-button>
        </router-link>
      </div>
      <div class="recommended-section">
        <h2 class="recommended-title">You might also like</h2>
        <div class="recommended-grid">
          <div v-for="item in recommendedProducts" :key="item.id" class="product-card">
            <div class="product-card-image-wrap">
              <img :src="item.image" :alt="item.name" class="product-card-image" />
            </div>
            <div class="product-card-body">
              <div class="product-card-category">{{ item.category }}</div>
              <div class="product-card-name">{{ item.name }}</div>
              <div class="price-display price-display--sm">${{ item.price }}</div>
            </div>
          </div>
        </div>
      </div>
      <div class="trust-bar">
        <div class="trust-item">
          <span class="trust-icon trust-icon--primary">✓</span>
          <span class="trust-label">30-Day Money-Back Guarantee</span>
        </div>
        <div class="trust-item">
          <span class="trust-icon trust-icon--secondary">🔒</span>
          <span class="trust-label">SSL Encrypted</span>
        </div>
        <div class="trust-item">
          <span class="trust-icon trust-icon--secondary">★</span>
          <span class="trust-label">10,000+ Happy Customers</span>
        </div>
        <div class="trust-item">
          <span class="trust-icon trust-icon--secondary">●</span>
          <span class="trust-label">Powered by <span class="stripe-badge-name">Stripe</span></span>
        </div>
      </div>
    </div>

    <!-- Checkout Steps -->
    <div v-else>
      <div class="mb-32">
        <h1 class="page-heading">Secure Checkout</h1>
      </div>

      <!-- Progress Bar -->
      <div class="checkout-progress">
        <div class="progress-step">
          <div
            class="progress-circle"
            :class="{
              'progress-circle--active': currentStep === 1,
              'progress-circle--completed': currentStep > 1,
            }"
          >
            <template v-if="currentStep > 1">✓</template>
            <template v-else>1</template>
          </div>
          <span
            class="progress-label"
            :class="{
              'progress-label--active': currentStep === 1,
              'progress-label--completed': currentStep > 1,
            }"
          >Shipping</span>
        </div>
        <div class="progress-line" :class="{ 'progress-line--completed': currentStep > 1 }"></div>
        <div class="progress-step">
          <div
            class="progress-circle"
            :class="{
              'progress-circle--active': currentStep === 2,
              'progress-circle--completed': currentStep > 2,
            }"
          >
            <template v-if="currentStep > 2">✓</template>
            <template v-else>2</template>
          </div>
          <span
            class="progress-label"
            :class="{
              'progress-label--active': currentStep === 2,
              'progress-label--completed': currentStep > 2,
            }"
          >Payment</span>
        </div>
        <div class="progress-line" :class="{ 'progress-line--completed': currentStep > 2 }"></div>
        <div class="progress-step">
          <div
            class="progress-circle"
            :class="{
              'progress-circle--active': currentStep === 3,
            }"
          >3</div>
          <span
            class="progress-label"
            :class="{ 'progress-label--active': currentStep === 3 }"
          >Review</span>
        </div>
      </div>

      <div class="cart-page-grid">
        <!-- Step 1: Shipping -->
        <div v-if="currentStep === 1">
          <form @submit.prevent="goToPayment">
            <a-card :bordered="true" class="mb-24">
              <template #title>
                <h2 class="text-lg font-semibold text-primary-color">Shipping Address</h2>
              </template>
              <a-form layout="vertical">
                <div class="form-grid-2">
                  <a-form-item label="First Name" required>
                    <a-input v-model:value="form.firstName" />
                  </a-form-item>
                  <a-form-item label="Last Name" required>
                    <a-input v-model:value="form.lastName" />
                  </a-form-item>
                </div>
                <a-form-item label="Address Line 1" required>
                  <a-input v-model:value="form.address1" />
                </a-form-item>
                <a-form-item label="Address Line 2 (optional)">
                  <a-input v-model:value="form.address2" />
                </a-form-item>
                <div class="form-grid-2">
                  <a-form-item label="City" required>
                    <a-input v-model:value="form.city" />
                  </a-form-item>
                  <a-form-item label="State" required>
                    <a-input v-model:value="form.state" :maxlength="2" />
                  </a-form-item>
                </div>
                <div class="form-grid-2">
                  <a-form-item label="ZIP Code" required>
                    <a-input v-model:value="form.zip" />
                  </a-form-item>
                  <a-form-item label="Country" required>
                    <a-input v-model:value="form.country" :maxlength="2" />
                  </a-form-item>
                </div>
              </a-form>
            </a-card>
            <a-button type="primary" html-type="submit" size="large" block>
              Continue to Payment
            </a-button>
          </form>
        </div>

        <!-- Step 2: Payment -->
        <div v-if="currentStep === 2">
          <div class="payment-card">
            <div class="payment-card-header">
              <h2 class="payment-card-title">Payment</h2>
              <span class="stripe-badge">
                Secured by <span class="stripe-badge-name">Stripe</span>
              </span>
            </div>
            <div id="payment-element" class="payment-element">
              <div v-if="processingPayment" class="payment-loading">
                <a-spin size="large" />
                <span class="ml-12 text-secondary">Loading payment form...</span>
              </div>
            </div>
            <div class="payment-inline-trust">
              <span class="payment-trust-item">
                <span class="payment-trust-icon">✓</span> PCI Compliant
              </span>
              <span class="payment-trust-item">
                <span class="payment-trust-icon">🔒</span> SSL Encrypted
              </span>
              <span class="payment-trust-item">
                <span class="payment-trust-icon">✓</span> No data stored
              </span>
            </div>
          </div>

          <a-alert
            v-if="error"
            :message="error"
            type="error"
            show-icon
            class="mb-24"
          />

          <div class="payment-cta-wrapper">
            <a-button
              type="primary"
              size="large"
              block
              :disabled="processingPayment || !paymentReady"
              :loading="processingPayment"
              @click="handleCheckout"
            >
              <LockOutlined v-if="!processingPayment" />
              {{ processingPayment ? 'Processing...' : `Place Order — ${cart.formattedTotal}` }}
            </a-button>
            <p class="payment-lock-text">
              <LockOutlined /> Encrypted &amp; secure payment
            </p>
          </div>

          <a-button
            class="mt-16"
            size="large"
            block
            @click="currentStep = 1"
          >
            Back to Shipping
          </a-button>
        </div>

        <!-- Step 3: Review -->
        <div v-if="currentStep === 3">
          <a-card :bordered="true" class="mb-24">
            <h2 class="text-lg font-semibold text-primary-color mb-16">Review Your Order</h2>

            <!-- Shipping Section -->
            <div class="review-section">
              <div class="review-section-header">
                <span class="review-section-title">Shipping Address</span>
                <button class="review-edit-link" @click="currentStep = 1">Edit</button>
              </div>
              <div class="review-card">
                <p class="font-medium text-primary-color">{{ form.firstName }} {{ form.lastName }}</p>
                <p class="text-secondary">{{ form.address1 }}</p>
                <p v-if="form.address2" class="text-secondary">{{ form.address2 }}</p>
                <p class="text-secondary">{{ form.city }}, {{ form.state }} {{ form.zip }}</p>
                <p class="text-secondary">{{ form.country }}</p>
              </div>
            </div>

            <!-- Payment Section -->
            <div class="review-section">
              <div class="review-section-header">
                <span class="review-section-title">Payment Method</span>
                <button class="review-edit-link" @click="currentStep = 2">Edit</button>
              </div>
              <div class="review-card">
                <p class="text-secondary">
                  <span class="stripe-badge-name">Stripe</span> payment ending in ••••
                </p>
              </div>
            </div>

            <!-- Items Section -->
            <div class="review-section">
              <div class="review-section-header">
                <span class="review-section-title">Order Items</span>
              </div>
              <div class="checkout-items-scroll" style="max-height: unset;">
                <div
                  v-for="item in cart.items"
                  :key="item.id"
                  class="checkout-item"
                  style="border-bottom: 1px solid var(--color-border); padding-bottom: 12px;"
                >
                  <div class="checkout-item-image">
                    <img
                      v-if="item.product.images?.length"
                      :src="item.product.images[0]"
                      :alt="item.product.name"
                    />
                  </div>
                  <div class="checkout-item-info">
                    <p class="checkout-item-name">{{ item.product.name }}</p>
                    <p v-if="item.variant" class="checkout-item-variant">
                      {{ item.variant.label }}
                    </p>
                    <p class="checkout-item-qty">Qty: {{ item.quantity }}</p>
                  </div>
                  <span class="checkout-item-price">
                    {{ formatPrice(item.subtotal) }}
                  </span>
                </div>
              </div>
            </div>
          </a-card>

          <div class="trust-bar">
            <div class="trust-item">
              <span class="trust-icon trust-icon--primary">✓</span>
              <span class="trust-label">30-Day Money-Back Guarantee</span>
            </div>
            <div class="trust-item">
              <span class="trust-icon trust-icon--secondary">🔒</span>
              <span class="trust-label">SSL Encrypted</span>
            </div>
            <div class="trust-item">
              <span class="trust-icon trust-icon--secondary">★</span>
              <span class="trust-label">10,000+ Happy Customers</span>
            </div>
            <div class="trust-item">
              <span class="trust-icon trust-icon--secondary">●</span>
              <span class="trust-label">Powered by <span class="stripe-badge-name">Stripe</span></span>
            </div>
          </div>

          <a-alert
            v-if="error"
            :message="error"
            type="error"
            show-icon
            class="mt-24 mb-24"
          />

          <div class="mt-24">
            <a-button
              type="primary"
              size="large"
              block
              :disabled="processingPayment"
              :loading="processingPayment"
              @click="handleCheckout"
            >
              <LockOutlined v-if="!processingPayment" />
              {{ processingPayment ? 'Processing...' : `Place Order — ${cart.formattedTotal}` }}
            </a-button>
            <p class="payment-lock-text">
              <LockOutlined /> Encrypted &amp; secure payment
            </p>
          </div>

          <a-button
            class="mt-16"
            size="large"
            block
            @click="currentStep = 2"
          >
            Back to Payment
          </a-button>
        </div>

        <!-- Order Summary Sidebar -->
        <div>
          <a-card :bordered="true" class="order-summary">
            <template #title>
              <h2 class="text-lg font-semibold text-primary-color">Order Summary</h2>
            </template>

            <div class="checkout-items-scroll">
              <div
                v-for="item in cart.items"
                :key="item.id"
                class="checkout-item"
              >
                <div class="checkout-item-image">
                  <img
                    v-if="item.product.images?.length"
                    :src="item.product.images[0]"
                    :alt="item.product.name"
                  />
                </div>
                <div class="checkout-item-info">
                  <p class="checkout-item-name">{{ item.product.name }}</p>
                  <p v-if="item.variant" class="checkout-item-variant">
                    {{ item.variant.size }} / {{ item.variant.color }}
                  </p>
                  <p class="checkout-item-qty">Qty: {{ item.quantity }}</p>
                </div>
                <span class="checkout-item-price">
                  {{ formatPrice(item.subtotal) }}
                </span>
              </div>
            </div>

            <a-divider class="divider-spaced" />

            <div class="order-summary-rows">
              <div class="order-summary-row">
                <span class="text-secondary">Subtotal</span>
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

            <div class="trust-bar" style="margin-top: 16px;">
              <div class="trust-item">
                <span class="trust-icon trust-icon--primary">✓</span>
                <span class="trust-label">Money-back guarantee</span>
              </div>
              <div class="trust-item">
                <span class="trust-icon trust-icon--secondary">🔒</span>
                <span class="trust-label">SSL Secure</span>
              </div>
              <div class="trust-item">
                <span class="trust-icon trust-icon--secondary">★</span>
                <span class="trust-label">47 customers today</span>
              </div>
            </div>
          </a-card>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { message } from 'ant-design-vue'
import { formatPrice } from '../lib/utils'
import { LockOutlined, MailOutlined } from '@ant-design/icons-vue'
import api from '../lib/api'

const router = useRouter()
const cart = useCartStore()

const form = reactive({
  firstName: '',
  lastName: '',
  address1: '',
  address2: '',
  city: '',
  state: '',
  zip: '',
  country: 'US',
})

const error = ref(null)
const processingPayment = ref(false)
const paymentReady = ref(false)
const orderComplete = ref(false)
const orderId = ref(null)
const checkoutInitialized = ref(false)
const currentStep = ref(1)
let stripe = null
let elements = null

const recommendedProducts = [
  { id: 1, name: 'Organic Cotton Tee', category: 'TOPS', price: '45.00', image: 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=300&h=300&fit=crop' },
  { id: 2, name: 'Recycled Denim Jeans', category: 'BOTTOMS', price: '89.00', image: 'https://images.unsplash.com/photo-1542272604-787c3835535d?w=300&h=300&fit=crop' },
  { id: 3, name: 'Hemp Canvas Sneakers', category: 'FOOTWEAR', price: '65.00', image: 'https://images.unsplash.com/photo-1525966222134-fcfa99b8ae77?w=300&h=300&fit=crop' },
]

function getOrCreateIdempotencyKey() {
  const storageKey = 'checkout_idempotency_key'
  let key = sessionStorage.getItem(storageKey)
  if (!key) {
    key = crypto.randomUUID()
    sessionStorage.setItem(storageKey, key)
  }
  return key
}
const idempotencyKey = getOrCreateIdempotencyKey()

let stripeScript = null

onMounted(async () => {
  await cart.fetchCart()

  try {
    const { data: config } = await api.get('/config')
    stripeScript = document.createElement('script')
    stripeScript.src = 'https://js.stripe.com/v3/'
    stripeScript.onload = () => {
      stripe = window.Stripe(config.stripe_publishable_key)
      // Initialize payment when user reaches step 2
      if (currentStep.value === 2) {
        initializePayment()
      }
    }
    stripeScript.onerror = () => {
      error.value = 'Failed to load payment system. Please refresh the page.'
    }
    document.head.appendChild(stripeScript)
  } catch (err) {
    error.value = 'Failed to load checkout configuration. Please refresh the page.'
  }
})

onUnmounted(() => {
  if (stripeScript && stripeScript.parentNode) {
    stripeScript.parentNode.removeChild(stripeScript)
  }
})

// Initialize payment when navigating to step 2
watch(currentStep, (newStep) => {
  if (newStep === 2 && stripe && !checkoutInitialized.value) {
    initializePayment()
  }
})

async function initializePayment() {
  if (!stripe || checkoutInitialized.value) return
  
  if (cart.total === 0) {
    error.value = 'Your cart is empty. Please add items before checkout.'
    return
  }
  
  checkoutInitialized.value = true

  try {
    const { data } = await api.post('/checkout/init', {
      idempotency_key: idempotencyKey,
    })

    elements = stripe.elements({
      clientSecret: data.client_secret,
    })

    const paymentElement = elements.create('payment')
    paymentElement.mount('#payment-element')
    paymentReady.value = true
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to initialize payment'
    checkoutInitialized.value = false
  }
}

function goToPayment() {
  if (!form.firstName || !form.lastName || !form.address1 || !form.city || !form.state || !form.zip || !form.country) {
    message.warning('Please fill in all required fields')
    return
  }
  currentStep.value = 2
}

async function handleCheckout() {
  if (!stripe || !elements) return

  processingPayment.value = true
  error.value = null

  try {
    const { data: confirmData } = await api.post('/checkout/confirm', {
      shipping_address: {
        first_name: form.firstName,
        last_name: form.lastName,
        address_1: form.address1,
        address_2: form.address2,
        city: form.city,
        state: form.state,
        zip: form.zip,
        country: form.country,
      },
      idempotency_key: idempotencyKey,
    })

    orderId.value = confirmData.order_id

    if (!orderId.value) {
      error.value = 'Order confirmation failed. Please try again.'
      processingPayment.value = false
      return
    }

    const { error: stripeError } = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: `${window.location.origin}/orders/${orderId.value}`,
      },
    })

    if (stripeError) {
      error.value = stripeError.message
    } else {
      orderComplete.value = true
      message.success('Order placed successfully!')
      await cart.fetchCart()
    }
  } catch (err) {
    error.value = err.response?.data?.error || 'Payment failed. Please try again.'
  } finally {
    processingPayment.value = false
  }
}
</script>
