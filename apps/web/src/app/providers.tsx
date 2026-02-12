'use client'

import { ReactNode } from 'react'
import { AuthProvider } from '@/presentation/hooks/useAuth'
import { ApiAuthRepository } from '@/infrastructure/repositories/ApiAuthRepository'
import { LocalTokenStorage } from '@/infrastructure/auth/LocalTokenStorage'
import { apiClient } from '@xenios/api-client'

const authRepo = new ApiAuthRepository()
const tokenStorage = new LocalTokenStorage()

// Restore auth token from storage on app load
if (typeof window !== 'undefined') {
  const storedToken = tokenStorage.getAccessToken()
  if (storedToken) {
    apiClient.setAuthToken(storedToken)
  }
}

export function Providers({ children }: { children: ReactNode }) {
  return (
    <AuthProvider authRepo={authRepo} tokenStorage={tokenStorage}>
      {children}
    </AuthProvider>
  )
}
