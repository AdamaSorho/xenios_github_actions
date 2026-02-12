import React from 'react'
import { renderHook, act } from '@testing-library/react'
import { AuthProvider, useAuth } from '@/presentation/hooks/useAuth'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorage } from '@/domain/repositories/TokenStorage'

function createMockAuthRepo(): jest.Mocked<AuthRepository> {
  return {
    login: jest.fn(),
    register: jest.fn(),
    refresh: jest.fn(),
    logout: jest.fn(),
  }
}

function createMockTokenStorage(): jest.Mocked<TokenStorage> {
  return {
    getAccessToken: jest.fn().mockReturnValue(null),
    getRefreshToken: jest.fn().mockReturnValue(null),
    setTokens: jest.fn(),
    clearTokens: jest.fn(),
  }
}

function createWrapper(
  authRepo: AuthRepository,
  tokenStorage: TokenStorage
) {
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <AuthProvider authRepo={authRepo} tokenStorage={tokenStorage}>
        {children}
      </AuthProvider>
    )
  }
}

describe('useAuth branch coverage', () => {
  test('useAuth_WithoutProvider_ThrowsError', () => {
    expect(() => {
      renderHook(() => useAuth())
    }).toThrow('useAuth must be used within an AuthProvider')
  })

  test('login_NonErrorObject_SetsGenericMessage', async () => {
    const mockAuthRepo = createMockAuthRepo()
    const mockTokenStorage = createMockTokenStorage()
    mockAuthRepo.login.mockRejectedValue('string error')

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage),
    })

    await act(async () => {
      try {
        await result.current.login({
          email: 'a@b.com',
          password: 'pass',
        })
      } catch {
        // expected
      }
    })

    expect(result.current.error).toBe('Login failed')
  })

  test('register_NonErrorObject_SetsGenericMessage', async () => {
    const mockAuthRepo = createMockAuthRepo()
    const mockTokenStorage = createMockTokenStorage()
    mockAuthRepo.register.mockRejectedValue('string error')

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage),
    })

    await act(async () => {
      try {
        await result.current.register({
          email: 'a@b.com',
          password: 'pass',
          name: 'Test',
          role: 'coach',
        })
      } catch {
        // expected
      }
    })

    expect(result.current.error).toBe('Registration failed')
  })

  test('logout_ApiError_StillClearsLocalState', async () => {
    const mockAuthRepo = createMockAuthRepo()
    const mockTokenStorage = createMockTokenStorage()
    mockAuthRepo.login.mockResolvedValue({
      user: {
        id: 'user-1',
        email: 'a@b.com',
        name: 'Test',
        role: 'coach',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      tokens: {
        access_token: 'at',
        refresh_token: 'rt',
      },
    })
    mockAuthRepo.logout.mockRejectedValue(new Error('Network error'))

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage),
    })

    // Login first
    await act(async () => {
      await result.current.login({ email: 'a@b.com', password: 'pass' })
    })

    expect(result.current.isAuthenticated).toBe(true)

    // Logout should still clear local state even if API fails
    await act(async () => {
      await result.current.logout()
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBeNull()
    expect(mockTokenStorage.clearTokens).toHaveBeenCalled()
  })
})
