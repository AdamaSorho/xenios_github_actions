'use client'

import { ReactNode, useEffect } from 'react'
import { AuthProvider } from '@/presentation/hooks/useAuth'
import {
  tokenStorage,
  tokenManager,
  loginUseCase,
  registerUseCase,
  logoutUseCase,
} from '@/infrastructure/container'

export function Providers({ children }: { children: ReactNode }) {
  useEffect(() => {
    const storedToken = tokenStorage.getAccessToken()
    tokenManager.restoreToken(storedToken)
  }, [])

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
