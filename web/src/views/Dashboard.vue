<template>
  <div class="pa-6">
    <!-- Page Header -->
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-gray-800 mb-2">Dashboard</h1>
      <p class="text-gray-600">Welcome to CareFlow - Healthcare Management System</p>
    </div>

    <!-- Stats Grid -->
    <v-row class="mb-8">
      <v-col cols="12" sm="6" md="3">
        <v-card class="border-l-4 border-l-blue-500">
          <v-card-text>
            <div class="text-sm text-gray-600 font-medium">Total Patients</div>
            <div class="text-3xl font-bold text-gray-800 mt-2">{{ patientStore.patientCount }}</div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card class="border-l-4 border-l-green-500">
          <v-card-text>
            <div class="text-sm text-gray-600 font-medium">Total Appointments</div>
            <div class="text-3xl font-bold text-gray-800 mt-2">
              {{ appointmentStore.appointmentCount }}
            </div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card class="border-l-4 border-l-yellow-500">
          <v-card-text>
            <div class="text-sm text-gray-600 font-medium">Upcoming</div>
            <div class="text-3xl font-bold text-gray-800 mt-2">
              {{ appointmentStore.upcomingAppointments.length }}
            </div>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" sm="6" md="3">
        <v-card class="border-l-4 border-l-red-500">
          <v-card-text>
            <div class="text-sm text-gray-600 font-medium">Cancelled</div>
            <div class="text-3xl font-bold text-gray-800 mt-2">
              {{ appointmentStore.appointments.filter(a => a.status === 'cancelled').length }}
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Quick Actions -->
    <v-row class="mb-8">
      <v-col cols="12">
        <v-card>
          <v-card-title class="pb-4">Quick Actions</v-card-title>
          <v-divider></v-divider>
          <v-card-text>
            <div class="flex gap-4 flex-wrap">
              <router-link to="/patients/new">
                <v-btn color="primary" variant="elevated">
                  <v-icon start>mdi-plus</v-icon>
                  New Patient
                </v-btn>
              </router-link>

              <router-link to="/appointments/new">
                <v-btn color="primary" variant="elevated">
                  <v-icon start>mdi-plus</v-icon>
                  New Appointment
                </v-btn>
              </router-link>

              <router-link to="/patients">
                <v-btn color="secondary" variant="outlined">
                  <v-icon start>mdi-list</v-icon>
                  View Patients
                </v-btn>
              </router-link>

              <router-link to="/appointments">
                <v-btn color="secondary" variant="outlined">
                  <v-icon start>mdi-list</v-icon>
                  View Appointments
                </v-btn>
              </router-link>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Upcoming Appointments Preview -->
    <v-row>
      <v-col cols="12" md="8">
        <v-card>
          <v-card-title class="pb-4">Upcoming Appointments</v-card-title>
          <v-divider></v-divider>

          <v-list v-if="appointmentStore.upcomingAppointments.length > 0">
            <v-list-item
              v-for="appointment in appointmentStore.upcomingAppointments.slice(0, 5)"
              :key="appointment.id"
            >
              <template v-slot:prepend>
                <v-icon color="primary">mdi-calendar-check</v-icon>
              </template>
              <v-list-item-title>
                {{ appointment.description }}
              </v-list-item-title>
              <v-list-item-subtitle>
                {{ formatDate(appointment.start) }}
              </v-list-item-subtitle>
            </v-list-item>
          </v-list>

          <v-card-text v-else class="text-center text-gray-600 py-8">
            No upcoming appointments
          </v-card-text>
        </v-card>
      </v-col>

      <!-- System Status -->
      <v-col cols="12" md="4">
        <v-card>
          <v-card-title class="pb-4">System Status</v-card-title>
          <v-divider></v-divider>
          <v-card-text>
            <div class="space-y-4">
              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-600">API Gateway</span>
                <v-chip size="small" color="success" variant="elevated">
                  <v-icon start size="small">mdi-check-circle</v-icon>
                  Healthy
                </v-chip>
              </div>

              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-600">Database</span>
                <v-chip size="small" color="success" variant="elevated">
                  <v-icon start size="small">mdi-check-circle</v-icon>
                  Healthy
                </v-chip>
              </div>

              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-600">Message Queue</span>
                <v-chip size="small" color="success" variant="elevated">
                  <v-icon start size="small">mdi-check-circle</v-icon>
                  Healthy
                </v-chip>
              </div>

              <div class="flex items-center justify-between">
                <span class="text-sm text-gray-600">Observability</span>
                <v-chip size="small" color="warning" variant="elevated">
                  <v-icon start size="small">mdi-alert-circle</v-icon>
                  Monitoring
                </v-chip>
              </div>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { usePatientStore } from '@/stores/patient'
import { useAppointmentStore } from '@/stores/appointment'

const patientStore = usePatientStore()
const appointmentStore = useAppointmentStore()

onMounted(async () => {
  // Load data on component mount
  await patientStore.fetchPatients()
  await appointmentStore.fetchAppointments()
})

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>
