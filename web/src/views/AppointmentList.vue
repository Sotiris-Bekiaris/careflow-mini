<template>
  <div class="pa-6">
    <!-- Page Header -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-3xl font-bold text-gray-800">Appointments</h1>
        <p class="text-gray-600 mt-1">Schedule and manage appointments</p>
      </div>

      <router-link to="/appointments/new">
        <v-btn color="primary" prepend-icon="mdi-plus" size="large">
          New Appointment
        </v-btn>
      </router-link>
    </div>

    <!-- Search and Filter -->
    <v-card class="mb-6">
      <v-card-text class="pt-4">
        <v-row>
          <v-col cols="12" md="6">
            <v-text-field
              v-model="searchQuery"
              label="Search appointments..."
              prepend-inner-icon="mdi-magnify"
              variant="outlined"
              density="compact"
              clearable
            />
          </v-col>

          <v-col cols="12" md="6">
            <v-select
              v-model="appointmentStore.filterStatus"
              :items="statusOptions"
              label="Filter by status"
              variant="outlined"
              density="compact"
              @update:model-value="appointmentStore.setFilterStatus"
            />
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>

    <!-- Loading State -->
    <v-card v-if="appointmentStore.loading" class="mb-6">
      <v-card-text class="text-center py-8">
        <v-progress-circular indeterminate color="primary" />
        <p class="mt-4 text-gray-600">Loading appointments...</p>
      </v-card-text>
    </v-card>

    <!-- Error State -->
    <v-card v-else-if="appointmentStore.error" class="mb-6" color="error">
      <v-card-text class="d-flex align-center gap-3">
        <v-icon color="white">mdi-alert-circle</v-icon>
        <div>
          <p class="text-white font-medium">Error loading appointments</p>
          <p class="text-white text-sm">{{ appointmentStore.error }}</p>
        </div>
      </v-card-text>
    </v-card>

    <!-- Appointments Table -->
    <v-card v-else>
      <v-table v-if="appointmentStore.hasAppointments">
        <thead>
          <tr>
            <th class="text-left">Description</th>
            <th class="text-left">Start Time</th>
            <th class="text-left">End Time</th>
            <th class="text-left">Status</th>
            <th class="text-left">Participant</th>
            <th class="text-center">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="appointment in filteredAppointments" :key="appointment.id">
            <td>
              <strong>{{ appointment.description }}</strong>
            </td>
            <td>{{ formatDateTime(appointment.start) }}</td>
            <td>{{ formatDateTime(appointment.end) }}</td>
            <td>
              <v-chip :color="getStatusColor(appointment.status)" variant="elevated" size="small">
                {{ capitalize(appointment.status) }}
              </v-chip>
            </td>
            <td>
              {{ appointment.participant?.[0]?.actor?.display || 'N/A' }}
            </td>
            <td class="text-center">
              <v-btn
                icon
                size="x-small"
                variant="text"
                color="primary"
                title="View"
              >
                <v-icon size="small">mdi-eye</v-icon>
              </v-btn>

              <router-link :to="`/appointments/${appointment.id}/edit`">
                <v-btn
                  icon
                  size="x-small"
                  variant="text"
                  color="info"
                  title="Edit"
                >
                  <v-icon size="small">mdi-pencil</v-icon>
                </v-btn>
              </router-link>

              <v-btn
                icon
                size="x-small"
                variant="text"
                color="error"
                title="Delete"
                @click="deleteAppointment(appointment.id)"
              >
                <v-icon size="small">mdi-trash-can</v-icon>
              </v-btn>
            </td>
          </tr>
        </tbody>
      </v-table>

      <v-card-text v-else class="text-center py-12 text-gray-600">
        <v-icon size="48" class="mb-4 text-gray-400">mdi-calendar-check</v-icon>
        <p class="text-lg">No appointments found</p>
        <router-link to="/appointments/new" class="mt-4">
          <v-btn color="primary" variant="outlined">
            Create First Appointment
          </v-btn>
        </router-link>
      </v-card-text>
    </v-card>

    <!-- Pagination (Placeholder) -->
    <div v-if="appointmentStore.hasAppointments" class="mt-6 flex justify-center">
      <v-pagination
        v-model="currentPage"
        :length="totalPages"
        color="primary"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppointmentStore } from '@/stores/appointment'

const appointmentStore = useAppointmentStore()
const searchQuery = ref('')
const currentPage = ref(1)
const itemsPerPage = 10

const statusOptions = [
  { title: 'All', value: 'all' },
  { title: 'Proposed', value: 'proposed' },
  { title: 'Pending', value: 'pending' },
  { title: 'Booked', value: 'booked' },
  { title: 'Arrived', value: 'arrived' },
  { title: 'Fulfilled', value: 'fulfilled' },
  { title: 'Cancelled', value: 'cancelled' },
  { title: 'No Show', value: 'noshow' },
]

onMounted(async () => {
  await appointmentStore.fetchAppointments()
})

const filteredAppointments = computed(() => {
  let filtered = appointmentStore.filteredAppointments

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(appointment => {
      return appointment.description.toLowerCase().includes(query)
    })
  }

  return filtered
})

const totalPages = computed(() => {
  return Math.ceil(filteredAppointments.value.length / itemsPerPage)
})

const getStatusColor = (status: string) => {
  const colors: Record<string, string> = {
    proposed: 'info',
    pending: 'warning',
    booked: 'success',
    arrived: 'primary',
    fulfilled: 'success',
    cancelled: 'error',
    noshow: 'error',
  }
  return colors[status] || 'secondary'
}

const formatDateTime = (dateString: string) => {
  return new Date(dateString).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const capitalize = (text: string) => {
  return text
    .replace(/([A-Z])/g, ' $1')
    .charAt(0)
    .toUpperCase() + text.slice(1)
}

const deleteAppointment = async (id: string) => {
  if (confirm('Are you sure you want to delete this appointment?')) {
    try {
      await appointmentStore.deleteAppointment(id)
      appointmentStore.clearError()
    } catch (error) {
      console.error('Failed to delete appointment:', error)
    }
  }
}
</script>
