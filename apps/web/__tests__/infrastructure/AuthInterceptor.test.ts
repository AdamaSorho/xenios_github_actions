import { AuthInterceptor } from '@/infrastructure/auth/AuthInterceptor'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { AuthTokenManager } from '@/domain/repositories/AuthTokenManager'
import {
  createMockAuthRepo,
  createMockTokenStorage,
  createMockTokenManager,
} from '../helpers/mockFactories'

// Mock apiClient with controllable methods
const mockGet = jest.fn()
const mockPost = jest.fn()
const mockPut = jest.fn()
const mockDelete = jest.fn()

// Create fresh wrapper functions that delegate to the jest.fn() mocks
function resetApiClientMock() {
  const { apiClient } = require('@xenios/api-client')
  apiClient.get = (...args: unknown[]) => mockGet(...args)
  apiClient.post = (...args: unknown[]) => mockPost(...args)
  apiClient.put = (...args: unknown[]) => mockPut(...args)
  apiClient.delete = (...args: unknown[]) => mockDelete(...args)
}

jest.mock('@xenios/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    put: (...args: unknown[]) => mockPut(...args),
    delete: (...args: unknown[]) => mockDelete(...args),
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
  },
}))

describe('AuthInterceptor', () => {
  let tokenStorage: jest.Mocked<TokenStorage>
  let authRepo: jest.Mocked<AuthRepository>
  let tokenManager: jest.Mocked<AuthTokenManager>

  beforeEach(() => {
    jest.clearAllMocks()
    // Reset apiClient methods to original mock delegates before each test
    resetApiClientMock()
    tokenStorage = createMockTokenStorage()
    authRepo = createMockAuthRepo()
    tokenManager = createMockTokenManager()
  })

  test('install_NonUnauthorizedResponse_PassesThrough', async () => {
    mockGet.mockResolvedValue({ ok: true, data: { id: 1 }, error: null })

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const result = await apiClient.get('/api/users')

    expect(result).toEqual({ ok: true, data: { id: 1 }, error: null })
    expect(authRepo.refresh).not.toHaveBeenCalled()
  })

  test('install_UnauthorizedGet_RefreshesAndRetries', async () => {
    mockGet
      .mockResolvedValueOnce({ ok: false, data: null, error: 'Unauthorized' })
      .mockResolvedValueOnce({ ok: true, data: { id: 1 }, error: null })

    tokenStorage.getRefreshToken.mockReturnValue('refresh-token')
    authRepo.refresh.mockResolvedValue({
      access_token: 'new-access-token',
      refresh_token: 'new-refresh-token',
    })

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const result = await apiClient.get('/api/users')

    expect(authRepo.refresh).toHaveBeenCalledWith('refresh-token')
    expect(tokenStorage.setTokens).toHaveBeenCalledWith({
      access_token: 'new-access-token',
      refresh_token: 'new-refresh-token',
    })
    expect(tokenManager.setAuthToken).toHaveBeenCalledWith('new-access-token')
    expect(result).toEqual({ ok: true, data: { id: 1 }, error: null })
  })

  test('install_UnauthorizedGet_NoRefreshToken_ClearsAuth', async () => {
    mockGet.mockResolvedValue({ ok: false, data: null, error: 'Unauthorized' })
    tokenStorage.getRefreshToken.mockReturnValue(null)

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const result = await apiClient.get('/api/users')

    expect(tokenStorage.clearTokens).toHaveBeenCalled()
    expect(tokenManager.clearAuthToken).toHaveBeenCalled()
    expect(result).toEqual({ ok: false, data: null, error: 'Unauthorized' })
  })

  test('install_UnauthorizedGet_RefreshFails_ClearsAuth', async () => {
    mockGet.mockResolvedValue({ ok: false, data: null, error: 'Unauthorized' })
    tokenStorage.getRefreshToken.mockReturnValue('refresh-token')
    authRepo.refresh.mockRejectedValue(new Error('Refresh failed'))

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const result = await apiClient.get('/api/users')

    expect(tokenStorage.clearTokens).toHaveBeenCalled()
    expect(tokenManager.clearAuthToken).toHaveBeenCalled()
    expect(result).toEqual({ ok: false, data: null, error: 'Unauthorized' })
  })

  test('install_AuthEndpoints_SkipsInterceptor', async () => {
    mockPost.mockResolvedValue({ ok: false, data: null, error: 'Unauthorized' })

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const result = await apiClient.post('/auth/login', { email: 'a@b.com' })

    expect(authRepo.refresh).not.toHaveBeenCalled()
    expect(result).toEqual({ ok: false, data: null, error: 'Unauthorized' })
  })

  test('install_TokenExpiredError_TriggersRefresh', async () => {
    mockPut
      .mockResolvedValueOnce({ ok: false, data: null, error: 'token expired' })
      .mockResolvedValueOnce({ ok: true, data: { updated: true }, error: null })

    tokenStorage.getRefreshToken.mockReturnValue('refresh-token')
    authRepo.refresh.mockResolvedValue({
      access_token: 'new-at',
      refresh_token: 'new-rt',
    })

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const result = await apiClient.put('/api/users/1', { name: 'Test' })

    expect(authRepo.refresh).toHaveBeenCalled()
    expect(result).toEqual({ ok: true, data: { updated: true }, error: null })
  })

  test('install_DeleteUnauthorized_RefreshesAndRetries', async () => {
    mockDelete
      .mockResolvedValueOnce({ ok: false, data: null, error: '401' })
      .mockResolvedValueOnce({ ok: true, data: null, error: null })

    tokenStorage.getRefreshToken.mockReturnValue('refresh-token')
    authRepo.refresh.mockResolvedValue({
      access_token: 'new-at',
      refresh_token: 'new-rt',
    })

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const result = await apiClient.delete('/api/users/1')

    expect(authRepo.refresh).toHaveBeenCalled()
    expect(result).toEqual({ ok: true, data: null, error: null })
  })

  test('install_Concurrent401s_SharesSingleRefresh', async () => {
    // Both GET calls return 401 initially, then succeed after refresh
    mockGet
      .mockResolvedValueOnce({ ok: false, data: null, error: 'Unauthorized' })
      .mockResolvedValueOnce({ ok: false, data: null, error: 'Unauthorized' })
      .mockResolvedValueOnce({ ok: true, data: { id: 1 }, error: null })
      .mockResolvedValueOnce({ ok: true, data: { id: 2 }, error: null })

    tokenStorage.getRefreshToken.mockReturnValue('refresh-token')
    authRepo.refresh.mockResolvedValue({
      access_token: 'new-at',
      refresh_token: 'new-rt',
    })

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const [result1, result2] = await Promise.all([
      apiClient.get('/api/users'),
      apiClient.get('/api/posts'),
    ])

    // Only one refresh call should have been made despite two 401s
    expect(authRepo.refresh).toHaveBeenCalledTimes(1)
    expect(result1).toEqual({ ok: true, data: { id: 1 }, error: null })
    expect(result2).toEqual({ ok: true, data: { id: 2 }, error: null })
  })

  test('install_401Unauthorized_ExactMatch_TriggersRefresh', async () => {
    mockGet
      .mockResolvedValueOnce({ ok: false, data: null, error: '401 Unauthorized' })
      .mockResolvedValueOnce({ ok: true, data: { id: 1 }, error: null })

    tokenStorage.getRefreshToken.mockReturnValue('refresh-token')
    authRepo.refresh.mockResolvedValue({
      access_token: 'new-at',
      refresh_token: 'new-rt',
    })

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const result = await apiClient.get('/api/users')

    expect(authRepo.refresh).toHaveBeenCalled()
    expect(result).toEqual({ ok: true, data: { id: 1 }, error: null })
  })

  test('install_NonMatchingError_DoesNotTriggerRefresh', async () => {
    mockGet.mockResolvedValue({
      ok: false,
      data: null,
      error: 'something unauthorized happened',
    })

    const interceptor = new AuthInterceptor(tokenStorage, authRepo, tokenManager)
    interceptor.install()

    const { apiClient } = require('@xenios/api-client')
    const result = await apiClient.get('/api/users')

    expect(authRepo.refresh).not.toHaveBeenCalled()
    expect(result).toEqual({
      ok: false,
      data: null,
      error: 'something unauthorized happened',
    })
  })
})

