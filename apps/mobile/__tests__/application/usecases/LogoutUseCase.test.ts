import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { AuthTokens } from '@/domain/entities/AuthTokens'

const storedTokens: AuthTokens = {
  accessToken: 'access-token',
  refreshToken: 'refresh-token',
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

function createMockTokenStorage(): jest.Mocked<TokenStorageRepository> {
  return {
    saveTokens: jest.fn(),
    getTokens: jest.fn(),
    clearTokens: jest.fn(),
  }
}

describe('LogoutUseCase', () => {
  let useCase: LogoutUseCase
  let authRepo: jest.Mocked<AuthRepository>
  let tokenStorage: jest.Mocked<TokenStorageRepository>

  beforeEach(() => {
    authRepo = createMockAuthRepo()
    tokenStorage = createMockTokenStorage()
    useCase = new LogoutUseCase(authRepo, tokenStorage)
  })

  it('should logout and clear tokens', async () => {
    tokenStorage.getTokens.mockResolvedValue(storedTokens)
    authRepo.logout.mockResolvedValue(undefined)

    await useCase.execute()

    expect(authRepo.logout).toHaveBeenCalledWith('access-token')
    expect(tokenStorage.clearTokens).toHaveBeenCalled()
  })

  it('should clear tokens even when no stored tokens', async () => {
    tokenStorage.getTokens.mockResolvedValue(null)

    await useCase.execute()

    expect(authRepo.logout).not.toHaveBeenCalled()
    expect(tokenStorage.clearTokens).toHaveBeenCalled()
  })

  it('should clear tokens even when server logout fails', async () => {
    tokenStorage.getTokens.mockResolvedValue(storedTokens)
    authRepo.logout.mockRejectedValue(new Error('Network error'))

    await useCase.execute()

    expect(tokenStorage.clearTokens).toHaveBeenCalled()
  })
})
