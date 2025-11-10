<template>
  <v-app>
    <div class="app-shell" :class="{ 'app-shell--collapsed': !ui.sidebarOpen }">
      <aside class="app-shell__sidebar" :class="{ 'app-shell__sidebar--collapsed': !ui.sidebarOpen }">
        <div class="app-shell__brand">
          <div class="app-shell__brand-mark">
            <span>CF</span>
          </div>
          <div>
            <p class="app-shell__brand-title">CareFlow</p>
            <p class="app-shell__brand-subtitle">Unified patient journey</p>
          </div>
        </div>

        <nav class="app-shell__nav">
          <RouterLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="app-shell__nav-item"
            :class="{ 'app-shell__nav-item--active': isActive(item.path) }"
          >
            <v-icon size="20">{{ item.icon }}</v-icon>
            <span>{{ item.label }}</span>
          </RouterLink>
        </nav>

        <div class="app-shell__sidebar-footer">
          <p class="app-shell__sidebar-meta">{{ systemStatus }}</p>
          <v-btn icon variant="text" size="small" @click="toggleSidebar">
            <v-icon>{{ ui.sidebarOpen ? 'mdi-chevron-left' : 'mdi-chevron-right' }}</v-icon>
          </v-btn>
        </div>
      </aside>

      <section class="app-shell__content">
        <header class="app-shell__topbar">
          <div class="app-shell__topbar-left">
            <v-btn icon variant="text" class="lg-hidden" @click="toggleSidebar">
              <v-icon>mdi-menu</v-icon>
            </v-btn>
            <div>
              <p class="app-shell__eyebrow">CareFlow Mini</p>
              <h1 class="app-shell__page-title">{{ currentTitle }}</h1>
            </div>
          </div>

          <div class="app-shell__topbar-actions">
            <v-responsive class="app-shell__search" max-width="320">
              <v-text-field
                v-model="search"
                hide-details
                rounded="pill"
                density="compact"
                prepend-inner-icon="mdi-magnify"
                label="Search"
                variant="solo"
              />
            </v-responsive>

            <v-btn class="app-shell__status-chip" color="primary" variant="tonal" size="small">
              <v-icon start size="16">mdi-pulse</v-icon>
              Stack {{ systemStore.healthyCount }}/{{ systemStore.services.length || navItems.length }} ready
            </v-btn>

            <v-btn icon variant="text" @click="showNotifications = !showNotifications">
              <v-badge :content="ui.notifications.length" color="error" v-if="ui.notifications.length">
                <template #badge>
                  <span>{{ ui.notifications.length }}</span>
                </template>
                <v-icon>mdi-bell</v-icon>
              </v-badge>
              <template v-else>
                <v-icon>mdi-bell-outline</v-icon>
              </template>
            </v-btn>

            <v-menu offset-y>
              <template #activator="{ props }">
                <v-btn v-bind="props" class="app-shell__avatar" icon variant="flat">
                  <span>SB</span>
                </v-btn>
              </template>
              <v-list>
                <v-list-item title="Product Tour" prepend-icon="mdi-compass" />
                <v-list-item title="Settings" prepend-icon="mdi-cog" />
                <v-divider></v-divider>
                <v-list-item title="Sign out" prepend-icon="mdi-logout" />
              </v-list>
            </v-menu>
          </div>
        </header>

        <main class="app-shell__main">
          <router-view />
        </main>
      </section>
    </div>

    <v-navigation-drawer
      v-model="showNotifications"
      location="right"
      temporary
      width="360"
    >
      <v-card-title class="d-flex align-center justify-space-between">
        Notifications
        <v-btn icon variant="text" @click="showNotifications = false">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>
      <v-divider></v-divider>
      <div v-if="ui.notifications.length === 0" class="pa-4 text-center text-medium-emphasis">
        You're all caught up
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
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useUIStore } from '@/stores/ui'
import { useSystemStore } from '@/stores/system'
import AppNotificationItem from '@/components/AppNotificationItem.vue'

const ui = useUIStore()
const systemStore = useSystemStore()
const showNotifications = ref(false)
const route = useRoute()
const search = ref('')

const navItems = [
  { label: 'Dashboard', path: '/', icon: 'mdi-home-outline' },
  { label: 'Patients', path: '/patients', icon: 'mdi-account-heart-outline' },
  { label: 'Appointments', path: '/appointments', icon: 'mdi-calendar-clock' },
]

