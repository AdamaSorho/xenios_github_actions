import React from 'react'
import { render, fireEvent, waitFor } from '@testing-library/react-native'
import { LoginScreen } from '@/presentation/screens/LoginScreen'
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

function renderLoginScreen(
  authContext: AuthContextValue = createMockAuthContext(),
  onNavigateToRegister: () => void = jest.fn()
) {
  return render(
    <AuthContext.Provider value={authContext}>
      <LoginScreen onNavigateToRegister={onNavigateToRegister} />
    </AuthContext.Provider>
  )
}

describe('LoginScreen', () => {
  it('should render login form', () => {
    const { getByTestId, getByText } = renderLoginScreen()

    expect(getByText('Welcome Back')).toBeTruthy()
    expect(getByTestId('email-input')).toBeTruthy()
    expect(getByTestId('password-input')).toBeTruthy()
    expect(getByTestId('login-button')).toBeTruthy()
  })

  it('should show validation errors for empty fields', async () => {
    const auth = createMockAuthContext()
    const { getByTestId } = renderLoginScreen(auth)

    fireEvent.press(getByTestId('login-button'))

    await waitFor(() => {
      expect(getByTestId('error-container')).toBeTruthy()
    })

    expect(auth.login).not.toHaveBeenCalled()
  })

  it('should show validation error for invalid email', async () => {
    const auth = createMockAuthContext()
    const { getByTestId } = renderLoginScreen(auth)

    fireEvent.changeText(getByTestId('email-input'), 'invalid-email')
    fireEvent.changeText(getByTestId('password-input'), 'password123')
    fireEvent.press(getByTestId('login-button'))

    await waitFor(() => {
      expect(getByTestId('error-container')).toBeTruthy()
    })

    expect(auth.login).not.toHaveBeenCalled()
  })

  it('should call login with valid credentials', async () => {
    const auth = createMockAuthContext()
    const { getByTestId } = renderLoginScreen(auth)

    fireEvent.changeText(getByTestId('email-input'), 'test@example.com')
    fireEvent.changeText(getByTestId('password-input'), 'password123')
    fireEvent.press(getByTestId('login-button'))

    await waitFor(() => {
      expect(auth.login).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password123',
      })
    })
  })

  it('should show server error from auth context', () => {
    const auth = createMockAuthContext({ error: 'Invalid credentials' })
    const { getByTestId } = renderLoginScreen(auth)

    expect(getByTestId('error-container')).toBeTruthy()
  })

  it('should show loading state', () => {
    const auth = createMockAuthContext({ isLoading: true })
    const { getByTestId } = renderLoginScreen(auth)

    expect(getByTestId('login-button').props.accessibilityState.disabled).toBe(
      true
    )
  })

  it('should navigate to register screen', () => {
    const onNavigateToRegister = jest.fn()
    const { getByTestId } = renderLoginScreen(
      createMockAuthContext(),
      onNavigateToRegister
    )

    fireEvent.press(getByTestId('register-link'))

    expect(onNavigateToRegister).toHaveBeenCalled()
  })

  it('should clear error on login attempt', async () => {
    const auth = createMockAuthContext({ error: 'Previous error' })
    const { getByTestId } = renderLoginScreen(auth)

    fireEvent.changeText(getByTestId('email-input'), 'test@example.com')
    fireEvent.changeText(getByTestId('password-input'), 'password123')
    fireEvent.press(getByTestId('login-button'))

    expect(auth.clearError).toHaveBeenCalled()
  })
})
