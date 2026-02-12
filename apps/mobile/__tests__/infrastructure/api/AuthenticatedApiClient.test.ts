import { AuthenticatedApiClient } from '@/infrastructure/api/AuthenticatedApiClient'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { apiClient } from '@xenios/api-client'
import { AuthTokens } from '@/domain/entities/AuthTokens'

jest.mock('@xenios/api-client', () => ({
  apiClient: {
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
  },
}))

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>

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
  }
}

const storedTokens: AuthTokens = {
  accessToken: 'stored-access',
  refreshToken: 'stored-refresh',
}

const newTokens: AuthTokens = {
  accessToken: 'new-access',
  refreshToken: 'new-refresh',
}

describe('AuthenticatedApiClient', () => {
  let client: AuthenticatedApiClient
  let tokenStorage: jest.Mocked<TokenStorageRepository>
  let authRepo: jest.Mocked<AuthRepository>

  beforeEach(() => {
    jest.clearAllMocks()
    tokenStorage = createMockTokenStorage()
    authRepo = createMockAuthRepo()
    client = new AuthenticatedApiClient(tokenStorage, authRepo)
  })

  describe('configureAuth', () => {
    it('should set auth token when tokens exist', async () => {
      tokenStorage.getTokens.mockResolvedValue(storedTokens)

      await client.configureAuth()

      expect(mockApiClient.setAuthToken).toHaveBeenCalledWith('stored-access')
    })

    it('should not set auth token when no tokens', async () => {
      tokenStorage.getTokens.mockResolvedValue(null)

      await client.configureAuth()

      expect(mockApiClient.setAuthToken).not.toHaveBeenCalled()
    })
  })

  describe('handleUnauthorized', () => {
    it('should refresh tokens and return true on success', async () => {
      tokenStorage.getTokens.mockResolvedValue(storedTokens)
      authRepo.refreshToken.mockResolvedValue(newTokens)

      const result = await client.handleUnauthorized()

      expect(result).toBe(true)
      expect(authRepo.refreshToken).toHaveBeenCalledWith('stored-refresh')
      expect(tokenStorage.saveTokens).toHaveBeenCalledWith(newTokens)
      expect(mockApiClient.setAuthToken).toHaveBeenCalledWith('new-access')
    })

    it('should clear tokens and return false on refresh failure', async () => {
      tokenStorage.getTokens.mockResolvedValue(storedTokens)
      authRepo.refreshToken.mockRejectedValue(new Error('Expired'))

      const result = await client.handleUnauthorized()

      expect(result).toBe(false)
      expect(tokenStorage.clearTokens).toHaveBeenCalled()
      expect(mockApiClient.clearAuthToken).toHaveBeenCalled()
    })

    it('should clear tokens when no stored tokens available', async () => {
      tokenStorage.getTokens.mockResolvedValue(null)

      const result = await client.handleUnauthorized()

      expect(result).toBe(false)
      expect(tokenStorage.clearTokens).toHaveBeenCalled()
    })

    it('should not start multiple refreshes concurrently', async () => {
      tokenStorage.getTokens.mockResolvedValue(storedTokens)
      authRepo.refreshToken.mockImplementation(
        () =>
          new Promise((resolve) =>
            setTimeout(() => resolve(newTokens), 100)
          )
      )

      const [result1, result2] = await Promise.all([
        client.handleUnauthorized(),
        client.handleUnauthorized(),
      ])

      expect(result1).toBe(true)
      expect(result2).toBe(true)
      // Should only call refreshToken once despite two concurrent calls
      expect(authRepo.refreshToken).toHaveBeenCalledTimes(1)
    })
  })

  describe('clearAuth', () => {
    it('should clear auth token from api client', () => {
      client.clearAuth()
      expect(mockApiClient.clearAuthToken).toHaveBeenCalled()
    })
  })

  describe('setAccessToken', () => {
    it('should set auth token on api client', () => {
      client.setAccessToken('new-token')
      expect(mockApiClient.setAuthToken).toHaveBeenCalledWith('new-token')
    })
  })
})
