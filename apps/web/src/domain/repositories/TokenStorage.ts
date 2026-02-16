import { AuthTokens } from '../entities/AuthTokens'
import { AuthUser } from '../entities/AuthUser'

/**
 * TokenStorage interface - defines auth session persistence operations.
 *
 * NOTE: This is an INTERFACE only - no storage implementation details here!
 * Implementations live in the infrastructure layer.
 */
export interface TokenStorage {
  getAccessToken(): string | null
  getRefreshToken(): string | null
  setTokens(tokens: AuthTokens): void
  clearTokens(): void
  getUser(): AuthUser | null
  setUser(user: AuthUser): void
}
