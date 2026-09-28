import { createRouter, createWebHistory } from 'vue-router'
import { debugEnabled } from '@/utils/debug'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/AdminLayout.vue'),
      children: [
        {
          path: '',
          name: 'children',
          component: () => import('@/features/children/ChildrenView.vue'),
        },
        {
          path: 'calendar/:month?',
          name: 'calendar',
          component: () => import('@/features/calendar/CalendarView.vue'),
        },
        {
          path: 'day/:date?',
          name: 'day',
          component: () => import('@/features/day/DayView.vue'),
        },
        {
          path: 'month/:month?',
          name: 'month',
          component: () => import('@/features/attendance/AttendanceView.vue'),
        },
        {
          path: 'attendance/:month?',
          redirect: (to) => ({ name: 'month', params: to.params }),
        },
        {
          path: 'debug',
          name: 'debug',
          component: () => import('@/features/debug/DebugView.vue'),
          beforeEnter: () => debugEnabled || { name: 'not-found', params: { pathMatch: ['debug'] } },
        },
      ],
    },
    {
      path: '/report/:uid',
      name: 'report',
      component: () => import('@/features/report/ReportView.vue'),
      props: true,
    },
    {
      path: '/report/:uid/child/:id',
      name: 'child',
      component: () => import('@/features/child/ChildView.vue'),
      props: true,
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
    },
  ],
})

export default router
