<template>
  <div class="admin-page" v-if="product">
    <div class="admin-page-header--row">
      <div>
        <a-button type="link" @click="$router.push('/admin/products')" style="padding-left: 0">
          ← Back to Products
        </a-button>
        <h1 class="page-heading">{{ product.name }}</h1>
      </div>
      <a-tag :class="product.status === 'published' ? 'tag-success' : 'tag-neutral'">
        {{ product.status }}
      </a-tag>
    </div>

    <a-tabs v-model:activeKey="activeTab">
      <!-- OVERVIEW TAB -->
      <a-tab-pane key="overview" tab="Overview">
        <a-form layout="vertical" :model="productForm" @finish="saveProduct" style="max-width: 700px">
          <a-form-item label="Name">
            <a-input v-model:value="productForm.name" />
          </a-form-item>
          <a-form-item label="Description">
            <a-textarea v-model:value="productForm.description" :rows="4" />
          </a-form-item>
          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 16px">
            <a-form-item label="Price (cents)">
              <a-input-number v-model:value="productForm.price" :min="0" style="width: 100%" />
            </a-form-item>
            <a-form-item label="Compare at Price (cents)">
              <a-input-number v-model:value="productForm.compare_at_price" :min="0" style="width: 100%" />
            </a-form-item>
          </div>
          <a-form-item label="Status">
            <a-select v-model:value="productForm.status">
              <a-select-option value="published">Published</a-select-option>
              <a-select-option value="draft">Draft</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="Categories">
            <a-select
              v-model:value="productForm.category_ids"
              mode="multiple"
              placeholder="Select categories"
              :options="categories.map(cat => ({ label: cat.name, value: cat.id }))"
            />
          </a-form-item>
          <a-form-item>
            <a-button type="primary" html-type="submit" :loading="saving">Save Changes</a-button>
          </a-form-item>
        </a-form>
      </a-tab-pane>

      <!-- IMAGES TAB -->
      <a-tab-pane key="images" tab="Images">
        <div style="margin-bottom: 16px">
          <a-upload
            :before-upload="handleImageUpload"
            :show-upload-list="false"
            accept="image/jpeg,image/png,image/webp,image/gif"
          >
            <a-button type="primary" :loading="uploadingImage">
              <template #icon><UploadOutlined /></template>
              Upload Image
            </a-button>
          </a-upload>
          <p class="admin-hint">Max 10MB. JPG, PNG, WebP, or GIF.</p>
        </div>

        <div v-if="images.length === 0" style="color: #999; padding: 24px 0">
          No images yet. Upload one to display it on the storefront.
        </div>

        <div v-else class="admin-image-grid">
          <a-card v-for="(img, idx) in images" :key="img.path" size="small" class="admin-image-card">
            <img :src="img.url" class="admin-image-card-thumb" />
            <div style="margin-top: 8px">
              <a-tag v-if="img.isPrimary" class="tag-success">Primary</a-tag>
            </div>
            <a-space style="margin-top: 8px">
              <a-button size="small" :disabled="idx === 0" @click="moveImage(idx, -1)">↑</a-button>
              <a-button size="small" :disabled="idx === images.length - 1" @click="moveImage(idx, 1)">↓</a-button>
              <a-button size="small" :disabled="img.isPrimary" @click="setPrimaryImage(idx)">Set Primary</a-button>
              <a-button size="small" danger @click="deleteImage(img)">Delete</a-button>
            </a-space>
          </a-card>
        </div>
      </a-tab-pane>

      <!-- OPTION GROUPS TAB -->
      <a-tab-pane key="options" tab="Option Groups">
        <div style="margin-bottom: 16px">
          <a-button type="primary" @click="showGroupModal = true">
            <template #icon><PlusOutlined /></template>
            Add Option Group
          </a-button>
        </div>

        <div v-if="optionGroups.length === 0" style="color: #999; padding: 24px 0">
          No option groups yet. Add one to enable variant selection (e.g., Size, Color).
        </div>

        <a-card v-for="group in optionGroups" :key="group.id" style="margin-bottom: 12px">
          <template #title>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>{{ group.name }} <a-tag class="tag-neutral">{{ group.values?.length || 0 }} values</a-tag></span>
              <a-space>
                <a-button size="small" @click="editGroup(group)">Edit</a-button>
                <a-button size="small" danger @click="deleteGroup(group)">Delete</a-button>
              </a-space>
            </div>
          </template>

          <div style="margin-bottom: 12px">
            <a-button size="small" @click="showValueModal(group)">
              <template #icon><PlusOutlined /></template>
              Add Value
            </a-button>
          </div>

          <a-table
            :dataSource="group.values || []"
            :columns="valueColumns"
            :pagination="false"
            rowKey="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'value'">
                {{ record.value }}
              </template>
              <template v-if="column.key === 'price_modifier'">
                {{ record.price_modifier > 0 ? '+' + formatPrice(record.price_modifier) : record.price_modifier < 0 ? formatPrice(record.price_modifier) : '—' }}
              </template>
              <template v-if="column.key === 'actions'">
                <a-space>
                  <a-button type="link" size="small" @click="editValue(group, record)">Edit</a-button>
                  <a-button type="link" danger size="small" @click="deleteValue(group, record)">Delete</a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </a-card>

        <!-- Option Group Modal -->
        <a-modal
          v-model:open="showGroupModal"
          :title="editingGroup ? 'Edit Option Group' : 'Add Option Group'"
          @ok="saveGroup"
          @cancel="closeGroupModal"
        >
          <a-form layout="vertical">
            <a-form-item label="Group Name" required>
              <a-input v-model:value="groupForm.name" placeholder="e.g., Size, Color, Material" />
            </a-form-item>
          </a-form>
        </a-modal>

        <!-- Option Value Modal -->
        <a-modal
          v-model:open="showValueModalVisible"
          :title="editingValue ? 'Edit Option Value' : 'Add Option Value'"
          @ok="saveValue"
          @cancel="closeValueModal"
        >
          <a-form layout="vertical">
            <a-form-item label="Value" required>
              <a-input v-model:value="valueForm.value" placeholder="e.g., Small, Red, Cotton" />
            </a-form-item>
            <a-form-item label="Price Adjustment (cents)">
              <a-input-number v-model:value="valueForm.price_modifier" style="width: 100%" />
              <div style="color: #999; font-size: 12px">Additional cost when this value is selected. 0 = no adjustment.</div>
            </a-form-item>
          </a-form>
        </a-modal>
      </a-tab-pane>

      <!-- VARIANTS TAB -->
      <a-tab-pane key="variants" tab="Variants">
        <div style="margin-bottom: 16px; display: flex; align-items: center; gap: 12px; flex-wrap: wrap">
          <a-button type="primary" :disabled="bulkEditMode" @click="openVariantModal()">
            <template #icon><PlusOutlined /></template>
            Add Variant
          </a-button>
          <a-button
            :loading="generating"
            :disabled="bulkEditMode || optionGroups.length === 0"
            @click="generateVariants"
          >
            Generate Variants
          </a-button>
          <template v-if="!bulkEditMode">
            <a-button :disabled="variants.length === 0" @click="startBulkEdit">Bulk Edit</a-button>
          </template>
          <template v-else>
            <a-button type="primary" :loading="savingBulkEdit" @click="saveBulkEdit">
              Save Changes ({{ dirtyBulkEditCount }})
            </a-button>
            <a-button :disabled="savingBulkEdit" @click="cancelBulkEdit">Cancel</a-button>
          </template>
          <span v-if="optionGroups.length === 0" style="color: #faad14">
            Add option groups first before creating variants.
          </span>
        </div>

        <a-table
          :dataSource="variants"
          :columns="variantColumns"
          :pagination="false"
          rowKey="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'label'">
              <span class="admin-cell-bold">{{ record.label || '(unlabeled)' }}</span>
            </template>
            <template v-if="column.key === 'sku'">
              <a-input
                v-if="bulkEditMode"
                v-model:value="bulkEditRows[record.id].sku"
                size="small"
                style="width: 160px"
              />
              <a-tag v-else class="tag-neutral">{{ record.sku || '—' }}</a-tag>
            </template>
            <template v-if="column.key === 'stock'">
              <a-input-number
                v-if="bulkEditMode"
                v-model:value="bulkEditRows[record.id].stock"
                :min="0"
                size="small"
                style="width: 100px"
              />
              <a-tag v-else :class="record.stock > 0 ? 'tag-success' : 'tag-error'">{{ record.stock }}</a-tag>
            </template>
            <template v-if="column.key === 'option_values'">
              <a-tag v-for="ov in (record.option_values || [])" :key="ov.option_value_id" class="tag-neutral" style="margin-bottom: 2px">
                {{ ov.option_group_name }}: {{ ov.value }}
              </a-tag>
            </template>
            <template v-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" :disabled="bulkEditMode" @click="openVariantModal(record)">Edit</a-button>
                <a-button type="link" danger size="small" :disabled="bulkEditMode" @click="deleteVariant(record)">Delete</a-button>
              </a-space>
            </template>
          </template>
        </a-table>

        <!-- Variant Modal -->
        <a-modal
          v-model:open="showVariantModal"
          :title="editingVariant ? 'Edit Variant' : 'Add Variant'"
          width="600px"
          @ok="saveVariant"
          @cancel="closeVariantModal"
        >
          <a-form layout="vertical">
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 16px">
              <a-form-item label="SKU">
                <a-input v-model:value="variantForm.sku" placeholder="e.g., TSHIRT-S-BLK" />
              </a-form-item>
              <a-form-item label="Stock" required>
                <a-input-number v-model:value="variantForm.stock" :min="0" style="width: 100%" />
              </a-form-item>
            </div>

            <template v-for="group in optionGroups" :key="group.id">
              <a-form-item :label="group.name" required>
                <a-select
                  v-model:value="variantForm.option_value_ids[group.id]"
                  :placeholder="`Select ${group.name}`"
                >
                  <a-select-option v-for="val in (group.values || [])" :key="val.id" :value="val.id">
                    {{ val.value }}{{ val.price_modifier ? ` (+${formatPrice(val.price_modifier)})` : '' }}
                  </a-select-option>
                </a-select>
              </a-form-item>
            </template>

            <div v-if="optionGroups.length === 0" style="color: #faad14; padding: 12px 0">
              No option groups defined. Add option groups in the "Option Groups" tab first, or create this variant without options.
            </div>
          </a-form>
        </a-modal>
      </a-tab-pane>

      <!-- CUSTOMIZATIONS TAB -->
      <a-tab-pane key="customizations" tab="Customizations">
        <div style="margin-bottom: 16px">
          <a-button type="primary" @click="showCustomizationModal = true">
            <template #icon><PlusOutlined /></template>
            Add Customization Field
          </a-button>
        </div>

        <div v-if="customizations.length === 0" style="color: #999; padding: 24px 0">
          No customization fields. Add fields like "Engraving text" or "Gift wrap" for customers to personalize their order.
        </div>

        <a-table
          :dataSource="customizations"
          :columns="customizationColumns"
          :pagination="false"
          rowKey="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <span class="admin-cell-bold">{{ record.name }}</span>
            </template>
            <template v-if="column.key === 'field_type'">
              <a-tag class="tag-neutral">{{ record.field_type }}</a-tag>
            </template>
            <template v-if="column.key === 'required'">
              <a-tag :class="record.required ? 'tag-warning' : 'tag-neutral'">{{ record.required ? 'Required' : 'Optional' }}</a-tag>
            </template>
            <template v-if="column.key === 'default_value'">
              {{ record.default_value || '—' }}
            </template>
            <template v-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" @click="editCustomization(record)">Edit</a-button>
                <a-button type="link" danger size="small" @click="deleteCustomization(record)">Delete</a-button>
              </a-space>
            </template>
          </template>
        </a-table>

        <!-- Customization Field Modal -->
        <a-modal
          v-model:open="showCustomizationModal"
          :title="editingCustomization ? 'Edit Customization Field' : 'Add Customization Field'"
          @ok="saveCustomization"
          @cancel="closeCustomizationModal"
        >
          <a-form layout="vertical">
            <a-form-item label="Field Name" required>
              <a-input v-model:value="customizationForm.name" placeholder="e.g., Engraving Text, Gift Wrap" />
            </a-form-item>
            <a-form-item label="Field Type" required>
              <a-select v-model:value="customizationForm.field_type">
                <a-select-option value="text">Text Input</a-select-option>
                <a-select-option value="textarea">Text Area</a-select-option>
                <a-select-option value="file">File Upload</a-select-option>
              </a-select>
            </a-form-item>
            <a-form-item label="Required">
              <a-switch v-model:checked="customizationForm.required" />
            </a-form-item>
            <a-form-item label="Default Value">
              <a-input v-model:value="customizationForm.default_value" />
            </a-form-item>
          </a-form>
        </a-modal>
      </a-tab-pane>
    </a-tabs>
  </div>

  <div v-else-if="loading" style="padding: 48px; text-align: center">Loading...</div>
  <div v-else style="padding: 48px; text-align: center; color: #999">Product not found.</div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined, UploadOutlined } from '@ant-design/icons-vue'
