import type { RouteLocationNormalizedLoaded } from 'vue-router'

export function routeParam(route: RouteLocationNormalizedLoaded, name: string): string {
  const value = route.params[name]
  return typeof value === 'string' ? value : ''
}
