import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { isProtectedRoute, isAuthRoute } from './middleware.helpers'

/**
 * Cookie flag name set by LocalTokenStorage when tokens are stored.
 * This is NOT the JWT itself — it's a simple "1" flag indicating
 * the user has an active session. Actual JWT validation happens
 * server-side on each API call.
 */
const AUTH_COOKIE_NAME = 'xenios_has_token'

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  const hasToken = request.cookies.get(AUTH_COOKIE_NAME)?.value === '1'

  if (isProtectedRoute(pathname) && !hasToken) {
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('from', pathname)
    return NextResponse.redirect(loginUrl)
  }

  if (isAuthRoute(pathname) && hasToken) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/dashboard/:path*',
    '/insights/:path*',
    '/clients/:path*',
    '/sessions/:path*',
    '/analytics/:path*',
    '/settings/:path*',
    '/login',
    '/register',
    '/forgot-password',
  ],
}
