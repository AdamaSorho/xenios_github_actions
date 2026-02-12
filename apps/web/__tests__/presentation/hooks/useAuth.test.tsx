import React from 'react'
import { renderHook, act } from '@testing-library/react'
import { AuthProvider, useAuth } from '@/presentation/hooks/useAuth'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { AuthResponse } from '@/domain/entities/AuthUser'

const mockAuthResponse: AuthResponse = {
  user: {
    id: 'user-1',
    email: 'coach@example.com',
    name: 'Test Coach',
    role: 'coach',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  tokens: {
    access_token: 'access-token-123',
    refresh_token: 'refresh-token-456',
  },
}

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

describe('useAuth', () => {
  let mockAuthRepo: jest.Mocked<AuthRepository>
  let mockTokenStorage: jest.Mocked<TokenStorage>

  beforeEach(() => {
    mockAuthRepo = createMockAuthRepo()
    mockTokenStorage = createMockTokenStorage()
  })

  test('initialState_NoStoredToken_IsUnauthenticated', () => {
    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage),
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.isLoading).toBe(false)
  })

  test('login_ValidCredentials_SetsUserAndTokens', async () => {
    mockAuthRepo.login.mockResolvedValue(mockAuthResponse)

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage),
    })

    await act(async () => {
      await result.current.login({
        email: 'coach@example.com',
        password: 'password123',
      })
    })

    expect(result.current.user).toEqual(mockAuthResponse.user)
    expect(result.current.isAuthenticated).toBe(true)
    expect(mockTokenStorage.setTokens).toHaveBeenCalledWith(
      mockAuthResponse.tokens
    )
  })

  test('login_InvalidCredentials_SetsError', async () => {
    mockAuthRepo.login.mockRejectedValue(new Error('Invalid credentials'))

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage),
    })

    await act(async () => {
      try {
        await result.current.login({
          email: 'coach@example.com',
          password: 'wrong',
        })
      } catch {
        // expected
      }
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.error).toBe('Invalid credentials')
  })

  test('register_ValidInput_SetsUserAndTokens', async () => {
    mockAuthRepo.register.mockResolvedValue(mockAuthResponse)

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage),
    })

    await act(async () => {
      await result.current.register({
        email: 'coach@example.com',
        password: 'Password123!',
        name: 'Test Coach',
        role: 'coach',
      })
    })

    expect(result.current.user).toEqual(mockAuthResponse.user)
    expect(result.current.isAuthenticated).toBe(true)
    expect(mockTokenStorage.setTokens).toHaveBeenCalledWith(
      mockAuthResponse.tokens
    )
  })

  test('logout_ClearsUserAndTokens', async () => {
    mockAuthRepo.login.mockResolvedValue(mockAuthResponse)
    mockAuthRepo.logout.mockResolvedValue(undefined)

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage),
    })

    await act(async () => {
      await result.current.login({
        email: 'coach@example.com',
        password: 'password123',
      })
    })

    await act(async () => {
      await result.current.logout()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(mockTokenStorage.clearTokens).toHaveBeenCalled()
  })

  test('clearError_RemovesError', async () => {
    mockAuthRepo.login.mockRejectedValue(new Error('Invalid credentials'))

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(mockAuthRepo, mockTokenStorage),
    })

    await act(async () => {
      try {
        await result.current.login({
          email: 'coach@example.com',
          password: 'wrong',
        })
      } catch {
        // expected
      }
    })

    expect(result.current.error).toBe('Invalid credentials')

    act(() => {
      result.current.clearError()
    })

    expect(result.current.error).toBeNull()
  })
})
