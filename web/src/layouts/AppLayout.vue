<template>
  <v-app>
    <!-- Sidebar Navigation -->
    <v-navigation-drawer
      v-model="ui.sidebarOpen"
      permanent
      :width="260"
      class="bg-surface"
    >
      <div class="pa-4">
        <h2 class="text-xl font-bold text-primary">CareFlow</h2>
        <p class="text-xs text-gray-600 mt-1">Healthcare Management</p>
      </div>

      <v-divider></v-divider>

      <v-list>
        <v-list-item
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          :title="item.label"
          :prepend-icon="item.icon"
          active-color="primary"
        />
      </v-list>

      <v-divider></v-divider>

      <div class="pa-4 border-t">
        <v-btn
          variant="text"
          size="small"
          icon
          @click="ui.toggleDarkMode"
          class="mr-2"
        >
          <v-icon>{{ ui.isDarkMode ? 'mdi-white-balance-sunny' : 'mdi-moon-waning-crescent' }}</v-icon>
        </v-btn>
        <span class="text-xs text-gray-600">Theme</span>
      </div>
    </v-navigation-drawer>

    <!-- Top App Bar -->
    <v-app-bar
      color="primary"
      dark
      class="px-4"
    >
      <v-app-bar-nav-icon @click="ui.toggleSidebar" />
      <v-toolbar-title class="font-bold">CareFlow Dashboard</v-toolbar-title>
      <v-spacer></v-spacer>

      <!-- Notifications Icon -->
      <v-badge :content="ui.notifications.length" color="error">
        <v-btn icon variant="text" @click="showNotifications = !showNotifications">
          <v-icon>mdi-bell</v-icon>
        </v-btn>
      </v-badge>

      <!-- User Menu -->
      <v-menu>
        <template v-slot:activator="{ props }">
          <v-btn icon v-bind="props" variant="text">
            <v-icon>mdi-account-circle</v-icon>
          </v-btn>
        </template>

        <v-list>
          <v-list-item title="Profile" prepend-icon="mdi-account" />
          <v-list-item title="Settings" prepend-icon="mdi-cog" />
          <v-divider></v-divider>
          <v-list-item title="Logout" prepend-icon="mdi-logout" />
        </v-list>
      </v-menu>
    </v-app-bar>

    <!-- Notifications Panel -->
    <v-navigation-drawer v-model="showNotifications" location="right" temporary width="350">
      <v-card-title class="pa-4">Notifications</v-card-title>
      <v-divider></v-divider>

      <div v-if="ui.notifications.length === 0" class="pa-4 text-center text-gray-600">
        No notifications
      </div>

      <div v-else>
        <AppNotificationItem
          v-for="notification in ui.notifications"
          :key="notification.id"
          :notification="notification"
          @close="ui.removeNotification(notification.id)"
        />
      </div>
    </v-navigation-drawer>

    <!-- Main Content -->
    <v-main class="bg-gray-50">
      <router-view />
    </v-main>

    <!-- Snackbar for notifications -->
    <v-snackbar
      v-for="notification in ui.notifications"
      :key="notification.id"
      :model-value="true"
      :color="notification.type"
      :timeout="notification.duration"
      @update:model-value="ui.removeNotification(notification.id)"
    >
      {{ notification.message }}
    </v-snackbar>
  </v-app>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useUIStore } from '@/stores/ui'
import AppNotificationItem from '@/components/AppNotificationItem.vue'

const ui = useUIStore()
const showNotifications = ref(false)

const navItems = [
  { label: 'Dashboard', path: '/', icon: 'mdi-home' },
  { label: 'Patients', path: '/patients', icon: 'mdi-hospital-box' },
  { label: 'Appointments', path: '/appointments', icon: 'mdi-calendar-check' },
]
</script>

<style scoped>
:deep(.v-navigation-drawer__content) {
  overflow-y: auto;
}
</style>
