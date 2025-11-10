import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { UINotification } from './types'

let notificationId = 0

export const useUIStore = defineStore('ui', () => {
  // State
  const notifications = ref<UINotification[]>([])
  const isDarkMode = ref(false)
  const sidebarOpen = ref(true)

  // Actions - Notifications
  const addNotification = (
    message: string,
    type: 'success' | 'error' | 'warning' | 'info' = 'info',
    duration = 5000,
  ) => {
    const id = String(notificationId++)
    const notification: UINotification = {
      id,
      message,
      type,
      duration,
    }

    notifications.value.push(notification)

    if (duration && duration > 0) {
      setTimeout(() => {
        removeNotification(id)
      }, duration)
    }

    return id
  }

  const removeNotification = (id: string) => {
    notifications.value = notifications.value.filter(n => n.id !== id)
  }

  const clearNotifications = () => {
    notifications.value = []
  }

  // Actions - Theme
  const toggleDarkMode = () => {
    isDarkMode.value = !isDarkMode.value
  }

  const setDarkMode = (dark: boolean) => {
    isDarkMode.value = dark
  }

  // Actions - Sidebar
  const toggleSidebar = () => {
    sidebarOpen.value = !sidebarOpen.value
  }

  const setSidebarOpen = (open: boolean) => {
    sidebarOpen.value = open
  }

  return {
    // State
    notifications,
    isDarkMode,
    sidebarOpen,
    // Actions
    addNotification,
    removeNotification,
    clearNotifications,
    toggleDarkMode,
    setDarkMode,
    toggleSidebar,
    setSidebarOpen,
  }
})
