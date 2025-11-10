import type { Patient } from '@/stores/types'

const locale = 'en-US'

export const formatDate = (value?: string, options: Intl.DateTimeFormatOptions = {}) => {
  if (!value) return '—'
  return new Date(value).toLocaleDateString(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    ...options,
  })
}

export const formatDateTime = (value?: string, options: Intl.DateTimeFormatOptions = {}) => {
  if (!value) return '—'
  return new Date(value).toLocaleString(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    ...options,
  })
}

export const formatRelativeTime = (value?: string) => {
  if (!value) return 'N/A'
  const formatter = new Intl.RelativeTimeFormat(locale, { style: 'short' })
  const now = Date.now()
  const then = new Date(value).getTime()
  const diffMinutes = Math.round((then - now) / 60000)
  if (Math.abs(diffMinutes) < 60) {
    return formatter.format(diffMinutes, 'minutes')
  }
  const diffHours = Math.round(diffMinutes / 60)
  if (Math.abs(diffHours) < 24) {
    return formatter.format(diffHours, 'hours')
  }
  const diffDays = Math.round(diffHours / 24)
  return formatter.format(diffDays, 'days')
}

export const patientFullName = (patient?: Patient) => {
  const name = patient?.name?.[0]
  if (!name) return 'Unknown'
  return `${name.given?.join(' ') ?? ''} ${name.family ?? ''}`.trim()
}

export const primaryTelecom = (patient?: Patient, system: 'email' | 'phone' = 'email') => {
  return patient?.telecom?.find(t => t.system === system)?.value ?? 'Not provided'
}

export const primaryAddress = (patient?: Patient) => {
  const address = patient?.address?.[0]
  if (!address) return 'Not provided'
  return [address.line?.join(', '), address.city, address.state, address.postalCode]
    .filter(Boolean)
    .join(', ')
}
