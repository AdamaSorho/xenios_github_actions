import React from 'react'
import { renderHook, act } from '@testing-library/react'
import { AuthProvider, useAuth } from '@/presentation/hooks/useAuth'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { AuthTokenManager } from '@/domain/repositories/AuthTokenManager'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'
import { AuthRepository } from '@/domain/repositories/AuthRepository'

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
    getUser: jest.fn().mockReturnValue(null),
    setUser: jest.fn(),
  }
}

function createMockTokenManager(): jest.Mocked<AuthTokenManager> {
  return {
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
    restoreToken: jest.fn(),
  }
}

function createWrapper(
  mockAuthRepo: jest.Mocked<AuthRepository>,
  tokenStorage: TokenStorage,
  tokenManager: AuthTokenManager
) {
  const loginUseCase = new LoginUseCase(mockAuthRepo)
  const registerUseCase = new RegisterUseCase(mockAuthRepo)
  const logoutUseCase = new LogoutUseCase(mockAuthRepo)
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <AuthProvider
        loginUseCase={loginUseCase}
        registerUseCase={registerUseCase}
        logoutUseCase={logoutUseCase}
        tokenStorage={tokenStorage}
        tokenManager={tokenManager}
      >
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
    const mockTokenManager = createMockTokenManager()
    mockAuthRepo.login.mockRejectedValue('string error')

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage, mockTokenManager),
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
    const mockTokenManager = createMockTokenManager()
    mockAuthRepo.register.mockRejectedValue('string error')

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage, mockTokenManager),
    })

    await act(async () => {
      try {
        await result.current.register({
          email: 'a@b.com',
          password: 'password123',
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
    const mockTokenManager = createMockTokenManager()
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
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage, mockTokenManager),
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
    expect(mockTokenManager.clearAuthToken).toHaveBeenCalled()
  })
})
