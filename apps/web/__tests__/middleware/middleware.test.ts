import { isProtectedRoute, isAuthRoute } from '@/middleware.helpers'

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