import api from '../../lib/api'

const route = useRoute()
const productId = computed(() => Number(route.params.id))

const product = ref(null)
const loading = ref(true)
const saving = ref(false)
const activeTab = ref('overview')

const optionGroups = ref([])
const variants = ref([])
const customizations = ref([])
const images = ref([])
const categories = ref([])
const uploadingImage = ref(false)

const productForm = ref({ name: '', description: '', price: 0, compare_at_price: null, status: 'published', category_ids: [] })

const MAX_IMAGE_SIZE = 10 * 1024 * 1024
const ALLOWED_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']

// Option Group state
const showGroupModal = ref(false)
const editingGroup = ref(null)
const groupForm = ref({ name: '' })

// Option Value state
const showValueModalVisible = ref(false)
const editingValue = ref(null)
const activeGroup = ref(null)
const valueForm = ref({ value: '', price_modifier: 0 })

// Variant state
const showVariantModal = ref(false)
const editingVariant = ref(null)
const variantForm = ref({ sku: '', stock: 0, option_value_ids: {} })
const generating = ref(false)
const bulkEditMode = ref(false)
const bulkEditRows = ref({})
const savingBulkEdit = ref(false)

// Customization state
const showCustomizationModal = ref(false)
const editingCustomization = ref(null)
const customizationForm = ref({ name: '', field_type: 'text', required: false, default_value: '' })

