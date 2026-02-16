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

  test('setTokens_SetsCookieFlag', () => {
    storage.setTokens({
      access_token: 'access-123',
      refresh_token: 'refresh-456',
    })

    expect(document.cookie).toContain('xenios_has_token=1')
  })

  test('clearTokens_ClearsCookieFlag', () => {
    storage.setTokens({
      access_token: 'access-123',
      refresh_token: 'refresh-456',
    })

    storage.clearTokens()

    // After clearing, cookie should be expired (empty or removed)
    expect(document.cookie).not.toContain('xenios_has_token=1')
  })

  test('getUser_NoUser_ReturnsNull', () => {
    expect(storage.getUser()).toBeNull()
  })

  test('setUser_StoresUser_RetrievableAfterward', () => {
    const user = {
      id: 'user-1',
      email: 'coach@example.com',
      name: 'Test Coach',
      role: 'coach',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    storage.setUser(user)

    expect(storage.getUser()).toEqual(user)
  })

  test('clearTokens_AlsoClearsUser', () => {
    const user = {
      id: 'user-1',
      email: 'coach@example.com',
      name: 'Test Coach',
      role: 'coach',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    storage.setUser(user)
    storage.setTokens({
      access_token: 'access-123',
      refresh_token: 'refresh-456',
    })

    storage.clearTokens()

    expect(storage.getUser()).toBeNull()
  })

  test('getUser_InvalidJson_ReturnsNull', () => {
    localStorage.setItem('xenios_user', 'invalid-json')

    expect(storage.getUser()).toBeNull()
  })
})