describe('AuthInterceptor.isAuthEndpoint', () => {
  let interceptor: AuthInterceptor

  beforeEach(() => {
    resetApiClientMock()
    interceptor = new AuthInterceptor(
      createMockTokenStorage(),
      createMockAuthRepo(),
      createMockTokenManager()
    )
  })

  test('exactAuthPaths_Match', () => {
    expect(interceptor.isAuthEndpoint('/auth/login')).toBe(true)
    expect(interceptor.isAuthEndpoint('/auth/register')).toBe(true)
    expect(interceptor.isAuthEndpoint('/auth/refresh')).toBe(true)
  })

  test('logoutPath_Matches', () => {
    expect(interceptor.isAuthEndpoint('/auth/logout')).toBe(true)
  })

  test('nonAuthPaths_DoNotMatch', () => {
    expect(interceptor.isAuthEndpoint('/api/users')).toBe(false)
  })

  test('substringPaths_DoNotMatch', () => {
    expect(interceptor.isAuthEndpoint('/v2/auth/login-history')).toBe(false)
    expect(interceptor.isAuthEndpoint('/auth/register-admin')).toBe(false)
  })
})

describe('AuthInterceptor.isUnauthorized', () => {
  let interceptor: AuthInterceptor

  beforeEach(() => {
    resetApiClientMock()
    interceptor = new AuthInterceptor(
      createMockTokenStorage(),
      createMockAuthRepo(),
      createMockTokenManager()
    )
  })

  test('exactUnauthorized_Matches', () => {
    expect(interceptor.isUnauthorized('Unauthorized')).toBe(true)
    expect(interceptor.isUnauthorized('unauthorized')).toBe(true)
    expect(interceptor.isUnauthorized('UNAUTHORIZED')).toBe(true)
  })

  test('exactTokenExpired_Matches', () => {
    expect(interceptor.isUnauthorized('token expired')).toBe(true)
    expect(interceptor.isUnauthorized('Token Expired')).toBe(true)
  })

  test('exact401_Matches', () => {
    expect(interceptor.isUnauthorized('401')).toBe(true)
    expect(interceptor.isUnauthorized('401 Unauthorized')).toBe(true)
  })

  test('substringMatch_DoesNotMatch', () => {
    expect(interceptor.isUnauthorized('something unauthorized happened')).toBe(false)
    expect(interceptor.isUnauthorized('user token expired due to inactivity')).toBe(false)
    expect(interceptor.isUnauthorized('error 401 not found')).toBe(false)
  })

  test('otherErrors_DoNotMatch', () => {
    expect(interceptor.isUnauthorized('Not found')).toBe(false)
    expect(interceptor.isUnauthorized('Internal server error')).toBe(false)
    expect(interceptor.isUnauthorized('Bad request')).toBe(false)
  })
})
