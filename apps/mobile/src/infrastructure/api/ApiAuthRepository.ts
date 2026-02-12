import { AuthTokens } from '@/domain/entities/AuthTokens'
import { LoginCredentials, RegisterCredentials } from '@/domain/entities/AuthCredentials'
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

  async logout(accessToken: string): Promise<void> {
    const previousHeaders = { ...apiClient }
    apiClient.setAuthToken(accessToken)
    try {
      const response = await apiClient.post('/v1/auth/logout')
      if (!response.ok) {
        throw new Error(response.error || 'Logout failed')
      }
    } finally {
      // Token will be managed by the auth context
    }
  }

  private mapAuthResponse(data: BackendAuthResponse): AuthResponse {
    return {
      user: {
        id: data.user.id,
        email: data.user.email,
        name: data.user.name,
        role: data.user.role as AuthResponse['user']['role'],
        avatarUrl: data.user.avatar_url,
        createdAt: data.user.created_at,
        updatedAt: data.user.updated_at,
      },
      tokens: {
        accessToken: data.tokens.access_token,
        refreshToken: data.tokens.refresh_token,
      },
    }
  }
}
