import React from 'react'
import { render, fireEvent } from '@testing-library/react-native'
import { DashboardScreen } from '@/presentation/screens/DashboardScreen'
import { AuthContext, AuthContextValue } from '@/presentation/context/AuthContext'
import { AuthUser } from '@/domain/entities/AuthUser'

const mockUser: AuthUser = {
  id: 'user-1',
  email: 'test@example.com',
  name: 'Test User',
  role: 'coach',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

function createMockAuthContext(overrides: Partial<AuthContextValue> = {}): AuthContextValue {
  return {
    user: mockUser,
    isAuthenticated: true,
    isLoading: false,
    error: null,
    login: jest.fn().mockResolvedValue(undefined),
    register: jest.fn().mockResolvedValue(undefined),
    logout: jest.fn().mockResolvedValue(undefined),
    clearError: jest.fn(),
    ...overrides,
  }
}

describe('DashboardScreen', () => {
  it('should display user name', () => {
    const auth = createMockAuthContext()
    const { getByTestId } = render(
      <AuthContext.Provider value={auth}>
        <DashboardScreen />
      </AuthContext.Provider>
    )
    expect(getByTestId('greeting-text').props.children).toEqual([
      'Welcome, ',
      'Test User',
    ])
  })

  it('should display user role', () => {
    const auth = createMockAuthContext()
    const { getByTestId } = render(
      <AuthContext.Provider value={auth}>
        <DashboardScreen />
      </AuthContext.Provider>
    )
    expect(getByTestId('role-text').props.children).toEqual(['Role: ', 'coach'])
  })

  it('should display fallback when user is null', () => {
    const auth = createMockAuthContext({ user: null })
    const { getByTestId } = render(
      <AuthContext.Provider value={auth}>
        <DashboardScreen />
      </AuthContext.Provider>
    )
    expect(getByTestId('greeting-text').props.children).toEqual([
      'Welcome, ',
      'User',
    ])
  })

  it('should call logout on sign out press', () => {
    const auth = createMockAuthContext()
    const { getByTestId } = render(
      <AuthContext.Provider value={auth}>
        <DashboardScreen />
      </AuthContext.Provider>
    )
    fireEvent.press(getByTestId('logout-button'))
    expect(auth.logout).toHaveBeenCalled()
  })
})
