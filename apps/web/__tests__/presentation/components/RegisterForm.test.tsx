import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { RegisterForm } from '@/presentation/components/RegisterForm'
import { AuthProvider } from '@/presentation/hooks/useAuth'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { AuthTokenManager } from '@/domain/repositories/AuthTokenManager'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'

const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({ push: mockPush }),
}))

function createMockAuthRepo(): jest.Mocked<AuthRepository> {
  return {
    login: jest.fn(),
    register: jest.fn(),
    refresh: jest.fn(),
    logout: jest.fn(),
  }
}

function createMockTokenStorage(): jest.Mocked<TokenStorage> {
  return {
    getAccessToken: jest.fn().mockReturnValue(null),
    getRefreshToken: jest.fn().mockReturnValue(null),
    setTokens: jest.fn(),
    clearTokens: jest.fn(),
    getUser: jest.fn().mockReturnValue(null),
    setUser: jest.fn(),
  }
}

function createMockTokenManager(): jest.Mocked<AuthTokenManager> {
  return {
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
    restoreToken: jest.fn(),
  }
}

function renderRegisterForm(
  authRepo: jest.Mocked<AuthRepository> = createMockAuthRepo(),
  tokenStorage: TokenStorage = createMockTokenStorage(),
  tokenManager: AuthTokenManager = createMockTokenManager()
) {
  const loginUseCase = new LoginUseCase(authRepo)
  const registerUseCase = new RegisterUseCase(authRepo)
  const logoutUseCase = new LogoutUseCase(authRepo)
  return render(
    <AuthProvider
      loginUseCase={loginUseCase}
      registerUseCase={registerUseCase}
      logoutUseCase={logoutUseCase}
      tokenStorage={tokenStorage}
      tokenManager={tokenManager}
    >
      <RegisterForm />
    </AuthProvider>
  )
}

describe('RegisterForm', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders_AllFormFields', () => {
    renderRegisterForm()

    expect(screen.getByTestId('register-name')).toBeInTheDocument()
    expect(screen.getByTestId('register-email')).toBeInTheDocument()
    expect(screen.getByTestId('register-password')).toBeInTheDocument()
    expect(screen.getByTestId('register-role')).toBeInTheDocument()
    expect(screen.getByTestId('register-submit')).toBeInTheDocument()
  })

  test('submit_ValidInput_RedirectsToDashboard', async () => {
    const mockAuthRepo = createMockAuthRepo()
    mockAuthRepo.register.mockResolvedValue({
      user: {
        id: 'user-1',
        email: 'coach@example.com',
        name: 'Test Coach',
        role: 'coach',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      tokens: {
        access_token: 'access-token',
        refresh_token: 'refresh-token',
      },
    })

    renderRegisterForm(mockAuthRepo)

    fireEvent.change(screen.getByTestId('register-name'), {
      target: { value: 'Test Coach' },
    })
    fireEvent.change(screen.getByTestId('register-email'), {
      target: { value: 'coach@example.com' },
    })
    fireEvent.change(screen.getByTestId('register-password'), {
      target: { value: 'Password123!' },
    })
    fireEvent.click(screen.getByTestId('register-submit'))

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/dashboard')
    })
  })

  test('submit_ApiError_ShowsError', async () => {
    const mockAuthRepo = createMockAuthRepo()
    mockAuthRepo.register.mockRejectedValue(
      new Error('Email already registered')
    )

    renderRegisterForm(mockAuthRepo)

    fireEvent.change(screen.getByTestId('register-name'), {
      target: { value: 'Test' },
    })
    fireEvent.change(screen.getByTestId('register-email'), {
      target: { value: 'existing@example.com' },
    })
    fireEvent.change(screen.getByTestId('register-password'), {
      target: { value: 'Password123!' },
    })
    fireEvent.click(screen.getByTestId('register-submit'))

    await waitFor(() => {
      expect(screen.getByTestId('register-error')).toBeInTheDocument()
      expect(
        screen.getByText('Email already registered')
      ).toBeInTheDocument()
    })
  })
})
