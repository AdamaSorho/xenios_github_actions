import { GetAuthStateUseCase } from '@/application/usecases/GetAuthStateUseCase'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { AuthTokens } from '@/domain/entities/AuthTokens'
import { AuthUser } from '@/domain/entities/AuthUser'

const storedTokens: AuthTokens = {
  accessToken: 'access-token',
  refreshToken: 'refresh-token',
}

const mockUser: AuthUser = {
  id: 'user-1',
  email: 'test@example.com',
  name: 'Test User',
  role: 'coach',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

function createMockTokenStorage(): jest.Mocked<TokenStorageRepository> {
  return {
    saveTokens: jest.fn(),
    getTokens: jest.fn(),
    clearTokens: jest.fn(),
  }
}

function createMockAuthRepo(): jest.Mocked<AuthRepository> {
  return {
    login: jest.fn(),
    register: jest.fn(),
    refreshToken: jest.fn(),
    logout: jest.fn(),
    getCurrentUser: jest.fn(),
  }
}

describe('GetAuthStateUseCase', () => {
  let useCase: GetAuthStateUseCase
  let tokenStorage: jest.Mocked<TokenStorageRepository>
  let authRepo: jest.Mocked<AuthRepository>

  beforeEach(() => {
    tokenStorage = createMockTokenStorage()
    authRepo = createMockAuthRepo()
    useCase = new GetAuthStateUseCase(tokenStorage, authRepo)
  })

  it('should return authenticated state with user when tokens exist', async () => {
    tokenStorage.getTokens.mockResolvedValue(storedTokens)
    authRepo.getCurrentUser.mockResolvedValue(mockUser)

    const state = await useCase.execute()

    expect(state.isAuthenticated).toBe(true)
    expect(state.accessToken).toBe('access-token')
    expect(state.user).toEqual(mockUser)
  })

  it('should return unauthenticated state when no tokens', async () => {
    tokenStorage.getTokens.mockResolvedValue(null)

    const state = await useCase.execute()

    expect(state.isAuthenticated).toBe(false)
    expect(state.accessToken).toBeNull()
    expect(state.user).toBeNull()
  })

  it('should return authenticated with null user when getCurrentUser fails', async () => {
    tokenStorage.getTokens.mockResolvedValue(storedTokens)
    authRepo.getCurrentUser.mockRejectedValue(new Error('Network error'))

    const state = await useCase.execute()

    expect(state.isAuthenticated).toBe(true)
    expect(state.accessToken).toBe('access-token')
    expect(state.user).toBeNull()
  })
})
