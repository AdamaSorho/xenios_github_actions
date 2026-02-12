'use client'

import React, { createContext, useCallback, useContext, useState } from 'react'
import { AuthUser, LoginCredentials, RegisterInput } from '@/domain/entities/AuthUser'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { AuthTokenManager } from '@/domain/repositories/AuthTokenManager'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'

interface AuthContextValue {
  user: AuthUser | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
  login: (credentials: LoginCredentials) => Promise<void>
  register: (input: RegisterInput) => Promise<void>
  logout: () => Promise<void>
  clearError: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

interface AuthProviderProps {
  children: React.ReactNode
  loginUseCase: LoginUseCase
  registerUseCase: RegisterUseCase
  logoutUseCase: LogoutUseCase
  tokenStorage: TokenStorage
  tokenManager: AuthTokenManager
}

export function AuthProvider({
  children,
  loginUseCase,
  registerUseCase,
  logoutUseCase,
  tokenStorage,
  tokenManager,
}: AuthProviderProps) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const login = useCallback(
    async (credentials: LoginCredentials) => {
      setIsLoading(true)
      setError(null)
      try {
        const response = await loginUseCase.execute(credentials)
        tokenStorage.setTokens(response.tokens)
        tokenManager.setAuthToken(response.tokens.access_token)
        setUser(response.user)
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Login failed'
        setError(message)
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [loginUseCase, tokenStorage, tokenManager]
  )

  const register = useCallback(
    async (input: RegisterInput) => {
      setIsLoading(true)
      setError(null)
      try {
        const response = await registerUseCase.execute(input)
        tokenStorage.setTokens(response.tokens)
        tokenManager.setAuthToken(response.tokens.access_token)
        setUser(response.user)
      } catch (err) {
        const message =
          err instanceof Error ? err.message : 'Registration failed'
        setError(message)
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [registerUseCase, tokenStorage, tokenManager]
  )

  const logout = useCallback(async () => {
    setIsLoading(true)
    try {
      await logoutUseCase.execute()
    } catch {
      // Logout should clear local state even if the API call fails
    } finally {
      tokenStorage.clearTokens()
      tokenManager.clearAuthToken()
      setUser(null)
      setError(null)
      setIsLoading(false)
    }
  }, [logoutUseCase, tokenStorage, tokenManager])

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: user !== null,
        isLoading,
        error,
        login,
        register,
        logout,
        clearError,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
