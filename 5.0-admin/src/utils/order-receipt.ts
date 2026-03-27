export function buildOrderReceiptUrl(orderId: string | number, options: {
  autoPrint?: boolean
  autoDownload?: boolean
  template?: 'a4' | 'dot' | 'receipt'
} = {}) {
  const params = new URLSearchParams()
  params.set('orderId', String(orderId || ''))
  if (options.autoPrint)
    params.set('autoPrint', '1')
  if (options.autoDownload)
    params.set('autoDownload', '1')
  if (options.template)
    params.set('template', options.template)
  return `${window.location.origin}${window.location.pathname}#/print/order-receipt?${params.toString()}`
}

export function openOrderReceiptPage(orderId: string | number, options: {
  autoPrint?: boolean
  autoDownload?: boolean
  template?: 'a4' | 'dot' | 'receipt'
} = {}) {
  if (!orderId)
    return
  window.open(buildOrderReceiptUrl(orderId, options), '_blank')
}
