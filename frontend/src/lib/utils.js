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
    pending: 'warning',
    processing: 'warning',
    shipped: 'green',
    delivered: 'green',
    cancelled: 'red',
    refunded: 'red',
    active: 'green',
    inactive: 'default',
  }
  return map[status] || 'default'
}
