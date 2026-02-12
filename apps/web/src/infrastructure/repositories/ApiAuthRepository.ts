import {
  AuthResponse,
  LoginCredentials,
  RegisterInput,
} from '@/domain/entities/AuthUser'
import { AuthTokens } from '@/domain/entities/AuthTokens'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { apiClient } from '@xenios/api-client'

/**
 * ApiAuthRepository - implementation of AuthRepository using Backend API.
 *
 * IMPORTANT: Web NEVER accesses the database directly.
 * All auth operations go through the Backend API.
 */
export class ApiAuthRepository implements AuthRepository {
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/login', credentials)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Login failed')
    }
    return response.data
  }

  async register(input: RegisterInput): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>('/auth/register', input)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Registration failed')
    }
    return response.data
  }

  async refresh(refreshToken: string): Promise<AuthTokens> {
    const response = await apiClient.post<AuthTokens>('/auth/refresh', {
      refresh_token: refreshToken,
    })
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Token refresh failed')
    }
    return response.data
  }

  async logout(): Promise<void> {
    const response = await apiClient.post('/auth/logout')
    if (!response.ok) {
      throw new Error(response.error || 'Logout failed')
    }
  }
}
