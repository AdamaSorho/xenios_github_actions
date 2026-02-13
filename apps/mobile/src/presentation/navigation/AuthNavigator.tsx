import React, { useState } from 'react'
import { LoginScreen } from '@/presentation/screens/LoginScreen'
import { RegisterScreen } from '@/presentation/screens/RegisterScreen'

type AuthScreen = 'login' | 'register'

export function AuthNavigator() {
  const [screen, setScreen] = useState<AuthScreen>('login')

  if (screen === 'register') {
    return <RegisterScreen onNavigateToLogin={() => setScreen('login')} />
  }

  return <LoginScreen onNavigateToRegister={() => setScreen('register')} />
}
