import { apiClient } from '@xenios/api-client'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { AuthClientConfigurator } from '@/domain/repositories/AuthClientConfigurator'

export class AuthenticatedApiClient implements AuthClientConfigurator {
  private isRefreshing = false
  private refreshPromise: Promise<void> | null = null

  constructor(
    private readonly tokenStorage: TokenStorageRepository,
    private readonly authRepo: AuthRepository
  ) {}

  async configureAuth(): Promise<void> {
    const tokens = await this.tokenStorage.getTokens()
    if (tokens) {
      apiClient.setAuthToken(tokens.accessToken)
    }
  }

  async handleUnauthorized(): Promise<boolean> {
    if (this.isRefreshing) {
      if (this.refreshPromise) {
        await this.refreshPromise
        return true
      }
      return false
    }

    this.isRefreshing = true

    try {
      this.refreshPromise = this.performRefresh()
      await this.refreshPromise
      return true
    } catch {
      await this.tokenStorage.clearTokens()
      apiClient.clearAuthToken()
      return false
    } finally {
      this.isRefreshing = false
      this.refreshPromise = null
    }
  }

  private async performRefresh(): Promise<void> {
    const tokens = await this.tokenStorage.getTokens()
    if (!tokens) {
      throw new Error('No refresh token available')
    }

    const newTokens = await this.authRepo.refreshToken(tokens.refreshToken)
    await this.tokenStorage.saveTokens(newTokens)
    apiClient.setAuthToken(newTokens.accessToken)
  }

  clearAuth(): void {
    apiClient.clearAuthToken()
  }

  setAccessToken(token: string): void {
    apiClient.setAuthToken(token)
  }
}
