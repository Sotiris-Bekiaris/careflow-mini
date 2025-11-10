<template>
  <div class="page">
    <SectionHeader
      title="Appointments"
      description="Orchestration between the appointment service and downstream events"
      eyebrow="Scheduling"
    >
      <template #actions>
        <router-link to="/appointments/new">
          <v-btn color="primary" prepend-icon="mdi-plus">New appointment</v-btn>
        </router-link>
      </template>
    </SectionHeader>

    <v-card class="panel">
      <div class="filters">
        <v-text-field
          v-model="searchQuery"
          label="Search appointments"
          prepend-inner-icon="mdi-magnify"
          variant="solo"
          hide-details
          clearable
        />
        <v-select
          v-model="appointmentStore.filterStatus"
          :items="statusOptions"
          label="Status"
          variant="solo"
          hide-details
          @update:model-value="value => appointmentStore.setFilterStatus(value)"
        />
      </div>

      <v-divider></v-divider>

      <div v-if="appointmentStore.loading" class="state">
        <v-progress-circular indeterminate color="primary" />
        <p>Loading appointments…</p>
      </div>

      <v-alert v-else-if="appointmentStore.error" type="error" variant="tonal" class="ma-6">
        {{ appointmentStore.error }}
      </v-alert>

      <template v-else>
        <div v-if="pagedAppointments.length" class="table-wrapper">
          <v-table>
            <thead>
              <tr>
                <th>Description</th>
                <th>Start</th>
                <th>End</th>
                <th>Status</th>
                <th>Participant</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="appointment in pagedAppointments" :key="appointment.id">
                <td>
                  <div class="cell-primary">
                    <h4>{{ appointment.description }}</h4>
                    <p>{{ appointment.id }}</p>
                  </div>
                </td>
                <td>{{ formatDateTime(appointment.start) }}</td>
                <td>{{ formatDateTime(appointment.end) }}</td>
                <td>
                  <v-chip size="small" :color="statusColor(appointment.status)" variant="flat">
                    {{ formatStatus(appointment.status) }}
                  </v-chip>
                </td>
                <td>{{ appointment.participant?.[0]?.actor?.display || 'Unassigned' }}</td>
                <td class="actions">
                  <router-link :to="`/appointments/${appointment.id}/edit`">
                    <v-btn icon variant="text">
                      <v-icon>mdi-pencil-outline</v-icon>
                    </v-btn>
                  </router-link>
                  <v-btn icon variant="text" color="error" @click="deleteAppointment(appointment.id)">
                    <v-icon>mdi-trash-can-outline</v-icon>
                  </v-btn>
                </td>
              </tr>
            </tbody>
          </v-table>
        </div>

        <EmptyState
          v-else
          icon="mdi-calendar-multiselect"
          title="No appointments"
          description="Trigger the appointment service to emit booked events and update the dashboard."
        >
          <template #actions>
            <router-link to="/appointments/new">
              <v-btn color="primary">Create appointment</v-btn>
            </router-link>
          </template>
        </EmptyState>

        <div v-if="filteredAppointments.length > itemsPerPage" class="pagination">
          <v-pagination
            v-model="currentPage"
            :length="totalPages"
            rounded="circle"
            color="primary"
          />
        </div>
      </template>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppointmentStore } from '@/stores/appointment'
import SectionHeader from '@/components/ui/SectionHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import { formatDateTime } from '@/utils/formatters'

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
  if (!appointmentStore.appointments.length) {
    await appointmentStore.fetchAppointments()
  }
})

const filteredAppointments = computed(() => {
  let dataset = appointmentStore.filteredAppointments
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    dataset = dataset.filter(appointment => appointment.description.toLowerCase().includes(query))
  }
  return dataset
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredAppointments.value.length / itemsPerPage)))

const pagedAppointments = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage
  return filteredAppointments.value.slice(start, start + itemsPerPage)
})

const statusColor = (status: string) => {
  const colors: Record<string, string> = {
    booked: 'success',
    pending: 'warning',
    cancelled: 'error',
    fulfilled: 'primary',
    proposed: 'info',
    noshow: 'error',
  }
  return colors[status] || 'secondary'
}

const formatStatus = (status: string) => status.charAt(0).toUpperCase() + status.slice(1)

const deleteAppointment = async (id: string) => {
  if (!confirm('Delete this appointment?')) return
  try {
    await appointmentStore.deleteAppointment(id)
  } catch (error) {
    console.error('Failed to delete appointment:', error)
  }
}
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}

.panel {
  border-radius: var(--cf-radius-lg);
  box-shadow: var(--cf-shadow-soft);
}

.filters {
  display: grid;
  grid-template-columns: 1fr 200px;
  gap: 1rem;
  padding: 1.5rem;
}

.state {
  padding: 3rem 0;
  text-align: center;
  color: var(--cf-text-muted);
  display: grid;
  gap: 1rem;
}

.table-wrapper {
  padding: 1.5rem;
}

thead tr {
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 0.75rem;
  color: var(--cf-text-muted);
}

.cell-primary h4 {
  margin: 0;
}

.cell-primary p {
  margin: 0;
  color: var(--cf-text-muted);
  font-size: 0.85rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.25rem;
}

.pagination {
  padding: 1rem 0 2rem;
  display: flex;
  justify-content: center;
}

@media (max-width: 840px) {
  .filters {
    grid-template-columns: 1fr;
  }
  .table-wrapper {
    overflow-x: auto;
  }
}
</style>
