import { AuthUser } from '../entities/AuthUser'
import { AuthTokens } from '../entities/AuthTokens'
import { LoginCredentials, RegisterCredentials } from '../entities/AuthCredentials'

export interface AuthResponse {
  user: AuthUser
  tokens: AuthTokens
}

export interface AuthRepository {
  login(credentials: LoginCredentials): Promise<AuthResponse>
  register(credentials: RegisterCredentials): Promise<AuthResponse>
  refreshToken(refreshToken: string): Promise<AuthTokens>
  logout(accessToken: string): Promise<void>
}
