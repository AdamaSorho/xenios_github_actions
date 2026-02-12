import { AuthTokens } from '@/domain/entities/AuthTokens'
import { TokenStorage } from '@/domain/repositories/TokenStorage'

const ACCESS_TOKEN_KEY = 'xenios_access_token'
const REFRESH_TOKEN_KEY = 'xenios_refresh_token'

export class LocalTokenStorage implements TokenStorage {
  getAccessToken(): string | null {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(ACCESS_TOKEN_KEY)
  }

  getRefreshToken(): string | null {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(REFRESH_TOKEN_KEY)
  }

  setTokens(tokens: AuthTokens): void {
    if (typeof window === 'undefined') return
    localStorage.setItem(ACCESS_TOKEN_KEY, tokens.access_token)
    localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token)
  }

  clearTokens(): void {
    if (typeof window === 'undefined') return
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  }
}
