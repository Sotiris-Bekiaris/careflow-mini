<template>
  <div class="page">
    <SectionHeader
      :title="isEditing ? 'Edit patient' : 'Create patient'"
      description="FHIR-compliant payload sent to the Go patient service"
      eyebrow="Patient builder"
    >
      <template #actions>
        <router-link to="/patients">
          <v-btn variant="text">Back to list</v-btn>
        </router-link>
      </template>
    </SectionHeader>

    <v-card class="panel">
      <v-form ref="form" @submit.prevent="submitForm">
        <div class="form-grid">
          <div>
            <h3>Identity</h3>
            <p>Captured as FHIR name + demographic attributes.</p>
          </div>
          <div class="grid">
            <v-text-field
              v-model="formData.firstName"
              label="First name"
              :rules="[requiredRule]"
              variant="solo"
              density="comfortable"
            />
            <v-text-field
              v-model="formData.lastName"
              label="Last name"
              :rules="[requiredRule]"
              variant="solo"
              density="comfortable"
            />
            <v-select
              v-model="formData.gender"
              :items="genderOptions"
              label="Gender"
              variant="solo"
              density="comfortable"
            />
            <v-text-field
              v-model="formData.birthDate"
              label="Date of birth"
              type="date"
              variant="solo"
              density="comfortable"
            />
          </div>
        </div>

        <v-divider class="my-6"></v-divider>

        <div class="form-grid">
          <div>
            <h3>Contact</h3>
            <p>Primary telecom channels for care teams.</p>
          </div>
          <div class="grid">
            <v-text-field
              v-model="formData.email"
              label="Email"
              type="email"
              variant="solo"
              density="comfortable"
            />
            <v-text-field
              v-model="formData.phone"
              label="Phone"
              type="tel"
              variant="solo"
              density="comfortable"
            />
            <v-text-field v-model="formData.addressLine" label="Address line" variant="solo" />
            <v-text-field v-model="formData.city" label="City" variant="solo" />
            <v-text-field v-model="formData.state" label="State/Province" variant="solo" />
            <v-text-field v-model="formData.postalCode" label="Postal code" variant="solo" />
            <v-text-field v-model="formData.country" label="Country" variant="solo" />
          </div>
        </div>

        <v-divider class="my-6"></v-divider>

        <v-switch v-model="formData.active" inset color="primary" label="Active record" />

        <div class="actions">
          <v-btn type="submit" color="primary" size="large" :loading="patientStore.loading">
            <v-icon start>mdi-check</v-icon>
            {{ isEditing ? 'Update patient' : 'Create patient' }}
          </v-btn>
          <router-link to="/patients">
            <v-btn size="large" variant="text">Cancel</v-btn>
          </router-link>
        </div>
      </v-form>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePatientStore } from '@/stores/patient'
import { useUIStore } from '@/stores/ui'
import SectionHeader from '@/components/ui/SectionHeader.vue'

const route = useRoute()
const router = useRouter()
const patientStore = usePatientStore()
const uiStore = useUIStore()
const form = ref()

const isEditing = computed(() => !!route.params.id)

const formData = reactive({
  firstName: '',
  lastName: '',
  gender: 'unknown',
  birthDate: '',
  email: '',
  phone: '',
  addressLine: '',
  city: '',
  state: '',
  postalCode: '',
  country: '',
  active: true,
})

const genderOptions = ['female', 'male', 'other', 'unknown']
const requiredRule = (value: string) => (!!value && value.trim().length > 0) || 'Required'

onMounted(async () => {
  if (!isEditing.value) return
  await patientStore.fetchPatientById(route.params.id as string)
  const patient = patientStore.currentPatient
  if (!patient) return
  const name = patient.name?.[0]
  const address = patient.address?.[0]
  const email = patient.telecom?.find(t => t.system === 'email')
  const phone = patient.telecom?.find(t => t.system === 'phone')
  formData.firstName = name?.given?.[0] ?? ''
  formData.lastName = name?.family ?? ''
  formData.gender = patient.gender ?? 'unknown'
  formData.birthDate = patient.birthDate ?? ''
  formData.email = email?.value ?? ''
  formData.phone = phone?.value ?? ''
  formData.addressLine = address?.line?.[0] ?? ''
  formData.city = address?.city ?? ''
  formData.state = address?.state ?? ''
  formData.postalCode = address?.postalCode ?? ''
  formData.country = address?.country ?? ''
  formData.active = patient.active
})

const submitForm = async () => {
  const result = await form.value?.validate()
  if (!result?.valid) return

  const payload = {
    resourceType: 'Patient' as const,
    name: [
      {
        given: [formData.firstName].filter(Boolean),
        family: formData.lastName,
      },
    ],
    gender: formData.gender as any,
    birthDate: formData.birthDate,
    telecom: [
      ...(formData.email ? [{ system: 'email' as const, value: formData.email }] : []),
      ...(formData.phone ? [{ system: 'phone' as const, value: formData.phone }] : []),
    ],
    address: [
      {
        line: formData.addressLine ? [formData.addressLine] : [],
        city: formData.city,
        state: formData.state,
        postalCode: formData.postalCode,
        country: formData.country,
      },
    ],
    active: formData.active,
    identifier: [],
  }

  try {
    if (isEditing.value) {
      await patientStore.updatePatient(route.params.id as string, payload as any)
    } else {
      await patientStore.createPatient(payload as any)
    }

    uiStore.addNotification(`Patient ${isEditing.value ? 'updated' : 'created'} successfully`, 'success')
    router.push('/patients')
  } catch (error) {
    uiStore.addNotification('Unable to save patient', 'error')
    console.error('Failed to save patient:', error)
  }
}
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.panel {
  border-radius: var(--cf-radius-lg);
  box-shadow: var(--cf-shadow-soft);
  padding: 2rem;
}

.form-grid {
  display: grid;
  grid-template-columns: 250px 1fr;
  gap: 2rem;
  align-items: start;
}

.form-grid h3 {
  margin: 0 0 0.25rem;
}

.form-grid p {
  margin: 0;
  color: var(--cf-text-muted);
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.actions {
  margin-top: 2rem;
  display: flex;
  gap: 1rem;
}

@media (max-width: 960px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