const valueColumns = [
  { title: 'Value', key: 'value' },
  { title: 'Price Modifier', key: 'price_modifier' },
  { title: 'Actions', key: 'actions' },
]

const variantColumns = [
  { title: 'Label', key: 'label' },
  { title: 'SKU', key: 'sku' },
  { title: 'Stock', key: 'stock' },
  { title: 'Options', key: 'option_values' },
  { title: 'Actions', key: 'actions' },
]

const customizationColumns = [
  { title: 'Name', key: 'name' },
  { title: 'Type', key: 'field_type' },
  { title: 'Required', key: 'required' },
  { title: 'Default', key: 'default_value' },
  { title: 'Actions', key: 'actions' },
]

function formatPrice(cents) {
  return '$' + (Math.abs(cents) / 100).toFixed(2)
}

// --- Load data ---
async function loadProduct() {
  try {
    const { data } = await api.get(`/admin/products/${productId.value}`)
    product.value = data
    productForm.value = {
      name: data.name,
      description: data.description,
      price: data.price,
      compare_at_price: data.compare_at_price,
      status: data.status,
      category_ids: (data.categories || []).map((cat) => cat.id),
    }
    optionGroups.value = data.option_groups || []
    variants.value = data.variants || []
    customizations.value = data.customizations || []
    await loadImages()
  } catch {
    product.value = null
  } finally {
    loading.value = false
  }
}

