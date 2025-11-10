<template>
  <div class="pa-6">
    <!-- Page Header -->
    <div class="flex justify-between items-center mb-6">
      <router-link to="/patients">
        <v-btn icon variant="text" color="primary">
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>
      </router-link>

      <h1 class="text-3xl font-bold text-gray-800 flex-1">
        {{ isEditing ? 'Edit Patient' : 'Create Patient' }}
      </h1>
    </div>

    <!-- Form Card -->
    <v-card>
      <v-card-text class="pt-6">
        <v-form ref="form" @submit.prevent="submitForm">
          <v-row>
            <!-- First Name -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.firstName"
                label="First Name"
                variant="outlined"
                density="compact"
                rules="required"
                required
              />
            </v-col>

            <!-- Last Name -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.lastName"
                label="Last Name"
                variant="outlined"
                density="compact"
                rules="required"
                required
              />
            </v-col>

            <!-- Gender -->
            <v-col cols="12" sm="6">
              <v-select
                v-model="formData.gender"
                :items="['male', 'female', 'other', 'unknown']"
                label="Gender"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Date of Birth -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.birthDate"
                label="Date of Birth"
                type="date"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Email -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.email"
                label="Email"
                type="email"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Phone -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.phone"
                label="Phone"
                type="tel"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Address -->
            <v-col cols="12">
              <v-text-field
                v-model="formData.addressLine"
                label="Address"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- City -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.city"
                label="City"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- State -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.state"
                label="State/Province"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Postal Code -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.postalCode"
                label="Postal Code"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Country -->
            <v-col cols="12" sm="6">
              <v-text-field
                v-model="formData.country"
                label="Country"
                variant="outlined"
                density="compact"
              />
            </v-col>

            <!-- Active Status -->
            <v-col cols="12">
              <v-checkbox
                v-model="formData.active"
                label="Active"
                color="primary"
              />
            </v-col>
          </v-row>

          <!-- Form Actions -->
          <v-row class="mt-6">
            <v-col cols="12" class="d-flex gap-3">
              <v-btn type="submit" color="primary" size="large">
                <v-icon start>mdi-check</v-icon>
                {{ isEditing ? 'Update Patient' : 'Create Patient' }}
              </v-btn>

              <v-btn
                variant="outlined"
                size="large"
                @click="$router.push('/patients')"
              >
                Cancel
              </v-btn>
            </v-col>
          </v-row>
        </v-form>
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePatientStore } from '@/stores/patient'

const route = useRoute()
const router = useRouter()
const patientStore = usePatientStore()
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

onMounted(async () => {
  if (isEditing.value) {
    const id = route.params.id as string
    await patientStore.fetchPatientById(id)

    if (patientStore.currentPatient) {
      const patient = patientStore.currentPatient
      const name = patient.name[0]
      const address = patient.address[0]
      const email = patient.telecom?.find(t => t.system === 'email')
      const phone = patient.telecom?.find(t => t.system === 'phone')

      formData.firstName = name?.given?.[0] || ''
      formData.lastName = name?.family || ''
      formData.gender = patient.gender || 'unknown'
      formData.birthDate = patient.birthDate || ''
      formData.email = email?.value || ''
      formData.phone = phone?.value || ''
      formData.addressLine = address?.line?.[0] || ''
      formData.city = address?.city || ''
      formData.state = address?.state || ''
      formData.postalCode = address?.postalCode || ''
      formData.country = address?.country || ''
      formData.active = patient.active
    }
  }
})

const submitForm = async () => {
  if (form.value && await form.value.validate()) {
    // TODO: Build proper Patient object from form data
    const patientData = {
      resourceType: 'Patient' as const,
      name: [{
        given: [formData.firstName],
        family: formData.lastName,
      }],
      gender: formData.gender as any,
      birthDate: formData.birthDate,
      telecom: [
        { system: 'email' as const, value: formData.email },
        { system: 'phone' as const, value: formData.phone },
      ],
      address: [{
        line: formData.addressLine ? [formData.addressLine] : [],
        city: formData.city,
        state: formData.state,
        postalCode: formData.postalCode,
        country: formData.country,
      }],
      active: formData.active,
      identifier: [],
    }

    try {
      if (isEditing.value) {
        await patientStore.updatePatient(route.params.id as string, patientData as any)
      } else {
        await patientStore.createPatient(patientData as any)
      }

      router.push('/patients')
      patientStore.addNotification(
        `Patient ${isEditing.value ? 'updated' : 'created'} successfully`,
        'success',
      )
    } catch (error) {
      console.error('Failed to save patient:', error)
    }
  }
}
</script>
