import { LocalTokenStorage } from '@/infrastructure/auth/LocalTokenStorage'

describe('LocalTokenStorage', () => {
  let storage: LocalTokenStorage

  beforeEach(() => {
    storage = new LocalTokenStorage()
    // Clear any stored data
    localStorage.clear()
  })

  test('getAccessToken_NoToken_ReturnsNull', () => {
    expect(storage.getAccessToken()).toBeNull()
  })

  test('getRefreshToken_NoToken_ReturnsNull', () => {
    expect(storage.getRefreshToken()).toBeNull()
  })

  test('setTokens_StoresTokens_RetrievableAfterward', () => {
    storage.setTokens({
      access_token: 'access-123',
      refresh_token: 'refresh-456',
    })

    expect(storage.getAccessToken()).toBe('access-123')
    expect(storage.getRefreshToken()).toBe('refresh-456')
  })

  test('clearTokens_RemovesAllTokens', () => {
    storage.setTokens({
      access_token: 'access-123',
      refresh_token: 'refresh-456',
    })

    storage.clearTokens()

    expect(storage.getAccessToken()).toBeNull()
    expect(storage.getRefreshToken()).toBeNull()
  })

  test('setTokens_OverwritesPreviousTokens', () => {
    storage.setTokens({
      access_token: 'old-access',
      refresh_token: 'old-refresh',
    })

    storage.setTokens({
      access_token: 'new-access',
      refresh_token: 'new-refresh',
    })

    expect(storage.getAccessToken()).toBe('new-access')
    expect(storage.getRefreshToken()).toBe('new-refresh')
  })
})
