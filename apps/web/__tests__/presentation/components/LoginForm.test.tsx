import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { LoginForm } from '@/presentation/components/LoginForm'
import { AuthProvider } from '@/presentation/hooks/useAuth'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { AuthTokenManager } from '@/domain/repositories/AuthTokenManager'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'

// Mock next/navigation
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
  }
}

function createMockTokenManager(): jest.Mocked<AuthTokenManager> {
  return {
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
    restoreToken: jest.fn(),
  }
}

function renderLoginForm(
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
      <LoginForm />
    </AuthProvider>
  )
}

describe('LoginForm', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders_AllFormFields', () => {
    renderLoginForm()

    expect(screen.getByTestId('login-email')).toBeInTheDocument()
    expect(screen.getByTestId('login-password')).toBeInTheDocument()
    expect(screen.getByTestId('login-submit')).toBeInTheDocument()
  })

  test('submit_ValidCredentials_RedirectsToDashboard', async () => {
    const mockAuthRepo = createMockAuthRepo()
    mockAuthRepo.login.mockResolvedValue({
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

    renderLoginForm(mockAuthRepo)

    fireEvent.change(screen.getByTestId('login-email'), {
      target: { value: 'coach@example.com' },
    })
    fireEvent.change(screen.getByTestId('login-password'), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByTestId('login-submit'))

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/dashboard')
    })
  })

  test('submit_InvalidCredentials_ShowsError', async () => {
    const mockAuthRepo = createMockAuthRepo()
    mockAuthRepo.login.mockRejectedValue(new Error('Invalid credentials'))

    renderLoginForm(mockAuthRepo)

    fireEvent.change(screen.getByTestId('login-email'), {
      target: { value: 'coach@example.com' },
    })
    fireEvent.change(screen.getByTestId('login-password'), {
      target: { value: 'wrong' },
    })
    fireEvent.click(screen.getByTestId('login-submit'))

    await waitFor(() => {
      expect(screen.getByTestId('login-error')).toBeInTheDocument()
      expect(screen.getByText('Invalid credentials')).toBeInTheDocument()
    })
  })

  test('submit_ShowsLoadingState', async () => {
    const mockAuthRepo = createMockAuthRepo()
    mockAuthRepo.login.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    )

    renderLoginForm(mockAuthRepo)

    fireEvent.change(screen.getByTestId('login-email'), {
      target: { value: 'coach@example.com' },
    })
    fireEvent.change(screen.getByTestId('login-password'), {
      target: { value: 'password123' },
    })
    fireEvent.click(screen.getByTestId('login-submit'))

    expect(screen.getByText('Signing in...')).toBeInTheDocument()
  })
})
