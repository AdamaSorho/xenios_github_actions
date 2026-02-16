import { ApiAuthTokenManager } from '@/infrastructure/auth/ApiAuthTokenManager'

// Mock the api-client module
const mockSetAuthToken = jest.fn()
const mockClearAuthToken = jest.fn()
jest.mock('@xenios/api-client', () => ({
  apiClient: {
    setAuthToken: (...args: unknown[]) => mockSetAuthToken(...args),
    clearAuthToken: (...args: unknown[]) => mockClearAuthToken(...args),
  },
}))

describe('ApiAuthTokenManager', () => {
  let tokenManager: ApiAuthTokenManager

  beforeEach(() => {
    jest.clearAllMocks()
    tokenManager = new ApiAuthTokenManager()
  })

  test('setAuthToken_DelegatesToApiClient', () => {
    tokenManager.setAuthToken('test-token')
    expect(mockSetAuthToken).toHaveBeenCalledWith('test-token')
  })

  test('clearAuthToken_DelegatesToApiClient', () => {
    tokenManager.clearAuthToken()
    expect(mockClearAuthToken).toHaveBeenCalled()
  })

  test('restoreToken_WithToken_SetsAuthToken', () => {
    tokenManager.restoreToken('stored-token')
    expect(mockSetAuthToken).toHaveBeenCalledWith('stored-token')
  })

  test('restoreToken_WithNull_DoesNotSetAuthToken', () => {
    tokenManager.restoreToken(null)
    expect(mockSetAuthToken).not.toHaveBeenCalled()
  })
})
