import React, { createContext, useContext, useState, useCallback, useEffect } from 'react'
import { AuthUser } from '@/domain/entities/AuthUser'
import { LoginCredentials, RegisterCredentials } from '@/domain/entities/AuthCredentials'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'
import { GetAuthStateUseCase } from '@/application/usecases/GetAuthStateUseCase'
import type { AuthClientConfigurator } from '@/domain/repositories/AuthClientConfigurator'

export interface AuthContextValue {
  user: AuthUser | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
  login: (credentials: LoginCredentials) => Promise<void>
  register: (credentials: RegisterCredentials) => Promise<void>
  logout: () => Promise<void>
  clearError: () => void
}

export const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export interface AuthProviderDeps {
  loginUseCase: LoginUseCase
  registerUseCase: RegisterUseCase
  logoutUseCase: LogoutUseCase
  getAuthStateUseCase: GetAuthStateUseCase
  authenticatedApiClient: AuthClientConfigurator
}

interface AuthProviderProps {
  deps: AuthProviderDeps
  children: React.ReactNode
}

export function AuthProvider({ deps, children }: AuthProviderProps) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let mounted = true

    async function checkAuth() {
      try {
        await deps.authenticatedApiClient.configureAuth()
        const state = await deps.getAuthStateUseCase.execute()
        if (mounted && state.isAuthenticated && state.user) {
          setUser(state.user)
        }
      } catch {
        // No stored auth state, user needs to login
      } finally {
        if (mounted) {
          setIsLoading(false)
        }
      }
    }

    checkAuth()
    return () => {
      mounted = false
    }
  }, [deps.getAuthStateUseCase, deps.authenticatedApiClient])

  const login = useCallback(
    async (credentials: LoginCredentials) => {
      setError(null)
      setIsLoading(true)
      try {
        const result = await deps.loginUseCase.execute(credentials)
        setUser(result.user)
        deps.authenticatedApiClient.setAccessToken(result.accessToken)
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Login failed'
        setError(message)
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [deps.loginUseCase, deps.authenticatedApiClient]
  )

  const register = useCallback(
    async (credentials: RegisterCredentials) => {
      setError(null)
      setIsLoading(true)
      try {
        const result = await deps.registerUseCase.execute(credentials)
        setUser(result.user)
        deps.authenticatedApiClient.setAccessToken(result.accessToken)
      } catch (err) {
        const message =
          err instanceof Error ? err.message : 'Registration failed'
        setError(message)
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [deps.registerUseCase, deps.authenticatedApiClient]
  )

  const logout = useCallback(async () => {
    setIsLoading(true)
    try {
      await deps.logoutUseCase.execute()
      setUser(null)
      deps.authenticatedApiClient.clearAuth()
    } finally {
      setIsLoading(false)
    }
  }, [deps.logoutUseCase, deps.authenticatedApiClient])

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  const value: AuthContextValue = {
    user,
    isAuthenticated: user !== null,
    isLoading,
    error,
    login,
    register,
    logout,
    clearError,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
