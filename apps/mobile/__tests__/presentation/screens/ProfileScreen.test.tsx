import React from 'react'
import { render, fireEvent } from '@testing-library/react-native'
import { ProfileScreen } from '@/presentation/screens/ProfileScreen'
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

describe('ProfileScreen', () => {
  it('should display user profile info', () => {
    const auth = createMockAuthContext()
    const { getByTestId } = render(
      <AuthContext.Provider value={auth}>
        <ProfileScreen />
      </AuthContext.Provider>
    )
    expect(getByTestId('profile-name').props.children).toBe('Test User')
    expect(getByTestId('profile-email').props.children).toBe('test@example.com')
    expect(getByTestId('profile-role').props.children).toBe('coach')
  })

  it('should display fallback when user is null', () => {
    const auth = createMockAuthContext({ user: null })
    const { getByTestId } = render(
      <AuthContext.Provider value={auth}>
        <ProfileScreen />
      </AuthContext.Provider>
    )
    expect(getByTestId('profile-name').props.children).toBe('Unknown')
    expect(getByTestId('profile-email').props.children).toBe('No email')
  })

  it('should call logout on sign out press', () => {
    const auth = createMockAuthContext()
    const { getByTestId } = render(
      <AuthContext.Provider value={auth}>
        <ProfileScreen />
      </AuthContext.Provider>
    )
    fireEvent.press(getByTestId('logout-button'))
    expect(auth.logout).toHaveBeenCalled()
  })
})
