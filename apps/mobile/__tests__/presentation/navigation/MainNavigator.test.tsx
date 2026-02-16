import React from 'react'
import { render, fireEvent } from '@testing-library/react-native'
import { MainNavigator } from '@/presentation/navigation/MainNavigator'
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

function createMockAuthContext(): AuthContextValue {
  return {
    user: mockUser,
    isAuthenticated: true,
    isLoading: false,
    error: null,
    login: jest.fn().mockResolvedValue(undefined),
    register: jest.fn().mockResolvedValue(undefined),
    logout: jest.fn().mockResolvedValue(undefined),
    clearError: jest.fn(),
  }
}

function renderMainNavigator() {
  return render(
    <AuthContext.Provider value={createMockAuthContext()}>
      <MainNavigator />
    </AuthContext.Provider>
  )
}

describe('MainNavigator', () => {
  it('should show dashboard tab by default', () => {
    const { getByTestId } = renderMainNavigator()
    expect(getByTestId('greeting-text')).toBeTruthy()
  })

  it('should switch to clients tab', () => {
    const { getByTestId, getByText } = renderMainNavigator()
    fireEvent.press(getByTestId('tab-clients'))
    expect(getByText('Manage your clients')).toBeTruthy()
  })

  it('should switch to sessions tab', () => {
    const { getByTestId, getByText } = renderMainNavigator()
    fireEvent.press(getByTestId('tab-sessions'))
    expect(getByText('Your training sessions')).toBeTruthy()
  })

  it('should switch to profile tab', () => {
    const { getByTestId } = renderMainNavigator()
    fireEvent.press(getByTestId('tab-profile'))
    expect(getByTestId('profile-name')).toBeTruthy()
  })

  it('should show all four tab buttons', () => {
    const { getByTestId } = renderMainNavigator()
    expect(getByTestId('tab-dashboard')).toBeTruthy()
    expect(getByTestId('tab-clients')).toBeTruthy()
    expect(getByTestId('tab-sessions')).toBeTruthy()
    expect(getByTestId('tab-profile')).toBeTruthy()
  })
})
