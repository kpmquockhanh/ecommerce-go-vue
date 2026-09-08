export function formatPrice(cents) {
  return `$${(cents / 100).toFixed(2)}`
}

export function formatDate(dateString) {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function statusClass(status) {
  const map = {
    pending: 'tag-warning',
    processing: 'tag-warning',
    shipped: 'tag-success',
    delivered: 'tag-success',
    cancelled: 'tag-error',
    refunded: 'tag-error',
    active: 'tag-success',
    inactive: 'tag-neutral',
  }
  return map[status] || 'tag-neutral'
}
