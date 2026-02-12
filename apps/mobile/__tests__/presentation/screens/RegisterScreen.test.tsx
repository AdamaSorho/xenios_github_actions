import React from 'react'
import { render, fireEvent, waitFor } from '@testing-library/react-native'
import { RegisterScreen } from '@/presentation/screens/RegisterScreen'
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

function renderRegisterScreen(
  authContext: AuthContextValue = createMockAuthContext(),
  onNavigateToLogin: () => void = jest.fn()
) {
  return render(
    <AuthContext.Provider value={authContext}>
      <RegisterScreen onNavigateToLogin={onNavigateToLogin} />
    </AuthContext.Provider>
  )
}

describe('RegisterScreen', () => {
  it('should render registration form', () => {
    const { getByTestId, getByText } = renderRegisterScreen()

    expect(getByText('Join Xenios today')).toBeTruthy()
    expect(getByTestId('name-input')).toBeTruthy()
    expect(getByTestId('email-input')).toBeTruthy()
    expect(getByTestId('password-input')).toBeTruthy()
    expect(getByTestId('role-coach')).toBeTruthy()
    expect(getByTestId('role-client')).toBeTruthy()
    expect(getByTestId('register-button')).toBeTruthy()
  })

  it('should show validation errors for empty fields', async () => {
    const auth = createMockAuthContext()
    const { getByTestId } = renderRegisterScreen(auth)

    // Clear the default name field value
    fireEvent.changeText(getByTestId('name-input'), '')
    fireEvent.press(getByTestId('register-button'))

    await waitFor(() => {
      expect(getByTestId('error-container')).toBeTruthy()
    })

    expect(auth.register).not.toHaveBeenCalled()
  })

  it('should show validation error for short password', async () => {
    const auth = createMockAuthContext()
    const { getByTestId } = renderRegisterScreen(auth)

    fireEvent.changeText(getByTestId('name-input'), 'John Doe')
    fireEvent.changeText(getByTestId('email-input'), 'john@example.com')
    fireEvent.changeText(getByTestId('password-input'), 'short')
    fireEvent.press(getByTestId('register-button'))

    await waitFor(() => {
      expect(getByTestId('error-container')).toBeTruthy()
    })

    expect(auth.register).not.toHaveBeenCalled()
  })

  it('should call register with valid credentials', async () => {
    const auth = createMockAuthContext()
    const { getByTestId } = renderRegisterScreen(auth)

    fireEvent.changeText(getByTestId('name-input'), 'John Doe')
    fireEvent.changeText(getByTestId('email-input'), 'john@example.com')
    fireEvent.changeText(getByTestId('password-input'), 'password123')
    fireEvent.press(getByTestId('role-coach'))
    fireEvent.press(getByTestId('register-button'))

    await waitFor(() => {
      expect(auth.register).toHaveBeenCalledWith({
        email: 'john@example.com',
        password: 'password123',
        name: 'John Doe',
        role: 'coach',
      })
    })
  })

  it('should default to client role', async () => {
    const auth = createMockAuthContext()
    const { getByTestId } = renderRegisterScreen(auth)

    fireEvent.changeText(getByTestId('name-input'), 'Jane Doe')
    fireEvent.changeText(getByTestId('email-input'), 'jane@example.com')
    fireEvent.changeText(getByTestId('password-input'), 'password123')
    fireEvent.press(getByTestId('register-button'))

    await waitFor(() => {
      expect(auth.register).toHaveBeenCalledWith(
        expect.objectContaining({ role: 'client' })
      )
    })
  })

  it('should allow switching role to coach', async () => {
    const auth = createMockAuthContext()
    const { getByTestId } = renderRegisterScreen(auth)

    fireEvent.press(getByTestId('role-coach'))
    fireEvent.changeText(getByTestId('name-input'), 'Jane')
    fireEvent.changeText(getByTestId('email-input'), 'jane@example.com')
    fireEvent.changeText(getByTestId('password-input'), 'password123')
    fireEvent.press(getByTestId('register-button'))

    await waitFor(() => {
      expect(auth.register).toHaveBeenCalledWith(
        expect.objectContaining({ role: 'coach' })
      )
    })
  })

  it('should show server error from auth context', () => {
    const auth = createMockAuthContext({ error: 'Email already exists' })
    const { getByTestId } = renderRegisterScreen(auth)

    expect(getByTestId('error-container')).toBeTruthy()
  })

  it('should show loading state', () => {
    const auth = createMockAuthContext({ isLoading: true })
    const { getByTestId } = renderRegisterScreen(auth)

    expect(
      getByTestId('register-button').props.accessibilityState.disabled
    ).toBe(true)
  })

  it('should navigate to login screen', () => {
    const onNavigateToLogin = jest.fn()
    const { getByTestId } = renderRegisterScreen(
      createMockAuthContext(),
      onNavigateToLogin
    )

    fireEvent.press(getByTestId('login-link'))

    expect(onNavigateToLogin).toHaveBeenCalled()
  })
})
