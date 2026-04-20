import dayjs from 'dayjs'
import { downloadOrderReceiptPdfApi } from '@/api/finance-center/order-manage'

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

function triggerBlobDownload(response: any) {
  const blob = new Blob([response.data], { type: response.headers['content-type'] || 'application/pdf' })
  const disposition = response.headers['content-disposition'] || ''
  const matched = disposition.match(/filename\*=UTF-8''([^;]+)/i)
  const fileName = matched ? decodeURIComponent(matched[1]) : `订单收据-${dayjs().format('YYYYMMDDHHmmss')}.pdf`
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

export async function downloadOrderReceiptPdf(orderId: string | number, options: {
  template?: 'a4' | 'dot' | 'receipt'
} = {}) {
  if (!orderId)
    return
  const response = await downloadOrderReceiptPdfApi({
    orderId,
    template: options.template,
  })
  triggerBlobDownload(response)
}
