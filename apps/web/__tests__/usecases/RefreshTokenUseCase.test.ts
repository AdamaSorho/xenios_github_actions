import { RefreshTokenUseCase } from '@/application/usecases/RefreshTokenUseCase'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { AuthTokens } from '@/domain/entities/AuthTokens'

describe('RefreshTokenUseCase', () => {
  const mockTokens: AuthTokens = {
    access_token: 'new-access-token',
    refresh_token: 'new-refresh-token',
  }

  let mockAuthRepo: jest.Mocked<AuthRepository>
  let useCase: RefreshTokenUseCase

  beforeEach(() => {
    mockAuthRepo = {
      login: jest.fn(),
      register: jest.fn(),
      refresh: jest.fn(),
      logout: jest.fn(),
    }
    useCase = new RefreshTokenUseCase(mockAuthRepo)
  })

  test('execute_ValidRefreshToken_ReturnsNewTokens', async () => {
    mockAuthRepo.refresh.mockResolvedValue(mockTokens)

    const result = await useCase.execute('old-refresh-token')

    expect(result).toEqual(mockTokens)
    expect(mockAuthRepo.refresh).toHaveBeenCalledWith('old-refresh-token')
  })

  test('execute_EmptyRefreshToken_ThrowsValidationError', async () => {
    await expect(useCase.execute('')).rejects.toThrow(
      'Refresh token is required'
    )
    expect(mockAuthRepo.refresh).not.toHaveBeenCalled()
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockAuthRepo.refresh.mockRejectedValue(new Error('Token expired'))

    await expect(useCase.execute('expired-token')).rejects.toThrow(
      'Token expired'
    )
  })
})
