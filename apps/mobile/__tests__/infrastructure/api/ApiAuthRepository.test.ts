import { ApiAuthRepository } from '@/infrastructure/api/ApiAuthRepository'
import { apiClient } from '@xenios/api-client'

jest.mock('@xenios/api-client', () => ({
  apiClient: {
    post: jest.fn(),
    get: jest.fn(),
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
  },
}))

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>

describe('ApiAuthRepository', () => {
  let repo: ApiAuthRepository

  beforeEach(() => {
    jest.clearAllMocks()
    repo = new ApiAuthRepository()
  })

  describe('login', () => {
    const credentials = { email: 'test@example.com', password: 'password123' }

    it('should login successfully', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: true,
        data: {
          user: {
            id: 'user-1',
            email: 'test@example.com',
            name: 'Test User',
            role: 'coach',
            avatar_url: null,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
          tokens: {
            access_token: 'access-123',
            refresh_token: 'refresh-456',
          },
        },
        error: null,
      })

      const result = await repo.login(credentials)

      expect(result.user.id).toBe('user-1')
      expect(result.user.email).toBe('test@example.com')
      expect(result.user.role).toBe('coach')
      expect(result.tokens.accessToken).toBe('access-123')
      expect(result.tokens.refreshToken).toBe('refresh-456')
      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/login', {
        email: 'test@example.com',
        password: 'password123',
      })
    })

    it('should throw on login failure', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: false,
        data: null,
        error: 'Invalid credentials',
      })

      await expect(repo.login(credentials)).rejects.toThrow(
        'Invalid credentials'
      )
    })

    it('should throw default message when no error returned', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: false,
        data: null,
        error: null,
      })

      await expect(repo.login(credentials)).rejects.toThrow('Login failed')
    })
  })

  describe('register', () => {
    const credentials = {
      email: 'new@example.com',
      password: 'password123',
      name: 'New User',
      role: 'client' as const,
    }

    it('should register successfully', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: true,
        data: {
          user: {
            id: 'user-2',
            email: 'new@example.com',
            name: 'New User',
            role: 'client',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
          tokens: {
            access_token: 'access-789',
            refresh_token: 'refresh-012',
          },
        },
        error: null,
      })

      const result = await repo.register(credentials)

      expect(result.user.id).toBe('user-2')
      expect(result.user.role).toBe('client')
      expect(result.tokens.accessToken).toBe('access-789')
      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/register', {
        email: 'new@example.com',
        password: 'password123',
        name: 'New User',
        role: 'client',
      })
    })

    it('should throw on registration failure', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: false,
        data: null,
        error: 'Email already exists',
      })

      await expect(repo.register(credentials)).rejects.toThrow(
        'Email already exists'
      )
    })
  })

  describe('refreshToken', () => {
    it('should refresh tokens successfully', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: true,
        data: {
          access_token: 'new-access',
          refresh_token: 'new-refresh',
        },
        error: null,
      })

      const result = await repo.refreshToken('old-refresh')

      expect(result.accessToken).toBe('new-access')
      expect(result.refreshToken).toBe('new-refresh')
      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/refresh', {
        refresh_token: 'old-refresh',
      })
    })

    it('should throw on refresh failure', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: false,
        data: null,
        error: 'Token expired',
      })

      await expect(repo.refreshToken('expired')).rejects.toThrow(
        'Token expired'
      )
    })
  })

  describe('logout', () => {
    it('should call logout endpoint', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: true,
        data: { message: 'logged out successfully' },
        error: null,
      })

      await repo.logout('access-token')

      expect(mockApiClient.setAuthToken).not.toHaveBeenCalled()
      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/logout')
    })

    it('should throw on logout failure', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: false,
        data: null,
        error: 'Unauthorized',
      })

      await expect(repo.logout('bad-token')).rejects.toThrow('Unauthorized')
    })
  })

  describe('getCurrentUser', () => {
    it('should fetch current user successfully', async () => {
      mockApiClient.get.mockResolvedValue({
        ok: true,
        data: {
          id: 'user-1',
          email: 'test@example.com',
          name: 'Test User',
          role: 'coach',
          avatar_url: 'https://example.com/avatar.png',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        error: null,
      })

      const user = await repo.getCurrentUser()

      expect(user.id).toBe('user-1')
      expect(user.email).toBe('test@example.com')
      expect(user.name).toBe('Test User')
      expect(user.role).toBe('coach')
      expect(user.avatarUrl).toBe('https://example.com/avatar.png')
      expect(mockApiClient.get).toHaveBeenCalledWith('/auth/me')
    })

    it('should throw on getCurrentUser failure', async () => {
      mockApiClient.get.mockResolvedValue({
        ok: false,
        data: null,
        error: 'Unauthorized',
      })

      await expect(repo.getCurrentUser()).rejects.toThrow('Unauthorized')
    })

    it('should throw default message when no error returned', async () => {
      mockApiClient.get.mockResolvedValue({
        ok: false,
        data: null,
        error: null,
      })

      await expect(repo.getCurrentUser()).rejects.toThrow(
        'Failed to get current user'
      )
    })

    it('should throw on invalid role from server', async () => {
      mockApiClient.get.mockResolvedValue({
        ok: true,
        data: {
          id: 'user-1',
          email: 'test@example.com',
          name: 'Test User',
          role: 'superadmin',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        error: null,
      })

      await expect(repo.getCurrentUser()).rejects.toThrow(
        'Invalid role received from server: superadmin'
      )
    })
  })

  describe('role validation', () => {
    it('should throw on invalid role during login', async () => {
      mockApiClient.post.mockResolvedValue({
        ok: true,
        data: {
          user: {
            id: 'user-1',
            email: 'test@example.com',
            name: 'Test User',
            role: 'unknown_role',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
          tokens: {
            access_token: 'access-123',
            refresh_token: 'refresh-456',
          },
        },
        error: null,
      })

      await expect(
        repo.login({ email: 'test@example.com', password: 'password123' })
      ).rejects.toThrow('Invalid role received from server: unknown_role')
    })
  })
})
