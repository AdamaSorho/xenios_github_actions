import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { Sidebar } from '@/presentation/components/Sidebar'
import { AuthProvider } from '@/presentation/hooks/useAuth'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { AuthTokenManager } from '@/domain/repositories/AuthTokenManager'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'

// Mock next/navigation
let mockPathname = '/dashboard'
jest.mock('next/navigation', () => ({
  usePathname: () => mockPathname,
}))

// Mock next/link
jest.mock('next/link', () => {
  return function MockLink({
    children,
    href,
    ...props
  }: {
    children: React.ReactNode
    href: string
    [key: string]: unknown
  }) {
    return (
      <a href={href} {...props}>
        {children}
      </a>
    )
  }
})

function createMockAuthRepo(): jest.Mocked<AuthRepository> {
  return {
    login: jest.fn(),
    register: jest.fn(),
    refresh: jest.fn(),
    logout: jest.fn().mockResolvedValue(undefined),
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

function renderSidebar(
  authRepo?: jest.Mocked<AuthRepository>,
) {
  const repo = authRepo ?? createMockAuthRepo()
  const storage = createMockTokenStorage()
  const tokenManager = createMockTokenManager()
  const loginUseCase = new LoginUseCase(repo)
  const registerUseCase = new RegisterUseCase(repo)
  const logoutUseCase = new LogoutUseCase(repo)

  return render(
    <AuthProvider
      loginUseCase={loginUseCase}
      registerUseCase={registerUseCase}
      logoutUseCase={logoutUseCase}
      tokenStorage={storage}
      tokenManager={tokenManager}
    >
      <Sidebar />
    </AuthProvider>
  )
}

describe('Sidebar', () => {
  beforeEach(() => {
    mockPathname = '/dashboard'
  })

  test('renders_AllNavigationLinks', () => {
    renderSidebar()

    expect(screen.getByTestId('nav-dashboard')).toBeInTheDocument()
    expect(screen.getByTestId('nav-clients')).toBeInTheDocument()
    expect(screen.getByTestId('nav-sessions')).toBeInTheDocument()
    expect(screen.getByTestId('nav-analytics')).toBeInTheDocument()
    expect(screen.getByTestId('nav-settings')).toBeInTheDocument()
  })

  test('renders_LogoutButton', () => {
    renderSidebar()

    expect(screen.getByTestId('logout-button')).toBeInTheDocument()
  })

  test('highlightsActiveRoute_Dashboard', () => {
    mockPathname = '/dashboard'
    renderSidebar()

    const dashboardLink = screen.getByTestId('nav-dashboard')
    expect(dashboardLink.className).toContain('bg-blue-50')
  })

  test('highlightsActiveRoute_Clients', () => {
    mockPathname = '/clients'
    renderSidebar()

    const clientsLink = screen.getByTestId('nav-clients')
    expect(clientsLink.className).toContain('bg-blue-50')
  })

  test('highlightsActiveRoute_ClientDetail', () => {
    mockPathname = '/clients/abc-123'
    renderSidebar()

    const clientsLink = screen.getByTestId('nav-clients')
    expect(clientsLink.className).toContain('bg-blue-50')
  })

  test('logout_ClickButton_CallsLogout', () => {
    const mockAuthRepo = createMockAuthRepo()
    renderSidebar(mockAuthRepo)

    fireEvent.click(screen.getByTestId('logout-button'))

    expect(mockAuthRepo.logout).toHaveBeenCalled()
  })
})
