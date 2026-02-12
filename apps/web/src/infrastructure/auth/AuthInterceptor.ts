import { apiClient } from '@xenios/api-client'
import { TokenStorage } from '@/domain/repositories/TokenStorage'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { AuthTokenManager } from '@/domain/repositories/AuthTokenManager'

interface ApiResponse<T> {
  data: T | null
  error: string | null
  ok: boolean
}

/**
 * AuthInterceptor - wraps API calls with automatic token refresh on 401.
 *
 * Uses a wrapper pattern that saves and delegates to the original apiClient
 * methods. When an API call returns a 401/unauthorized response, the
 * interceptor automatically attempts to refresh the access token. If refresh
 * succeeds, the original request is retried with the new token. If refresh
 * fails, tokens are cleared.
 *
 * Concurrent 401 responses are handled via a request queue: if a refresh is
 * already in progress, subsequent 401 responses wait for the refresh to
 * complete and then retry their original request.
 */
export class AuthInterceptor {
  private refreshPromise: Promise<boolean> | null = null
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

  private static readonly AUTH_ENDPOINTS = [
    '/auth/login',
    '/auth/register',
    '/auth/refresh',
  ]

  install(): void {
    const self = this

    apiClient.get = async function <T>(path: string) {
      return self.withRefresh<T>(() => self.originalGet<T>(path), path)
    }

    apiClient.post = async function <T>(path: string, body?: unknown) {
      if (self.isAuthEndpoint(path)) {
        return self.originalPost<T>(path, body)
      }
      return self.withRefresh<T>(() => self.originalPost<T>(path, body), path)
    }

    apiClient.put = async function <T>(path: string, body?: unknown) {
      return self.withRefresh<T>(() => self.originalPut<T>(path, body), path)
    }

    apiClient.delete = async function <T>(path: string) {
      return self.withRefresh<T>(() => self.originalDelete<T>(path), path)
    }
  }

  private async withRefresh<T>(
    request: () => Promise<ApiResponse<T>>,
    _path: string
  ): Promise<ApiResponse<T>> {
    const response = await request()
    if (!response.ok && response.error && this.isUnauthorized(response.error)) {
      const refreshed = await this.tryRefresh()
      if (refreshed) return request()
    }
    return response
  }

  isAuthEndpoint(path: string): boolean {
    return AuthInterceptor.AUTH_ENDPOINTS.some(
      (endpoint) => path === endpoint || path.endsWith(endpoint)
    )
  }

  isUnauthorized(error: string): boolean {
    const lower = error.toLowerCase()
    return lower === 'unauthorized' ||
           lower === 'token expired' ||
           lower === '401' ||
           lower === '401 unauthorized'
  }

  private async tryRefresh(): Promise<boolean> {
    if (this.refreshPromise) {
      return this.refreshPromise
    }

    this.refreshPromise = this.doRefresh()

    try {
      return await this.refreshPromise
    } finally {
      this.refreshPromise = null
    }
  }

  private async doRefresh(): Promise<boolean> {
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
    }
  }

  private clearAuth(): void {
    this.tokenStorage.clearTokens()
    this.tokenManager.clearAuthToken()
  }
}
