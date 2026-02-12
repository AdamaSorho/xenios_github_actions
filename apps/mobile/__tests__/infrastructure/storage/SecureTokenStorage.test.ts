import { SecureTokenStorage } from '@/infrastructure/storage/SecureTokenStorage'
import { AuthTokens } from '@/domain/entities/AuthTokens'
import * as Keychain from 'react-native-keychain'

jest.mock('react-native-keychain', () => ({
  setGenericPassword: jest.fn(),
  getGenericPassword: jest.fn(),
  resetGenericPassword: jest.fn(),
}))

const mockKeychain = Keychain as jest.Mocked<typeof Keychain>

const tokens: AuthTokens = {
  accessToken: 'access-123',
  refreshToken: 'refresh-456',
}

describe('SecureTokenStorage', () => {
  let storage: SecureTokenStorage

  beforeEach(() => {
    jest.clearAllMocks()
    storage = new SecureTokenStorage()
  })

  describe('saveTokens', () => {
    it('should save tokens to keychain', async () => {
      mockKeychain.setGenericPassword.mockResolvedValue({
        service: 'com.xenios.auth',
        storage: 'keychain',
      } as any)

      await storage.saveTokens(tokens)

      expect(mockKeychain.setGenericPassword).toHaveBeenCalledWith(
        'auth_tokens',
        JSON.stringify(tokens),
        { service: 'com.xenios.auth' }
      )
    })
  })

  describe('getTokens', () => {
    it('should retrieve tokens from keychain', async () => {
      mockKeychain.getGenericPassword.mockResolvedValue({
        username: 'auth_tokens',
        password: JSON.stringify(tokens),
        service: 'com.xenios.auth',
        storage: 'keychain',
      } as any)

      const result = await storage.getTokens()

      expect(result).toEqual(tokens)
      expect(mockKeychain.getGenericPassword).toHaveBeenCalledWith({
        service: 'com.xenios.auth',
      })
    })

    it('should return null when no credentials stored', async () => {
      mockKeychain.getGenericPassword.mockResolvedValue(false)

      const result = await storage.getTokens()

      expect(result).toBeNull()
    })

    it('should return null when stored data is invalid JSON', async () => {
      mockKeychain.getGenericPassword.mockResolvedValue({
        username: 'auth_tokens',
        password: 'not-json',
        service: 'com.xenios.auth',
        storage: 'keychain',
      } as any)

      const result = await storage.getTokens()

      expect(result).toBeNull()
    })

    it('should return null when parsed JSON is missing accessToken', async () => {
      mockKeychain.getGenericPassword.mockResolvedValue({
        username: 'auth_tokens',
        password: JSON.stringify({ refreshToken: 'refresh-456' }),
        service: 'com.xenios.auth',
        storage: 'keychain',
      } as any)

      const result = await storage.getTokens()

      expect(result).toBeNull()
    })

    it('should return null when parsed JSON is missing refreshToken', async () => {
      mockKeychain.getGenericPassword.mockResolvedValue({
        username: 'auth_tokens',
        password: JSON.stringify({ accessToken: 'access-123' }),
        service: 'com.xenios.auth',
        storage: 'keychain',
      } as any)

      const result = await storage.getTokens()

      expect(result).toBeNull()
    })

    it('should return null when parsed JSON has empty accessToken', async () => {
      mockKeychain.getGenericPassword.mockResolvedValue({
        username: 'auth_tokens',
        password: JSON.stringify({ accessToken: '', refreshToken: 'refresh-456' }),
        service: 'com.xenios.auth',
        storage: 'keychain',
      } as any)

      const result = await storage.getTokens()

      expect(result).toBeNull()
    })

    it('should return null when parsed JSON has non-string tokens', async () => {
      mockKeychain.getGenericPassword.mockResolvedValue({
        username: 'auth_tokens',
        password: JSON.stringify({ accessToken: 123, refreshToken: true }),
        service: 'com.xenios.auth',
        storage: 'keychain',
      } as any)

      const result = await storage.getTokens()

      expect(result).toBeNull()
    })

    it('should return null when parsed JSON is an array', async () => {
      mockKeychain.getGenericPassword.mockResolvedValue({
        username: 'auth_tokens',
        password: JSON.stringify(['access', 'refresh']),
        service: 'com.xenios.auth',
        storage: 'keychain',
      } as any)

      const result = await storage.getTokens()

      expect(result).toBeNull()
    })
  })

  describe('clearTokens', () => {
    it('should clear tokens from keychain', async () => {
      mockKeychain.resetGenericPassword.mockResolvedValue(true)

      await storage.clearTokens()

      expect(mockKeychain.resetGenericPassword).toHaveBeenCalledWith({
        service: 'com.xenios.auth',
      })
    })
  })
})
