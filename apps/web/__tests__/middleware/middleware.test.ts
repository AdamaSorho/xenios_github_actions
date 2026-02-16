import { isProtectedRoute, isAuthRoute } from '@/middleware.helpers'

// Mock NextResponse
const mockRedirect = jest.fn().mockReturnValue({ type: 'redirect' })
const mockNext = jest.fn().mockReturnValue({ type: 'next' })

jest.mock('next/server', () => ({
  NextResponse: {
    redirect: (...args: unknown[]) => mockRedirect(...args),
    next: (...args: unknown[]) => mockNext(...args),
  },
}))

// Import middleware after mocking next/server
import { middleware, config } from '@/middleware'

function createMockRequest(pathname: string, cookieValue?: string) {
  const url = `http://localhost:3000${pathname}`
  const cookies = {
    get: jest.fn().mockImplementation((name: string) => {
      if (name === 'xenios_has_token' && cookieValue !== undefined) {
        return { value: cookieValue }
      }
      return undefined
    }),
  }
  return {
    nextUrl: { pathname },
    url,
    cookies,
  } as unknown as Parameters<typeof middleware>[0]
}

describe('middleware helpers', () => {
  describe('isProtectedRoute', () => {
    test('dashboard_IsProtected', () => {
      expect(isProtectedRoute('/dashboard')).toBe(true)
    })

    test('clients_IsProtected', () => {
      expect(isProtectedRoute('/clients')).toBe(true)
    })

    test('clientDetail_IsProtected', () => {
      expect(isProtectedRoute('/clients/abc-123')).toBe(true)
    })

    test('sessions_IsProtected', () => {
      expect(isProtectedRoute('/sessions')).toBe(true)
    })

    test('sessionDetail_IsProtected', () => {
      expect(isProtectedRoute('/sessions/abc-123')).toBe(true)
    })

    test('analytics_IsProtected', () => {
      expect(isProtectedRoute('/analytics')).toBe(true)
    })

    test('settings_IsProtected', () => {
      expect(isProtectedRoute('/settings')).toBe(true)
    })

    test('login_IsNotProtected', () => {
      expect(isProtectedRoute('/login')).toBe(false)
    })

    test('register_IsNotProtected', () => {
      expect(isProtectedRoute('/register')).toBe(false)
    })

    test('forgotPassword_IsNotProtected', () => {
      expect(isProtectedRoute('/forgot-password')).toBe(false)
    })

    test('root_IsNotProtected', () => {
      expect(isProtectedRoute('/')).toBe(false)
    })
  })

  describe('isAuthRoute', () => {
    test('login_IsAuthRoute', () => {
      expect(isAuthRoute('/login')).toBe(true)
    })

    test('register_IsAuthRoute', () => {
      expect(isAuthRoute('/register')).toBe(true)
    })

    test('forgotPassword_IsAuthRoute', () => {
      expect(isAuthRoute('/forgot-password')).toBe(true)
    })

    test('dashboard_IsNotAuthRoute', () => {
      expect(isAuthRoute('/dashboard')).toBe(false)
    })
  })
})

describe('middleware', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('protectedRoute_NoToken_RedirectsToLogin', () => {
    const request = createMockRequest('/dashboard')

    middleware(request)

    expect(mockRedirect).toHaveBeenCalledTimes(1)
    const redirectUrl = mockRedirect.mock.calls[0][0] as URL
    expect(redirectUrl.pathname).toBe('/login')
    expect(redirectUrl.searchParams.get('from')).toBe('/dashboard')
  })

  test('protectedRoute_WithToken_PassesThrough', () => {
    const request = createMockRequest('/dashboard', '1')

    middleware(request)

    expect(mockNext).toHaveBeenCalledTimes(1)
    expect(mockRedirect).not.toHaveBeenCalled()
  })

  test('authRoute_WithToken_RedirectsToDashboard', () => {
    const request = createMockRequest('/login', '1')

    middleware(request)

    expect(mockRedirect).toHaveBeenCalledTimes(1)
    const redirectUrl = mockRedirect.mock.calls[0][0] as URL
    expect(redirectUrl.pathname).toBe('/dashboard')
  })

  test('authRoute_NoToken_PassesThrough', () => {
    const request = createMockRequest('/login')

    middleware(request)

    expect(mockNext).toHaveBeenCalledTimes(1)
    expect(mockRedirect).not.toHaveBeenCalled()
  })

  test('nonProtectedNonAuth_NoToken_PassesThrough', () => {
    const request = createMockRequest('/')

    middleware(request)

    expect(mockNext).toHaveBeenCalledTimes(1)
    expect(mockRedirect).not.toHaveBeenCalled()
  })

  test('nestedProtectedRoute_NoToken_RedirectsWithFromParam', () => {
    const request = createMockRequest('/clients/abc-123')

    middleware(request)

    expect(mockRedirect).toHaveBeenCalledTimes(1)
    const redirectUrl = mockRedirect.mock.calls[0][0] as URL
    expect(redirectUrl.pathname).toBe('/login')
    expect(redirectUrl.searchParams.get('from')).toBe('/clients/abc-123')
  })

  test('protectedRoute_InvalidCookieValue_RedirectsToLogin', () => {
    const request = createMockRequest('/dashboard', 'invalid')

    middleware(request)

    expect(mockRedirect).toHaveBeenCalledTimes(1)
    const redirectUrl = mockRedirect.mock.calls[0][0] as URL
    expect(redirectUrl.pathname).toBe('/login')
  })
})

describe('middleware config', () => {
  test('matcher_IncludesProtectedRoutes', () => {
    expect(config.matcher).toContain('/dashboard/:path*')
    expect(config.matcher).toContain('/clients/:path*')
    expect(config.matcher).toContain('/sessions/:path*')
    expect(config.matcher).toContain('/analytics/:path*')
    expect(config.matcher).toContain('/settings/:path*')
  })

  test('matcher_IncludesAuthRoutes', () => {
    expect(config.matcher).toContain('/login')
    expect(config.matcher).toContain('/register')
    expect(config.matcher).toContain('/forgot-password')
  })
})
