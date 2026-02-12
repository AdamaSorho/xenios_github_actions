import { RefreshTokenUseCase } from '@/application/usecases/RefreshTokenUseCase'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { AuthTokens } from '@/domain/entities/AuthTokens'

const currentTokens: AuthTokens = {
  accessToken: 'old-access',
  refreshToken: 'old-refresh',
}

const newTokens: AuthTokens = {
  accessToken: 'new-access',
  refreshToken: 'new-refresh',
}

function createMockAuthRepo(): jest.Mocked<AuthRepository> {
  return {
    login: jest.fn(),
    register: jest.fn(),
    refreshToken: jest.fn(),
    logout: jest.fn(),
  }
}

function createMockTokenStorage(): jest.Mocked<TokenStorageRepository> {
  return {
    saveTokens: jest.fn(),
    getTokens: jest.fn(),
    clearTokens: jest.fn(),
  }
}

describe('RefreshTokenUseCase', () => {
  let useCase: RefreshTokenUseCase
  let authRepo: jest.Mocked<AuthRepository>
  let tokenStorage: jest.Mocked<TokenStorageRepository>

  beforeEach(() => {
    authRepo = createMockAuthRepo()
    tokenStorage = createMockTokenStorage()
    useCase = new RefreshTokenUseCase(authRepo, tokenStorage)
  })

  it('should refresh tokens successfully', async () => {
    tokenStorage.getTokens.mockResolvedValue(currentTokens)
    authRepo.refreshToken.mockResolvedValue(newTokens)

    const result = await useCase.execute()

    expect(result).toEqual(newTokens)
    expect(authRepo.refreshToken).toHaveBeenCalledWith('old-refresh')
    expect(tokenStorage.saveTokens).toHaveBeenCalledWith(newTokens)
  })

  it('should throw when no stored tokens exist', async () => {
    tokenStorage.getTokens.mockResolvedValue(null)

    await expect(useCase.execute()).rejects.toThrow(
      'No refresh token available'
    )
    expect(authRepo.refreshToken).not.toHaveBeenCalled()
  })

  it('should propagate refresh errors', async () => {
    tokenStorage.getTokens.mockResolvedValue(currentTokens)
    authRepo.refreshToken.mockRejectedValue(new Error('Token expired'))

    await expect(useCase.execute()).rejects.toThrow('Token expired')
  })
})
