import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { authStore } from '../stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('../features/auth/LoginView.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    name: 'home',
    component: () => import('../features/auth/HomeView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/onboarding',
    name: 'onboarding',
    component: () => import('../features/onboarding/OnboardingView.vue'),
    meta: { requiresAuth: true, requiresOwner: true },
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

/**
 * Global route guard:
 * - Unauthenticated access to a protected route → redirect to /login with
 *   `redirect` query preserving the target.
 * - Authenticated access to /login → redirect to /.
 * - Non-Owner access to an Owner-only route → redirect to /.
 *   (Does NOT call the backend; uses the in-memory role only.)
 */
router.beforeEach((to) => {
  const isAuthenticated = authStore.isAuthenticated.value

  if (to.meta.requiresAuth && !isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.public && isAuthenticated) {
    return { name: 'home' }
  }

  if (to.meta.requiresOwner && !authStore.isOwner.value) {
    return { name: 'home', query: { denied: 'owner' } }
  }

  return true
})

export type AppRouter = typeof router
