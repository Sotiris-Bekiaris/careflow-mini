<template>
  <div class="dashboard">
    <section class="dashboard__hero">
      <div>
        <p class="dashboard__eyebrow">End-to-end patient journey</p>
        <h2>CareFlow Control Center</h2>
        <p>
          Observe every hop of the Go microservices stack, launch new requests, and prove the
          reliability story in one minimal console.
        </p>
        <div class="dashboard__hero-actions">
          <router-link to="/patients/new">
            <v-btn color="primary" size="large" prepend-icon="mdi-plus">New Patient</v-btn>
          </router-link>
          <router-link to="/appointments/new">
            <v-btn variant="text" size="large" prepend-icon="mdi-sparkles">Schedule demo flow</v-btn>
          </router-link>
        </div>
      </div>
      <div class="dashboard__hero-card">
        <p>Stack pulse</p>
        <h3>{{ systemStore.healthyCount }} services healthy</h3>
        <p class="text-muted">
          {{ systemStore.lastUpdated ? `Checked ${formatRelativeTime(systemStore.lastUpdated)}` : 'Awaiting first probe' }}
        </p>
      </div>
    </section>

    <SectionHeader
      class="mt-8"
      eyebrow="Signals"
      title="What the Go services are doing right now"
      description="Live counters pulled from the gateway and downstream patient + appointment services."
    />

    <v-row class="dashboard__metrics">
      <v-col cols="12" sm="6" md="3">
        <MetricCard
          title="Patients"
          :value="patientStore.patientCount.toString()"
          subtitle="FHIR resources"
          icon="mdi-account-heart"
          :loading="patientStore.loading"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <MetricCard
          title="Appointments"
          :value="appointmentStore.appointmentCount.toString()"
          subtitle="Scheduled"
          icon="mdi-calendar"
          :loading="appointmentStore.loading"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <MetricCard
          title="Upcoming"
          :value="appointmentStore.upcomingAppointments.length.toString()"
          subtitle="Next 30 days"
          icon="mdi-timer-sand"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <MetricCard
          title="Cancelled"
          :value="cancelledCount.toString()"
          subtitle="Action needed"
          icon="mdi-alert"
          :delta="{ value: trendLabel, trend: trendDirection, label: 'vs last sync' }"
        />
      </v-col>
    </v-row>

    <v-row class="gap-y-6">
      <v-col cols="12" md="8">
        <v-card class="panel">
          <div class="panel__header">
            <div>
              <p class="panel__eyebrow">Timeline</p>
              <h3>Upcoming appointments</h3>
            </div>
            <router-link to="/appointments">
              <v-btn variant="text" color="primary">View all</v-btn>
            </router-link>
          </div>
          <v-divider></v-divider>

          <div v-if="upcomingPreview.length" class="timeline">
            <div v-for="appointment in upcomingPreview" :key="appointment.id" class="timeline__item">
              <div class="timeline__dot"></div>
              <div>
                <p class="timeline__title">{{ appointment.description }}</p>
                <p class="timeline__meta">
                  {{ formatDateTime(appointment.start) }} —
                  {{ appointment.participant?.[0]?.actor?.display || 'Unassigned' }}
                </p>
              </div>
              <v-chip size="small" variant="flat" :color="statusColor(appointment.status)">
                {{ formatStatus(appointment.status) }}
              </v-chip>
            </div>
          </div>

          <EmptyState
            v-else
            icon="mdi-calendar-remove"
            title="No upcoming appointments"
            description="Create a booking to exercise the appointment service and emit downstream events."
          >
            <template #actions>
              <router-link to="/appointments/new">
                <v-btn color="primary">Create appointment</v-btn>
              </router-link>
            </template>
          </EmptyState>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <SystemStatusCard :services="systemStore.services" :loading="systemStore.loading" />
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { usePatientStore } from '@/stores/patient'
import { useAppointmentStore } from '@/stores/appointment'
import { useSystemStore } from '@/stores/system'
import MetricCard from '@/components/ui/MetricCard.vue'
import SectionHeader from '@/components/ui/SectionHeader.vue'
import SystemStatusCard from '@/components/ui/SystemStatusCard.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import { formatDateTime, formatRelativeTime } from '@/utils/formatters'

const patientStore = usePatientStore()
const appointmentStore = useAppointmentStore()
const systemStore = useSystemStore()

onMounted(async () => {
  if (!patientStore.patients.length) {
    await patientStore.fetchPatients()
  }
  if (!appointmentStore.appointments.length) {
    await appointmentStore.fetchAppointments()
  }
  if (!systemStore.services.length) {
    await systemStore.refresh()
  }
})

const upcomingPreview = computed(() => appointmentStore.upcomingAppointments.slice(0, 5))
const cancelledCount = computed(
  () => appointmentStore.appointments.filter(appointment => appointment.status === 'cancelled').length,
)

const trendLabel = computed(() => `${cancelledCount.value} alerts`)
const trendDirection = computed(() => (cancelledCount.value > 0 ? 'up' : 'flat'))

const statusColor = (status: string) => {
  const mapping: Record<string, string> = {
    cancelled: 'error',
    pending: 'warning',
    booked: 'success',
    proposed: 'info',
  }
  return mapping[status] || 'info'
}

const formatStatus = (status: string) => status.charAt(0).toUpperCase() + status.slice(1)
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.dashboard__hero {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 2rem;
  padding: 2.5rem;
  border-radius: var(--cf-radius-lg);
  background: linear-gradient(135deg, #fefefe 0%, #ecf2ff 40%, #f5f5f7 100%);
  box-shadow: var(--cf-shadow-soft);
}

.dashboard__hero h2 {
  font-size: 2.5rem;
  margin-bottom: 0.75rem;
}

.dashboard__hero p {
  margin: 0 0 1rem;
  color: var(--cf-text-muted);
}

.dashboard__hero-actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.dashboard__hero-card {
  border-radius: 24px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  padding: 1.5rem;
  background: rgba(255, 255, 255, 0.85);
}

.dashboard__hero-card h3 {
  font-size: 1.8rem;
  margin: 0.25rem 0;
}

.dashboard__eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--cf-text-muted);
  font-size: 0.8rem;
}

.dashboard__metrics {
  margin-bottom: 1rem;
}

.panel {
  border-radius: var(--cf-radius-lg);
  box-shadow: var(--cf-shadow-soft);
}

.panel__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 1.5rem 0.5rem;
}

.panel__eyebrow {
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--cf-text-muted);
  font-size: 0.75rem;
}

.panel h3 {
  margin: 0.2rem 0 0;
}

.timeline {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.timeline__item {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 1rem;
  align-items: center;
}

.timeline__dot {
  width: 14px;
  height: 14px;
  border-radius: 999px;
  background: #0a84ff;
  box-shadow: 0 0 0 6px rgba(10, 132, 255, 0.12);
}

.timeline__title {
  margin: 0;
  font-weight: 600;
}

.timeline__meta {
  margin: 0;
  color: var(--cf-text-muted);
}

.text-muted {
  color: var(--cf-text-muted);
}

@media (max-width: 960px) {
  .dashboard__hero {
    grid-template-columns: 1fr;
  }
}
</style>