async function loadCategories() {
  try {
    const { data } = await api.get('/admin/categories')
    categories.value = data.categories || []
  } catch {
    categories.value = []
  }
}

// --- Images ---
async function loadImages() {
  const rawPaths = product.value?.images || []
  if (rawPaths.length === 0) {
    images.value = []
    return
  }
  let urls = rawPaths
  try {
    const { data } = await api.get(`/images/product/${productId.value}`)
    if (data.images?.length === rawPaths.length) {
      urls = data.images
    }
  } catch {
    // fall back to raw paths if presigned URLs can't be fetched
  }
  images.value = rawPaths.map((path, idx) => ({
    path,
    url: urls[idx] || path,
    isPrimary: path === product.value.primary_image,
  }))
}

function handleImageUpload(file) {
  if (!ALLOWED_IMAGE_TYPES.includes(file.type)) {
    message.error('Only JPG, PNG, WebP, or GIF images are allowed.')
    return false
  }
  if (file.size > MAX_IMAGE_SIZE) {
    message.error('Image must be less than 10MB.')
    return false
  }

  uploadingImage.value = true
  const formData = new FormData()
  formData.append('image', file)
  api.post(`/admin/products/${productId.value}/images`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
    .then(async () => {
      message.success('Image uploaded')
      await loadProduct()
    })
    .catch((err) => {
      message.error(err.response?.data?.error || 'Failed to upload image')
    })
    .finally(() => {
      uploadingImage.value = false
    })

  return false
}

function moveImage(idx, direction) {
  const target = idx + direction
  if (target < 0 || target >= images.value.length) return
  const arr = [...images.value]
  ;[arr[idx], arr[target]] = [arr[target], arr[idx]]
  images.value = arr
  persistImageOrder()
}

function setPrimaryImage(idx) {
  images.value = images.value.map((img, i) => ({ ...img, isPrimary: i === idx }))
  persistImageOrder()
}

async function persistImageOrder() {
  try {
    const payload = {
      images: images.value.map((img, idx) => ({
        image_path: img.path,
        sort_order: idx,
        is_primary: img.isPrimary,
      })),
    }
    await api.put(`/admin/products/${productId.value}/images/order`, payload)
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to update image order')
  }
}

function deleteImage(img) {
  Modal.confirm({
    title: 'Delete this image?',
    content: 'This will remove it from the product permanently.',
    okText: 'Delete',
    okType: 'danger',
    onOk: async () => {
      try {
        await api.delete('/images/delete', { data: { path: img.path, product_id: productId.value } })
        message.success('Image deleted')
        await loadProduct()
      } catch (err) {
        message.error(err.response?.data?.error || 'Failed to delete image')
      }
    },
  })
}

// --- Product save ---
async function saveProduct() {
  saving.value = true
  try {
    await api.put(`/admin/products/${productId.value}`, productForm.value)
    message.success('Product updated')
    await loadProduct()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to save')
  } finally {
    saving.value = false
  }
}

// --- Option Groups ---
function editGroup(group) {
  editingGroup.value = group
  groupForm.value = { name: group.name }
  showGroupModal.value = true
}

function closeGroupModal() {
  showGroupModal.value = false
  editingGroup.value = null
  groupForm.value = { name: '' }
}

async function saveGroup() {
  try {
    if (editingGroup.value) {
      await api.put(`/admin/products/${productId.value}/option-groups/${editingGroup.value.id}`, groupForm.value)
    } else {
      await api.post(`/admin/products/${productId.value}/option-groups`, groupForm.value)
    }
    closeGroupModal()
    await loadProduct()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to save option group')
  }
}

async function deleteGroup(group) {
  Modal.confirm({
    title: `Delete option group "${group.name}"?`,
    content: 'This will also delete all its values.',
    okText: 'Delete',
    okType: 'danger',
    onOk: async () => {
      try {
        await api.delete(`/admin/products/${productId.value}/option-groups/${group.id}`)
        await loadProduct()
      } catch (err) {
        message.error(err.response?.data?.error || 'Failed to delete')
      }
    },
  })
}

// --- Option Values ---
function showValueModal(group) {
  activeGroup.value = group
  editingValue.value = null
  valueForm.value = { value: '', price_modifier: 0 }
  showValueModalVisible.value = true
}

function editValue(group, val) {
  activeGroup.value = group
  editingValue.value = val
  valueForm.value = { value: val.value, price_modifier: val.price_modifier || 0 }
  showValueModalVisible.value = true
}

function closeValueModal() {
  showValueModalVisible.value = false
  editingValue.value = null
  activeGroup.value = null
  valueForm.value = { value: '', price_modifier: 0 }
}

async function saveValue() {
  try {
    const groupId = activeGroup.value.id
    if (editingValue.value) {
      await api.put(`/admin/products/${productId.value}/option-groups/${groupId}/values/${editingValue.value.id}`, valueForm.value)
    } else {
      await api.post(`/admin/products/${productId.value}/option-groups/${groupId}/values`, valueForm.value)
    }
    closeValueModal()
    await loadProduct()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to save option value')
  }
}

async function deleteValue(group, val) {
  Modal.confirm({
    title: `Delete option value "${val.value}"?`,
    okText: 'Delete',
    okType: 'danger',
    onOk: async () => {
      try {
        await api.delete(`/admin/products/${productId.value}/option-groups/${group.id}/values/${val.id}`)
        await loadProduct()
      } catch (err) {
        message.error(err.response?.data?.error || 'Failed to delete')
      }
    },
  })
}

// --- Variants ---
function openVariantModal(variant = null) {
  editingVariant.value = variant
  if (variant) {
    const oidMap = {}
    for (const ov of (variant.option_values || [])) {
      oidMap[ov.option_group_id] = ov.option_value_id
    }
    variantForm.value = { sku: variant.sku || '', stock: variant.stock, option_value_ids: oidMap }
  } else {
    variantForm.value = { sku: '', stock: 0, option_value_ids: {} }
  }
  showVariantModal.value = true
}

function closeVariantModal() {
  showVariantModal.value = false
  editingVariant.value = null
  variantForm.value = { sku: '', stock: 0, option_value_ids: {} }
}

async function saveVariant() {
  const optionValues = Object.entries(variantForm.value.option_value_ids)
    .filter(([, v]) => v != null)
    .map(([, v]) => ({ option_value_id: v }))

  const payload = {
    sku: variantForm.value.sku,
    stock: variantForm.value.stock,
    option_values: optionValues,
  }

  try {
    if (editingVariant.value) {
      await api.put(`/admin/products/${productId.value}/variants/${editingVariant.value.id}`, payload)
    } else {
      await api.post(`/admin/products/${productId.value}/variants`, payload)
    }
    closeVariantModal()
    await loadProduct()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to save variant')
  }
}

async function deleteVariant(variant) {
  Modal.confirm({
    title: `Delete variant "${variant.label || variant.sku || 'this variant'}"?`,
    okText: 'Delete',
    okType: 'danger',
    onOk: async () => {
      try {
        await api.delete(`/admin/products/${productId.value}/variants/${variant.id}`)
        await loadProduct()
      } catch (err) {
        message.error(err.response?.data?.error || 'Failed to delete')
      }
    },
  })
}

async function generateVariants() {
  generating.value = true
  try {
    const { data } = await api.post(`/admin/products/${productId.value}/variants/generate`)
    if (data.created_count > 0) {
      const skippedNote = data.skipped_count > 0 ? ` (${data.skipped_count} already existed)` : ''
      message.success(`Generated ${data.created_count} variant${data.created_count === 1 ? '' : 's'}${skippedNote}`)
    } else {
      message.info('No new variants to generate — every combination already exists')
    }
    await loadProduct()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to generate variants')
  } finally {
    generating.value = false
  }
}

function startBulkEdit() {
  const rows = {}
  for (const v of variants.value) {
    rows[v.id] = { sku: v.sku || '', stock: v.stock }
  }
  bulkEditRows.value = rows
  bulkEditMode.value = true
}

function cancelBulkEdit() {
  bulkEditMode.value = false
  bulkEditRows.value = {}
}

function changedBulkEditItems() {
  return variants.value
    .filter((v) => {
      const edit = bulkEditRows.value[v.id]
      return edit && (edit.stock !== v.stock || (edit.sku || '') !== (v.sku || ''))
    })
    .map((v) => ({
      variant_id: v.id,
      stock: bulkEditRows.value[v.id].stock,
      sku: bulkEditRows.value[v.id].sku,
    }))
}

const dirtyBulkEditCount = computed(() => changedBulkEditItems().length)

async function saveBulkEdit() {
  const items = changedBulkEditItems()
  if (items.length === 0) {
    message.info('No changes to save')
    return
  }

  savingBulkEdit.value = true
  try {
    const { data } = await api.patch(`/admin/products/${productId.value}/variants/bulk`, { items })
    if (data.failed_count > 0) {
      const failures = (data.results || []).filter((r) => !r.success)
      message.warning(
        `${data.succeeded_count} updated, ${data.failed_count} failed: ` +
        failures.map((f) => `#${f.variant_id} (${f.error})`).join(', ')
      )
    } else {
      message.success(`Updated ${data.succeeded_count} variant${data.succeeded_count === 1 ? '' : 's'}`)
    }
    bulkEditMode.value = false
    bulkEditRows.value = {}
    await loadProduct()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to save changes')
  } finally {
    savingBulkEdit.value = false
  }
}

// --- Customizations ---
function editCustomization(field) {
  editingCustomization.value = field
  customizationForm.value = {
    name: field.name,
    field_type: field.field_type,
    required: field.required,
    default_value: field.default_value || '',
  }
  showCustomizationModal.value = true
}

function closeCustomizationModal() {
  showCustomizationModal.value = false
  editingCustomization.value = null
  customizationForm.value = { name: '', field_type: 'text', required: false, default_value: '' }
}

async function saveCustomization() {
  try {
    if (editingCustomization.value) {
      await api.put(`/admin/products/${productId.value}/customizations/${editingCustomization.value.id}`, customizationForm.value)
    } else {
      await api.post(`/admin/products/${productId.value}/customizations`, customizationForm.value)
    }
    closeCustomizationModal()
    await loadProduct()
  } catch (err) {
    message.error(err.response?.data?.error || 'Failed to save customization field')
  }
}

async function deleteCustomization(field) {
  Modal.confirm({
    title: `Delete customization field "${field.name}"?`,
    okText: 'Delete',
    okType: 'danger',
    onOk: async () => {
      try {
        await api.delete(`/admin/products/${productId.value}/customizations/${field.id}`)
        await loadProduct()
      } catch (err) {
        message.error(err.response?.data?.error || 'Failed to delete')
      }
    },
  })
}

onMounted(() => {
  loadProduct()
  loadCategories()
})
</script>
