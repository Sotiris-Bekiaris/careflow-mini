import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

// Lazy load views for code splitting
const Dashboard = () => import('@/views/Dashboard.vue')
const PatientList = () => import('@/views/PatientList.vue')
const PatientDetail = () => import('@/views/PatientDetail.vue')
const PatientForm = () => import('@/views/PatientForm.vue')
const AppointmentList = () => import('@/views/AppointmentList.vue')
const AppointmentForm = () => import('@/views/AppointmentForm.vue')
const NotFound = () => import('@/views/NotFound.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard,
    meta: {
      title: 'Dashboard',
    },
  },
  {
    path: '/patients',
    name: 'PatientList',
    component: PatientList,
    meta: {
      title: 'Patients',
    },
  },
  {
    path: '/patients/new',
    name: 'PatientCreate',
    component: PatientForm,
    meta: {
      title: 'Create Patient',
    },
  },
  {
    path: '/patients/:id',
    name: 'PatientDetail',
    component: PatientDetail,
    meta: {
      title: 'Patient Details',
    },
  },
  {
    path: '/patients/:id/edit',
    name: 'PatientEdit',
    component: PatientForm,
    meta: {
      title: 'Edit Patient',
    },
  },
  {
    path: '/appointments',
    name: 'AppointmentList',
    component: AppointmentList,
    meta: {
      title: 'Appointments',
    },
  },
  {
    path: '/appointments/new',
    name: 'AppointmentCreate',
    component: AppointmentForm,
    meta: {
      title: 'Create Appointment',
    },
  },
  {
    path: '/appointments/:id/edit',
    name: 'AppointmentEdit',
    component: AppointmentForm,
    meta: {
      title: 'Edit Appointment',
    },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFound,
    meta: {
      title: 'Page Not Found',
    },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// Update page title on route change
router.beforeEach((to, _from, next) => {
  const title = to.meta.title as string
  if (title) {
    document.title = `${title} | CareFlow`
  }
  next()
})

export default router
