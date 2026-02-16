import { AuthTokens } from '@/domain/entities/AuthTokens'
import { AuthUser } from '@/domain/entities/AuthUser'
import { TokenStorage } from '@/domain/repositories/TokenStorage'

const ACCESS_TOKEN_KEY = 'xenios_access_token'
const REFRESH_TOKEN_KEY = 'xenios_refresh_token'
const USER_KEY = 'xenios_user'
const AUTH_COOKIE_NAME = 'xenios_has_token'

/**
 * LocalTokenStorage - stores JWT tokens in localStorage.
 *
 * SECURITY NOTE: localStorage is accessible to any JavaScript running on the
 * same origin, making tokens vulnerable to XSS attacks. This is a known
 * trade-off chosen for simplicity in this skeleton implementation. For
 * production, consider migrating to httpOnly cookies set by the backend,
 * which are immune to XSS token theft.
 *
 * Additionally sets a non-httpOnly cookie flag (`xenios_has_token`) so that
 * Next.js middleware (which runs on the server and cannot access localStorage)
 * can detect whether the user has an active session for route protection.
 * This cookie contains no sensitive data — only a flag ("1").
 */
export class LocalTokenStorage implements TokenStorage {
  private get secureFlag(): string {
    if (typeof window !== 'undefined' && window.location.protocol === 'https:') {
      return '; Secure'
    }
    return ''
  }

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
    document.cookie = `${AUTH_COOKIE_NAME}=1; path=/; SameSite=Lax${this.secureFlag}`
  }

  clearTokens(): void {
    if (typeof window === 'undefined') return
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    document.cookie = `${AUTH_COOKIE_NAME}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax${this.secureFlag}`
  }

  getUser(): AuthUser | null {
    if (typeof window === 'undefined') return null
    const json = localStorage.getItem(USER_KEY)
    if (!json) return null
    try {
      return JSON.parse(json) as AuthUser
    } catch {
      return null
    }
  }

  setUser(user: AuthUser): void {
    if (typeof window === 'undefined') return
    localStorage.setItem(USER_KEY, JSON.stringify(user))
  }
}
