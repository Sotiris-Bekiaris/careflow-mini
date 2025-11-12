<template>
  <div class="page">
    <router-link to="/patients" class="breadcrumb">
      <v-icon size="18" class="mr-1">mdi-chevron-left</v-icon>
      Patients
    </router-link>

    <SectionHeader
      v-if="patientStore.currentPatient"
      :title="patientFullName(patientStore.currentPatient)"
      :description="primaryTelecom(patientStore.currentPatient)"
      eyebrow="Patient detail"
    >
      <template #actions>
        <router-link :to="`/patients/${patientStore.currentPatient.id}/edit`">
          <v-btn variant="text" prepend-icon="mdi-pencil">Edit</v-btn>
        </router-link>
        <router-link to="/appointments/new">
          <v-btn color="primary" prepend-icon="mdi-plus">New appointment</v-btn>
        </router-link>
      </template>
    </SectionHeader>

    <div v-if="patientStore.loading" class="state">
      <v-progress-circular indeterminate color="primary" />
      <p>Retrieving FHIR resource…</p>
    </div>

    <v-alert v-else-if="patientStore.error" type="error" variant="tonal">
      {{ patientStore.error }}
    </v-alert>

    <v-row v-else-if="patientStore.currentPatient" class="gap-y-6">
      <v-col cols="12" md="8">
        <v-card class="panel">
          <div class="panel__section">
            <h3>Demographics</h3>
            <div class="grid">
              <div>
                <p>Gender</p>
                <strong>{{ capitalize(patientStore.currentPatient.gender) }}</strong>
              </div>
              <div>
                <p>Date of birth</p>
                <strong>{{ formatDate(patientStore.currentPatient.birthDate) }}</strong>
              </div>
              <div>
                <p>Status</p>
                <v-chip :color="patientStore.currentPatient.active ? 'success' : 'warning'">
                  {{ patientStore.currentPatient.active ? 'Active' : 'Inactive' }}
                </v-chip>
              </div>
            </div>
          </div>

          <v-divider></v-divider>

          <div class="panel__section">
            <h3>Contact</h3>
            <div class="grid">
              <div>
                <p>Email</p>
                <strong>{{ primaryTelecom(patientStore.currentPatient, 'email') }}</strong>
              </div>
              <div>
                <p>Phone</p>
                <strong>{{ primaryTelecom(patientStore.currentPatient, 'phone') }}</strong>
              </div>
              <div>
                <p>Address</p>
                <strong>{{ primaryAddress(patientStore.currentPatient) }}</strong>
              </div>
            </div>
          </div>

          <v-divider v-if="patientStore.currentPatient.identifier?.length"></v-divider>

          <div v-if="patientStore.currentPatient.identifier?.length" class="panel__section">
            <h3>Identifiers</h3>
            <v-list>
              <v-list-item
                v-for="(identifier, idx) in patientStore.currentPatient.identifier"
                :key="idx"
              >
                <v-list-item-title>{{ identifier.system }}</v-list-item-title>
                <v-list-item-subtitle>{{ identifier.value }}</v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </div>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <v-card class="panel">
          <div class="panel__section">
            <p class="subtitle">Patient ID</p>
            <code>{{ patientStore.currentPatient.id }}</code>
          </div>
          <v-divider></v-divider>
          <div class="panel__section">
            <p class="subtitle">Created</p>
            <strong>{{ formatDateTime(patientStore.currentPatient.meta?.created) }}</strong>
          </div>
          <v-divider></v-divider>
          <div class="panel__section">
            <p class="subtitle">Last updated</p>
            <strong>{{ formatDateTime(patientStore.currentPatient.meta?.updated) }}</strong>
          </div>
          <v-divider></v-divider>
          <div class="panel__actions">
            <v-btn variant="outlined" block prepend-icon="mdi-download" @click="handleExport">
              Export FHIR JSON
            </v-btn>
            <v-btn color="error" block prepend-icon="mdi-trash-can" @click="handleDelete">
              Delete patient
            </v-btn>
          </div>
        </v-card>
      </v-col>

      <v-col cols="12">
        <v-card class="panel">
          <div class="panel__section">
            <h3>Lab Results & Observations</h3>
            <p v-if="observationStore.loading" class="text-muted">Loading observations...</p>
            <p v-else-if="!observationStore.hasObservations" class="text-muted">
              No lab results recorded for this patient.
            </p>
            <v-table v-else density="comfortable">
              <thead>
                <tr>
                  <th>Test</th>
                  <th>Result</th>
                  <th>Status</th>
                  <th>Date</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="observation in observationStore.observations" :key="observation.id">
                  <td>
                    <strong>{{ observation.code.text }}</strong>
                  </td>
                  <td>
                    {{ observation.value.value }} {{ observation.value.unit }}
                  </td>
                  <td>
                    <v-chip
                      size="small"
                      :color="observation.status === 'final' ? 'success' : 'warning'"
                    >
                      {{ observation.status }}
                    </v-chip>
                  </td>
                  <td>{{ formatDateTime(observation.effectiveDateTime) }}</td>
                </tr>
              </tbody>
            </v-table>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <EmptyState
      v-else
      icon="mdi-hospital-box"
      title="Patient not found"
      description="Select a record from the list or create a new one to test the service."
    >
      <template #actions>
        <router-link to="/patients">
          <v-btn variant="text">Back to list</v-btn>
        </router-link>
        <router-link to="/patients/new">
          <v-btn color="primary">New patient</v-btn>
        </router-link>
      </template>
    </EmptyState>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePatientStore } from '@/stores/patient'
import { useObservationStore } from '@/stores/observation'
import SectionHeader from '@/components/ui/SectionHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import { formatDate, formatDateTime, patientFullName, primaryTelecom, primaryAddress } from '@/utils/formatters'

const route = useRoute()
const router = useRouter()
const patientStore = usePatientStore()
const observationStore = useObservationStore()

onMounted(async () => {
  const patientId = route.params.id as string
  await patientStore.fetchPatientById(patientId)
  await observationStore.fetchPatientObservations(patientId)
})

const capitalize = (value: string) => value.charAt(0).toUpperCase() + value.slice(1)

const handleExport = () => {
  if (!patientStore.currentPatient) return
  const jsonStr = JSON.stringify(patientStore.currentPatient, null, 2)
  const blob = new Blob([jsonStr], { type: 'application/fhir+json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `patient-${patientStore.currentPatient.id}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

const handleDelete = async () => {
  if (!patientStore.currentPatient) return
  if (!confirm('Delete this patient? This cannot be undone.')) return
  try {
    await patientStore.deletePatient(patientStore.currentPatient.id)
    router.push('/patients')
  } catch (error) {
    console.error('Failed to delete patient:', error)
  }
}
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.breadcrumb {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  color: var(--cf-text-muted);
}

.state {
  padding: 3rem 0;
  text-align: center;
  color: var(--cf-text-muted);
  display: grid;
  gap: 1rem;
}

.panel {
  border-radius: var(--cf-radius-lg);
  box-shadow: var(--cf-shadow-soft);
}

.panel__section {
  padding: 1.5rem;
}

.panel__section .grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
}

.panel__section p {
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.7rem;
  color: var(--cf-text-muted);
}

.panel__section strong {
  font-size: 1rem;
}

.subtitle {
  margin: 0 0 0.25rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.7rem;
  color: var(--cf-text-muted);
}

code {
  background: rgba(15, 23, 42, 0.05);
  padding: 0.4rem 0.6rem;
  border-radius: 10px;
  display: inline-block;
}

.panel__actions {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1.5rem;
}
</style>
