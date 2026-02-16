import { AuthTokenManager } from '@/domain/repositories/AuthTokenManager'
import { apiClient } from '@xenios/api-client'

/**
 * ApiAuthTokenManager - manages auth tokens on the API client.
 *
 * This infrastructure implementation encapsulates the apiClient dependency,
 * keeping the presentation layer free from infrastructure concerns.
 */
export class ApiAuthTokenManager implements AuthTokenManager {
  setAuthToken(token: string): void {
    apiClient.setAuthToken(token)
  }

  clearAuthToken(): void {
    apiClient.clearAuthToken()
  }

  restoreToken(token: string | null): void {
    if (token) {
      apiClient.setAuthToken(token)
    }
  }
}
