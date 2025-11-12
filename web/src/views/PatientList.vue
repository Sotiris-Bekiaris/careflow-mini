<template>
  <div class="page">
    <SectionHeader
      title="Patients"
      description="FHIR-native records from the Go patient service"
      eyebrow="Registry"
    >
      <template #actions>
        <router-link to="/patients/new">
          <v-btn color="primary" prepend-icon="mdi-plus">Add patient</v-btn>
        </router-link>
      </template>
    </SectionHeader>

    <v-card class="panel">
      <div class="filters">
        <v-text-field
          v-model="searchQuery"
          prepend-inner-icon="mdi-magnify"
          label="Search by name or contact"
          variant="solo"
          density="comfortable"
          hide-details
          clearable
        />
        <v-select
          v-model="filterGender"
          :items="genderOptions"
          label="Gender"
          variant="solo"
          density="comfortable"
          hide-details
        />
      </div>

      <v-divider></v-divider>

      <div v-if="patientStore.loading" class="state">
        <v-progress-circular indeterminate color="primary" />
        <p>Loading patients from the gateway…</p>
      </div>

      <v-alert v-else-if="patientStore.error" type="error" variant="tonal" class="ma-6">
        {{ patientStore.error }}
      </v-alert>

      <template v-else>
        <div v-if="filteredPatients.length" class="table-wrapper">
          <v-table>
            <thead>
              <tr>
                <th>Patient</th>
                <th>Gender</th>
                <th>DOB</th>
                <th>Contact</th>
                <th>Status</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="patient in pagedPatients" :key="patient.id">
                <td>
                  <div class="cell-primary">
                    <h4>{{ patientFullName(patient) }}</h4>
                    <p>ID: {{ patient.id }}</p>
                  </div>
                </td>
                <td>
                  <v-chip size="small" variant="flat" color="black">{{
                    capitalize(patient.gender)
                  }}</v-chip>
                </td>
                <td>{{ formatDate(patient.birthDate) }}</td>
                <td>{{ primaryTelecom(patient, 'email') }}</td>
                <td>
                  <v-chip
                    size="small"
                    :color="patient.active ? 'success' : 'warning'"
                    variant="flat"
                  >
                    {{ patient.active ? 'Active' : 'Inactive' }}
                  </v-chip>
                </td>
                <td class="actions">
                  <router-link :to="`/patients/${patient.id}`">
                    <v-btn icon variant="text">
                      <v-icon>mdi-eye-outline</v-icon>
                    </v-btn>
                  </router-link>
                  <router-link :to="`/patients/${patient.id}/edit`">
                    <v-btn icon variant="text">
                      <v-icon>mdi-pencil-outline</v-icon>
                    </v-btn>
                  </router-link>
                  <v-btn icon variant="text" color="error" @click="deletePatient(patient.id)">
                    <v-icon>mdi-trash-can-outline</v-icon>
                  </v-btn>
                </td>
              </tr>
            </tbody>
          </v-table>
        </div>

        <EmptyState
          v-else
          icon="mdi-account-heart-outline"
          title="No patients yet"
          description="Create a patient to exercise the gRPC service and propagate data through the gateway."
        >
          <template #actions>
            <router-link to="/patients/new">
              <v-btn color="primary">Create patient</v-btn>
            </router-link>
          </template>
        </EmptyState>

        <div v-if="filteredPatients.length > itemsPerPage" class="pagination">
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
import { ref, computed, onMounted } from 'vue';
import { usePatientStore } from '@/stores/patient';
import EmptyState from '@/components/ui/EmptyState.vue';
import SectionHeader from '@/components/ui/SectionHeader.vue';
import { formatDate, patientFullName, primaryTelecom } from '@/utils/formatters';

const patientStore = usePatientStore();
const searchQuery = ref('');
const filterGender = ref('all');
const currentPage = ref(1);
const itemsPerPage = 10;

const genderOptions = [
  { title: 'All genders', value: 'all' },
  { title: 'Female', value: 'female' },
  { title: 'Male', value: 'male' },
  { title: 'Other', value: 'other' },
  { title: 'Unknown', value: 'unknown' },
];

onMounted(async () => {
  if (!patientStore.patients.length) {
    await patientStore.fetchPatients();
  }
});

const filteredPatients = computed(() => {
  let dataset = patientStore.sortedPatients;
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    dataset = dataset.filter(patient => {
      const name = patientFullName(patient).toLowerCase();
      const contact = primaryTelecom(patient).toLowerCase();
      return name.includes(query) || contact.includes(query);
    });
  }
  if (filterGender.value !== 'all') {
    dataset = dataset.filter(patient => patient.gender === filterGender.value);
  }
  return dataset;
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(filteredPatients.value.length / itemsPerPage)),
);

const pagedPatients = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage;
  return filteredPatients.value.slice(start, start + itemsPerPage);
});

const capitalize = (value: string) => value.charAt(0).toUpperCase() + value.slice(1);

const deletePatient = async (id: string) => {
  if (!confirm('Delete this patient record?')) return;
  try {
    await patientStore.deletePatient(id);
  } catch (error) {
    console.error('Failed to delete patient:', error);
  }
};
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
  grid-template-columns: 1fr 220px;
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
  padding: 0 1.5rem 1.5rem;
}

thead tr {
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 0.75rem;
  color: var(--cf-text-muted);
}

tbody tr {
  border-bottom: 1px solid rgba(15, 23, 42, 0.06);
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
  .actions {
    justify-content: flex-start;
  }
  .table-wrapper {
    overflow-x: auto;
  }
}
</style>
