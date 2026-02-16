import { AuthTokens } from '@/domain/entities/AuthTokens'
import { LoginCredentials, RegisterCredentials } from '@/domain/entities/AuthCredentials'
import { isValidRole } from '@/domain/entities/AuthUser'
import { AuthRepository, AuthResponse } from '@/domain/repositories/AuthRepository'
import { apiClient } from '@xenios/api-client'

interface BackendAuthResponse {
  user: {
    id: string
    email: string
    name: string
    role: string
    avatar_url?: string
    created_at: string
    updated_at: string
  }
  tokens: {
    access_token: string
    refresh_token: string
  }
}

interface BackendTokensResponse {
  access_token: string
  refresh_token: string
}

export class ApiAuthRepository implements AuthRepository {
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    const response = await apiClient.post<BackendAuthResponse>(
      '/auth/login',
      {
        email: credentials.email,
        password: credentials.password,
      }
    )

    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Login failed')
    }

    return this.mapAuthResponse(response.data)
  }

  async register(credentials: RegisterCredentials): Promise<AuthResponse> {
    const response = await apiClient.post<BackendAuthResponse>(
      '/auth/register',
      {
        email: credentials.email,
        password: credentials.password,
        name: credentials.name,
        role: credentials.role,
      }
    )

    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Registration failed')
    }

    return this.mapAuthResponse(response.data)
  }

  async refreshToken(refreshToken: string): Promise<AuthTokens> {
    const response = await apiClient.post<BackendTokensResponse>(
      '/auth/refresh',
      { refresh_token: refreshToken }
    )

    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Token refresh failed')
    }

    return {
      accessToken: response.data.access_token,
      refreshToken: response.data.refresh_token,
    }
  }

  async logout(_accessToken: string): Promise<void> {
    const response = await apiClient.post('/auth/logout')
    if (!response.ok) {
      throw new Error(response.error || 'Logout failed')
    }
  }

  async getCurrentUser(): Promise<AuthResponse['user']> {
    const response = await apiClient.get<BackendAuthResponse['user']>('/auth/me')

    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Failed to get current user')
    }

    return this.mapUser(response.data)
  }

  private mapUser(data: BackendAuthResponse['user']): AuthResponse['user'] {
    if (!isValidRole(data.role)) {
      throw new Error(`Invalid role received from server: ${data.role}`)
    }

    return {
      id: data.id,
      email: data.email,
      name: data.name,
      role: data.role,
      avatarUrl: data.avatar_url,
      createdAt: data.created_at,
      updatedAt: data.updated_at,
    }
  }

  private mapAuthResponse(data: BackendAuthResponse): AuthResponse {
    return {
      user: this.mapUser(data.user),
      tokens: {
        accessToken: data.tokens.access_token,
        refreshToken: data.tokens.refresh_token,
      },
    }
  }
}
