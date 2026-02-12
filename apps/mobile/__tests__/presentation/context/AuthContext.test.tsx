import React from 'react'
import { Text, TouchableOpacity } from 'react-native'
import { render, fireEvent, waitFor, act } from '@testing-library/react-native'
import { AuthProvider, useAuth, AuthProviderDeps } from '@/presentation/context/AuthContext'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'
import { GetAuthStateUseCase } from '@/application/usecases/GetAuthStateUseCase'
import { AuthClientConfigurator } from '@/domain/repositories/AuthClientConfigurator'
import { mockUser } from '../../helpers/mocks'

function createMockDeps(): AuthProviderDeps {
  return {
    loginUseCase: {
      execute: jest.fn(),
    } as unknown as LoginUseCase,
    registerUseCase: {
      execute: jest.fn(),
    } as unknown as RegisterUseCase,
    logoutUseCase: {
      execute: jest.fn(),
    } as unknown as LogoutUseCase,
    getAuthStateUseCase: {
      execute: jest.fn().mockResolvedValue({ isAuthenticated: false, accessToken: null, user: null }),
    } as unknown as GetAuthStateUseCase,
    authenticatedApiClient: {
      configureAuth: jest.fn(),
      setAccessToken: jest.fn(),
      clearAuth: jest.fn(),
    } as AuthClientConfigurator,
  }
}

function TestConsumer() {
  const { user, isAuthenticated, isLoading, error, login, register, logout, clearError } =
    useAuth()

  return (
    <>
      <Text testID="loading">{String(isLoading)}</Text>
      <Text testID="authenticated">{String(isAuthenticated)}</Text>
      <Text testID="user-name">{user?.name || 'none'}</Text>
      <Text testID="error">{error || 'none'}</Text>
      <TouchableOpacity
        testID="login-btn"
        onPress={() =>
          login({ email: 'test@example.com', password: 'password123' }).catch(
            () => {}
          )
        }
      />
      <TouchableOpacity
        testID="register-btn"
        onPress={() =>
          register({
            email: 'test@example.com',
            password: 'password123',
            name: 'Test',
            role: 'coach',
          }).catch(() => {})
        }
      />
      <TouchableOpacity testID="logout-btn" onPress={() => logout()} />
      <TouchableOpacity testID="clear-error-btn" onPress={clearError} />
    </>
  )
}

function renderWithAuth(deps: AuthProviderDeps) {
  return render(
    <AuthProvider deps={deps}>
      <TestConsumer />
    </AuthProvider>
  )
}

describe('AuthContext', () => {
  it('should start in loading state and resolve to unauthenticated', async () => {
    const deps = createMockDeps()
    const { getByTestId } = renderWithAuth(deps)

    await waitFor(() => {
      expect(getByTestId('loading').props.children).toBe('false')
    })

    expect(getByTestId('authenticated').props.children).toBe('false')
    expect(getByTestId('user-name').props.children).toBe('none')
  })

  it('should restore user and configure auth when stored tokens exist on mount', async () => {
    const deps = createMockDeps()
    ;(deps.getAuthStateUseCase.execute as jest.Mock).mockResolvedValue({
      isAuthenticated: true,
      accessToken: 'stored-token',
      user: mockUser,
    })

    const { getByTestId } = renderWithAuth(deps)

    await waitFor(() => {
      expect(deps.authenticatedApiClient.configureAuth).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(getByTestId('authenticated').props.children).toBe('true')
      expect(getByTestId('user-name').props.children).toBe('Test User')
    })
  })

  it('should login successfully', async () => {
    const deps = createMockDeps()
    ;(deps.loginUseCase.execute as jest.Mock).mockResolvedValue({
      user: mockUser,
      accessToken: 'new-token',
    })

    const { getByTestId } = renderWithAuth(deps)

    await waitFor(() => {
      expect(getByTestId('loading').props.children).toBe('false')
    })

    await act(async () => {
      fireEvent.press(getByTestId('login-btn'))
    })

    await waitFor(() => {
      expect(getByTestId('authenticated').props.children).toBe('true')
      expect(getByTestId('user-name').props.children).toBe('Test User')
    })

    expect(deps.authenticatedApiClient.setAccessToken).toHaveBeenCalledWith('new-token')
  })

  it('should set error on login failure', async () => {
    const deps = createMockDeps()
    ;(deps.loginUseCase.execute as jest.Mock).mockRejectedValue(
      new Error('Invalid credentials')
    )

    const { getByTestId } = renderWithAuth(deps)

    await waitFor(() => {
      expect(getByTestId('loading').props.children).toBe('false')
    })

    await act(async () => {
      fireEvent.press(getByTestId('login-btn'))
    })

    await waitFor(() => {
      expect(getByTestId('error').props.children).toBe('Invalid credentials')
    })
  })

  it('should register successfully', async () => {
    const deps = createMockDeps()
    ;(deps.registerUseCase.execute as jest.Mock).mockResolvedValue({
      user: mockUser,
      accessToken: 'new-token',
    })

    const { getByTestId } = renderWithAuth(deps)

    await waitFor(() => {
      expect(getByTestId('loading').props.children).toBe('false')
    })

    await act(async () => {
      fireEvent.press(getByTestId('register-btn'))
    })

    await waitFor(() => {
      expect(getByTestId('authenticated').props.children).toBe('true')
    })

    expect(deps.authenticatedApiClient.setAccessToken).toHaveBeenCalledWith('new-token')
  })

  it('should logout and clear user', async () => {
    const deps = createMockDeps()
    ;(deps.loginUseCase.execute as jest.Mock).mockResolvedValue({
      user: mockUser,
      accessToken: 'token',
    })
    ;(deps.logoutUseCase.execute as jest.Mock).mockResolvedValue(undefined)

    const { getByTestId } = renderWithAuth(deps)

    await waitFor(() => {
      expect(getByTestId('loading').props.children).toBe('false')
    })

    await act(async () => {
      fireEvent.press(getByTestId('login-btn'))
    })

    await waitFor(() => {
      expect(getByTestId('authenticated').props.children).toBe('true')
    })

    await act(async () => {
      fireEvent.press(getByTestId('logout-btn'))
    })

    await waitFor(() => {
      expect(getByTestId('authenticated').props.children).toBe('false')
      expect(getByTestId('user-name').props.children).toBe('none')
    })

    expect(deps.authenticatedApiClient.clearAuth).toHaveBeenCalled()
  })

  it('should clear error', async () => {
    const deps = createMockDeps()
    ;(deps.loginUseCase.execute as jest.Mock).mockRejectedValue(
      new Error('fail')
    )

    const { getByTestId } = renderWithAuth(deps)

    await waitFor(() => {
      expect(getByTestId('loading').props.children).toBe('false')
    })

    await act(async () => {
      fireEvent.press(getByTestId('login-btn'))
    })

    await waitFor(() => {
      expect(getByTestId('error').props.children).toBe('fail')
    })

    await act(async () => {
      fireEvent.press(getByTestId('clear-error-btn'))
    })

    expect(getByTestId('error').props.children).toBe('none')
  })
})

describe('useAuth', () => {
  it('should throw when used outside AuthProvider', () => {
    const consoleError = jest.spyOn(console, 'error').mockImplementation()

    function BadComponent() {
      useAuth()
      return null
    }

    expect(() => render(<BadComponent />)).toThrow(
      'useAuth must be used within an AuthProvider'
    )

    consoleError.mockRestore()
  })
})
