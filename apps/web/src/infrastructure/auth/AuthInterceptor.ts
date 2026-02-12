import { apiClient } from '@xenios/api-client'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { AuthTokenManager } from '@/domain/repositories/AuthTokenManager'

/**
 * AuthInterceptor - wraps API calls with automatic token refresh on 401.
 *
 * When an API call returns a 401 status, the interceptor automatically
 * attempts to refresh the access token using the stored refresh token.
 * If refresh succeeds, the original request is retried with the new token.
 * If refresh fails, tokens are cleared and the user is redirected to login.
 */
export class AuthInterceptor {
  private isRefreshing = false
  private originalGet: typeof apiClient.get
  private originalPost: typeof apiClient.post
  private originalPut: typeof apiClient.put
  private originalDelete: typeof apiClient.delete

  constructor(
    private readonly tokenStorage: TokenStorage,
    private readonly authRepo: AuthRepository,
    private readonly tokenManager: AuthTokenManager
  ) {
    this.originalGet = apiClient.get.bind(apiClient)
    this.originalPost = apiClient.post.bind(apiClient)
    this.originalPut = apiClient.put.bind(apiClient)
    this.originalDelete = apiClient.delete.bind(apiClient)
  }

  install(): void {
    const self = this

    apiClient.get = async function <T>(path: string) {
      const response = await self.originalGet<T>(path)
      if (!response.ok && response.error && self.isUnauthorized(response.error)) {
        const refreshed = await self.tryRefresh()
        if (refreshed) return self.originalGet<T>(path)
      }
      return response
    }

    apiClient.post = async function <T>(path: string, body?: unknown) {
      if (self.isAuthEndpoint(path)) {
        return self.originalPost<T>(path, body)
      }
      const response = await self.originalPost<T>(path, body)
      if (!response.ok && response.error && self.isUnauthorized(response.error)) {
        const refreshed = await self.tryRefresh()
        if (refreshed) return self.originalPost<T>(path, body)
      }
      return response
    }

    apiClient.put = async function <T>(path: string, body?: unknown) {
      const response = await self.originalPut<T>(path, body)
      if (!response.ok && response.error && self.isUnauthorized(response.error)) {
        const refreshed = await self.tryRefresh()
        if (refreshed) return self.originalPut<T>(path, body)
      }
      return response
    }

    apiClient.delete = async function <T>(path: string) {
      const response = await self.originalDelete<T>(path)
      if (!response.ok && response.error && self.isUnauthorized(response.error)) {
        const refreshed = await self.tryRefresh()
        if (refreshed) return self.originalDelete<T>(path)
      }
      return response
    }
  }

  private isAuthEndpoint(path: string): boolean {
    return path.includes('/auth/login') ||
           path.includes('/auth/register') ||
           path.includes('/auth/refresh')
  }

  private isUnauthorized(error: string): boolean {
    return error.toLowerCase().includes('unauthorized') ||
           error.toLowerCase().includes('token expired') ||
           error.includes('401')
  }

  private async tryRefresh(): Promise<boolean> {
    if (this.isRefreshing) return false
    this.isRefreshing = true

    try {
      const refreshToken = this.tokenStorage.getRefreshToken()
      if (!refreshToken) {
        this.clearAuth()
        return false
      }

      const newTokens = await this.authRepo.refresh(refreshToken)
      this.tokenStorage.setTokens(newTokens)
      this.tokenManager.setAuthToken(newTokens.access_token)
      return true
    } catch {
      this.clearAuth()
      return false
    } finally {
      this.isRefreshing = false
    }
  }

  private clearAuth(): void {
    this.tokenStorage.clearTokens()
    this.tokenManager.clearAuthToken()
  }
}
