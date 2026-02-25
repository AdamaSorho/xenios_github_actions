const PROTECTED_PREFIXES = [
  '/dashboard',
  '/insights',
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

