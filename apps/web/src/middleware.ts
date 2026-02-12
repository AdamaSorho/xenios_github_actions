import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { isProtectedRoute, isAuthRoute } from './middleware.helpers'

const TOKEN_COOKIE_NAME = 'xenios_access_token'

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  const token = request.cookies.get(TOKEN_COOKIE_NAME)?.value

  // For localStorage-based auth, we check a simpler cookie flag
  // The actual JWT validation happens on API calls
  const hasToken = !!token

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
    '/clients/:path*',
    '/sessions/:path*',
    '/analytics/:path*',
    '/settings/:path*',
    '/login',
    '/register',
    '/forgot-password',
  ],
}
