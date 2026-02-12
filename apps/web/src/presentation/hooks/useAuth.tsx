'use client'

import React, { createContext, useCallback, useContext, useState } from 'react'
import { AuthUser, LoginCredentials, RegisterInput } from '@/domain/entities/AuthUser'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { apiClient } from '@xenios/api-client'

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
  authRepo: AuthRepository
  tokenStorage: TokenStorage
}

export function AuthProvider({
  children,
  authRepo,
  tokenStorage,
}: AuthProviderProps) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const login = useCallback(
    async (credentials: LoginCredentials) => {
      setIsLoading(true)
      setError(null)
      try {
        const response = await authRepo.login(credentials)
        tokenStorage.setTokens(response.tokens)
        apiClient.setAuthToken(response.tokens.access_token)
        setUser(response.user)
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Login failed'
        setError(message)
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [authRepo, tokenStorage]
  )

  const register = useCallback(
    async (input: RegisterInput) => {
      setIsLoading(true)
      setError(null)
      try {
        const response = await authRepo.register(input)
        tokenStorage.setTokens(response.tokens)
        apiClient.setAuthToken(response.tokens.access_token)
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
    [authRepo, tokenStorage]
  )

  const logout = useCallback(async () => {
    setIsLoading(true)
    try {
      await authRepo.logout()
    } catch {
      // Logout should clear local state even if the API call fails
    } finally {
      tokenStorage.clearTokens()
      apiClient.clearAuthToken()
      setUser(null)
      setError(null)
      setIsLoading(false)
    }
  }, [authRepo, tokenStorage])

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