const isActive = (path: string) => route.path === path || route.path.startsWith(`${path}/`)
const currentTitle = computed(() => route.meta.title ?? 'Experience')
const systemStatus = computed(() => {
  if (!systemStore.services.length) return 'Stack probe pending'
  return `${systemStore.healthyCount}/${systemStore.services.length} services healthy`
})

const toggleSidebar = () => {
  ui.toggleSidebar()
}

onMounted(() => {
  if (!systemStore.services.length) {
    systemStore.refresh()
  }
})
</script>

<style scoped>
.app-shell {
  display: grid;
  grid-template-columns: 250px 1fr;
  min-height: 100vh;
  background: var(--cf-background);
}

.app-shell--collapsed {
  grid-template-columns: 110px 1fr;
}

.app-shell__sidebar {
  padding: 2rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
  border-right: 1px solid rgba(15, 23, 42, 0.05);
  background: linear-gradient(180deg, #fbfbfd 0%, #f6f7fb 100%);
}

.lg-hidden {
  display: none;
}

.app-shell__sidebar--collapsed {
  padding: 2rem 0.75rem;
}

.app-shell--collapsed .app-shell__nav-item {
  justify-content: center;
}

.app-shell--collapsed .app-shell__nav-item span,
.app-shell--collapsed .app-shell__brand-subtitle,
.app-shell--collapsed .app-shell__brand-title,
.app-shell--collapsed .app-shell__sidebar-meta {
  display: none;
}

.app-shell--collapsed .app-shell__brand {
  justify-content: center;
}

.app-shell__brand {
  display: flex;
  align-items: center;
  gap: 0.9rem;
}

.app-shell__brand-mark {
  width: 44px;
  height: 44px;
  border-radius: 16px;
  background: #0a84ff;
  color: #fff;
  display: grid;
  place-items: center;
  font-weight: 600;
}

.app-shell__brand-title {
  font-size: 1.2rem;
  font-weight: 600;
  margin: 0;
}

.app-shell__brand-subtitle {
  margin: 0;
  color: var(--cf-text-muted);
}

.app-shell__nav {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.app-shell__nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 14px;
  color: inherit;
  transition: background 0.2s ease;
}

.app-shell__nav-item:hover {
  background: rgba(10, 132, 255, 0.08);
}

.app-shell__nav-item--active {
  background: #fff;
  box-shadow: inset 0 0 0 1px rgba(10, 132, 255, 0.35);
}

.app-shell__sidebar-footer {
  margin-top: auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.85rem;
  color: var(--cf-text-muted);
}

.app-shell__content {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  padding: 2rem 2.5rem;
}

.app-shell__topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  gap: 1.5rem;
}

.app-shell__topbar-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.app-shell__eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--cf-text-muted);
  margin: 0;
  font-size: 0.75rem;
}

.app-shell__page-title {
  margin: 0;
  font-size: 1.8rem;
}

.app-shell__topbar-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.app-shell__search :deep(.v-field) {
  background: rgba(255, 255, 255, 0.9);
  box-shadow: inset 0 0 0 1px rgba(15, 23, 42, 0.05);
}

.app-shell__status-chip {
  font-weight: 600;
}

.app-shell__avatar {
  width: 42px;
  height: 42px;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 10px 25px rgba(15, 23, 42, 0.12);
  font-weight: 600;
}

.app-shell__main {
  flex: 1;
  background: #fff;
  border-radius: var(--cf-radius-lg);
  padding: 2.5rem;
  box-shadow: var(--cf-shadow-soft);
}

@media (max-width: 1080px) {
  .app-shell {
    grid-template-columns: 220px 1fr;
  }
}

@media (max-width: 900px) {
  .app-shell {
    grid-template-columns: 1fr;
  }
  .app-shell__sidebar {
    position: fixed;
    inset: 0 auto 0 0;
    transform: translateX(0);
    transition: transform 0.3s ease;
    z-index: 5;
  }
  .app-shell__sidebar--collapsed {
    transform: translateX(-100%);
  }
  .app-shell__topbar {
    flex-direction: column;
    align-items: flex-start;
  }
  .app-shell__topbar-actions {
    width: 100%;
    flex-wrap: wrap;
  }
  .lg-hidden {
    display: inline-flex;
  }
}
</style>
