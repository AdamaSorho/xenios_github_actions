const PROTECTED_PREFIXES = [
  '/dashboard',
  '/clients',
  '/sessions',
  '/analytics',
  '/settings',
]

const AUTH_ROUTES = ['/login', '/register', '/forgot-password']

export function isProtectedRoute(pathname: string): boolean {
  return PROTECTED_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(prefix + '/')
  )
}

export function isAuthRoute(pathname: string): boolean {
  return AUTH_ROUTES.includes(pathname)
}

export function getRedirectUrl(
  pathname: string,
  isAuthenticated: boolean
): string | null {
  if (isProtectedRoute(pathname) && !isAuthenticated) {
    return '/login'
  }

  if (isAuthRoute(pathname) && isAuthenticated) {
    return '/dashboard'
  }

  return null
}
