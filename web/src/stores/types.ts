/**
 * Type definitions for FHIR-aligned resources
 * These match the backend API structures
 */

export interface FHIRIdentifier {
  system: string
  value: string
}

export interface FHIRName {
  use: string
  given: string[]
  family: string
}

export interface FHIRAddress {
  use: string
  line: string[]
  city: string
  state: string
  postalCode: string
  country: string
}

export interface FHIRTelecom {
  system: 'phone' | 'email' | 'url'
  use: string
  value: string
}

export interface Patient {
  id: string
  resourceType: 'Patient'
  identifier: FHIRIdentifier[]
  name: FHIRName[]
  gender: 'male' | 'female' | 'other' | 'unknown'
  birthDate: string
  address: FHIRAddress[]
  telecom: FHIRTelecom[]
  active: boolean
  meta: {
    created: string
    updated: string
  }
}

export interface Appointment {
  id: string
  resourceType: 'Appointment'
  status: 'proposed' | 'pending' | 'booked' | 'arrived' | 'fulfilled' | 'cancelled' | 'noshow'
  description: string
  start: string
  end: string
  participant: Array<{
    actor: {
      reference: string
      display: string
    }
    status: string
  }>
  meta: {
    created: string
    updated: string
  }
}

export interface Observation {
  id: string
  resourceType: 'Observation'
  code: {
    coding: Array<{
      system: string
      code: string
      display: string
    }>
    text: string
  }
  value: {
    value: number
    unit: string
  }
  status: 'preliminary' | 'final' | 'amended'
  subject: {
    reference: string
    display: string
  }
  effectiveDateTime: string
}

export interface UINotification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  message: string
  duration?: number
}
