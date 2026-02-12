import React from 'react'
import { render, fireEvent } from '@testing-library/react-native'
import { AuthNavigator } from '@/presentation/navigation/AuthNavigator'
import { AuthContext, AuthContextValue } from '@/presentation/context/AuthContext'

function createMockAuthContext(): AuthContextValue {
  return {
    user: null,
    isAuthenticated: false,
    isLoading: false,
    error: null,
    login: jest.fn().mockResolvedValue(undefined),
    register: jest.fn().mockResolvedValue(undefined),
    logout: jest.fn().mockResolvedValue(undefined),
    clearError: jest.fn(),
  }
}

function renderAuthNavigator() {
  return render(
    <AuthContext.Provider value={createMockAuthContext()}>
      <AuthNavigator />
    </AuthContext.Provider>
  )
}

describe('AuthNavigator', () => {
  it('should show login screen by default', () => {
    const { getByText } = renderAuthNavigator()
    expect(getByText('Welcome Back')).toBeTruthy()
  })

  it('should navigate to register screen', () => {
    const { getByTestId, getByText } = renderAuthNavigator()
    fireEvent.press(getByTestId('register-link'))
    expect(getByText('Join Xenios today')).toBeTruthy()
  })

  it('should navigate back to login from register', () => {
    const { getByTestId, getByText } = renderAuthNavigator()

    // Go to register
    fireEvent.press(getByTestId('register-link'))
    expect(getByText('Join Xenios today')).toBeTruthy()

    // Go back to login
    fireEvent.press(getByTestId('login-link'))
    expect(getByText('Welcome Back')).toBeTruthy()
  })
})
