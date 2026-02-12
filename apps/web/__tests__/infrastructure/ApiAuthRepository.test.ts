import { ApiAuthRepository } from '@/infrastructure/repositories/ApiAuthRepository'
import { apiClient } from '@xenios/api-client'

jest.mock('@xenios/api-client', () => ({
  apiClient: {
    post: jest.fn(),
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('ApiAuthRepository', () => {
  let repo: ApiAuthRepository

  beforeEach(() => {
    jest.clearAllMocks()
    repo = new ApiAuthRepository()
  })

  describe('login', () => {
    test('login_ValidCredentials_ReturnsAuthResponse', async () => {
      const authResponse = {
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
      }

      mockedApiClient.post.mockResolvedValue({
        data: authResponse,
        error: null,
        ok: true,
      })

      const result = await repo.login({
        email: 'coach@example.com',
        password: 'password123',
      })

      expect(result).toEqual(authResponse)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/auth/login', {
        email: 'coach@example.com',
        password: 'password123',
      })
    })

    test('login_ApiError_ThrowsError', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: null,
        error: 'Invalid credentials',
        ok: false,
      })

      await expect(
        repo.login({ email: 'coach@example.com', password: 'wrong' })
      ).rejects.toThrow('Invalid credentials')
    })
  })

  describe('register', () => {
    test('register_ValidInput_ReturnsAuthResponse', async () => {
      const authResponse = {
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
      }

      mockedApiClient.post.mockResolvedValue({
        data: authResponse,
        error: null,
        ok: true,
      })

      const result = await repo.register({
        email: 'coach@example.com',
        password: 'Password123!',
        name: 'Test Coach',
        role: 'coach',
      })

      expect(result).toEqual(authResponse)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/auth/register', {
        email: 'coach@example.com',
        password: 'Password123!',
        name: 'Test Coach',
        role: 'coach',
      })
    })

    test('register_ApiError_ThrowsError', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: null,
        error: 'Email already registered',
        ok: false,
      })

      await expect(
        repo.register({
          email: 'coach@example.com',
          password: 'Password123!',
          name: 'Test',
          role: 'coach',
        })
      ).rejects.toThrow('Email already registered')
    })
  })

  describe('refresh', () => {
    test('refresh_ValidToken_ReturnsNewTokens', async () => {
      const tokens = {
        access_token: 'new-access',
        refresh_token: 'new-refresh',
      }

      mockedApiClient.post.mockResolvedValue({
        data: tokens,
        error: null,
        ok: true,
      })

      const result = await repo.refresh('old-refresh-token')

      expect(result).toEqual(tokens)
      expect(mockedApiClient.post).toHaveBeenCalledWith('/auth/refresh', {
        refresh_token: 'old-refresh-token',
      })
    })

    test('refresh_ApiError_ThrowsError', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: null,
        error: 'Token expired',
        ok: false,
      })

      await expect(repo.refresh('expired-token')).rejects.toThrow(
        'Token expired'
      )
    })
  })

  describe('logout', () => {
    test('logout_Success_Completes', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: { message: 'logged out successfully' },
        error: null,
        ok: true,
      })

      await expect(repo.logout()).resolves.toBeUndefined()
      expect(mockedApiClient.post).toHaveBeenCalledWith('/v1/auth/logout')
    })

    test('logout_ApiError_ThrowsError', async () => {
      mockedApiClient.post.mockResolvedValue({
        data: null,
        error: 'Unauthorized',
        ok: false,
      })

      await expect(repo.logout()).rejects.toThrow('Unauthorized')
    })
  })
})
