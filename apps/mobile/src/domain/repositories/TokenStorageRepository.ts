import { AuthTokens } from '../entities/AuthTokens'

export interface TokenStorageRepository {
  saveTokens(tokens: AuthTokens): Promise<void>
  getTokens(): Promise<AuthTokens | null>
  clearTokens(): Promise<void>
}
