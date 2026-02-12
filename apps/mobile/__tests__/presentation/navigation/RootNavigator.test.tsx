import React from 'react'
import { render } from '@testing-library/react-native'
import { RootNavigator } from '@/presentation/navigation/RootNavigator'
import { AuthContext, AuthContextValue } from '@/presentation/context/AuthContext'

function createMockAuthContext(overrides: Partial<AuthContextValue> = {}): AuthContextValue {
  return {
    user: null,
    isAuthenticated: false,
    isLoading: false,
    error: null,
    login: jest.fn().mockResolvedValue(undefined),
    register: jest.fn().mockResolvedValue(undefined),
    logout: jest.fn().mockResolvedValue(undefined),
    clearError: jest.fn(),
    ...overrides,
  }
}

function renderRootNavigator(authContext: AuthContextValue) {
  return render(
    <AuthContext.Provider value={authContext}>
      <RootNavigator />
    </AuthContext.Provider>
  )
}

describe('RootNavigator', () => {
  it('should show loading indicator when auth is loading', () => {
    const { getByTestId } = renderRootNavigator(
      createMockAuthContext({ isLoading: true })
    )

    expect(getByTestId('loading-indicator')).toBeTruthy()
  })

  it('should show auth navigator when not authenticated', () => {
    const { getByText } = renderRootNavigator(
      createMockAuthContext({ isAuthenticated: false })
    )

    expect(getByText('Welcome Back')).toBeTruthy()
  })

  it('should show main navigator when authenticated', () => {
    const { getByTestId } = renderRootNavigator(
      createMockAuthContext({
        isAuthenticated: true,
        user: {
          id: 'user-1',
          email: 'test@example.com',
          name: 'Test User',
          role: 'coach',
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-01T00:00:00Z',
        },
      })
    )

    expect(getByTestId('tab-dashboard')).toBeTruthy()
    expect(getByTestId('tab-clients')).toBeTruthy()
    expect(getByTestId('tab-sessions')).toBeTruthy()
    expect(getByTestId('tab-profile')).toBeTruthy()
  })
})
