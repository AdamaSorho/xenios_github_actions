import { GetAuthStateUseCase } from '@/application/usecases/GetAuthStateUseCase'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { AuthTokens } from '@/domain/entities/AuthTokens'

const storedTokens: AuthTokens = {
  accessToken: 'access-token',
  refreshToken: 'refresh-token',
}

function createMockTokenStorage(): jest.Mocked<TokenStorageRepository> {
  return {
    saveTokens: jest.fn(),
    getTokens: jest.fn(),
    clearTokens: jest.fn(),
  }
}

describe('GetAuthStateUseCase', () => {
  let useCase: GetAuthStateUseCase
  let tokenStorage: jest.Mocked<TokenStorageRepository>

  beforeEach(() => {
    tokenStorage = createMockTokenStorage()
    useCase = new GetAuthStateUseCase(tokenStorage)
  })

  it('should return authenticated state when tokens exist', async () => {
    tokenStorage.getTokens.mockResolvedValue(storedTokens)

    const state = await useCase.execute()

    expect(state.isAuthenticated).toBe(true)
    expect(state.accessToken).toBe('access-token')
  })

  it('should return unauthenticated state when no tokens', async () => {
    tokenStorage.getTokens.mockResolvedValue(null)

    const state = await useCase.execute()

    expect(state.isAuthenticated).toBe(false)
    expect(state.accessToken).toBeNull()
  })
})
