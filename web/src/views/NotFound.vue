<template>
  <div class="pa-6 min-h-screen flex items-center justify-center">
    <v-card class="max-w-md w-full text-center">
      <v-card-text class="pt-12 pb-12">
        <v-icon size="120" color="primary" class="mb-6">mdi-alert-circle</v-icon>

        <h1 class="text-4xl font-bold text-gray-800 mb-2">404</h1>
        <p class="text-xl text-gray-600 mb-6">Page Not Found</p>
        <p class="text-gray-600 mb-8">
          The page you're looking for doesn't exist or has been moved.
        </p>

        <v-row>
          <v-col cols="12">
            <router-link to="/">
              <v-btn color="primary" size="large" class="w-full" prepend-icon="mdi-home">
                Back to Home
              </v-btn>
            </router-link>
          </v-col>

          <v-col cols="12">
            <router-link to="/patients">
              <v-btn color="secondary" size="large" variant="outlined" class="w-full">
                View Patients
              </v-btn>
            </router-link>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup lang="ts">
// Placeholder for 404 page
</script>
