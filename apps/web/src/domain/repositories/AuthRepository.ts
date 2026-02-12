import {
  AuthResponse,
  LoginCredentials,
  RegisterInput,
} from '../entities/AuthUser'
import { AuthTokens } from '../entities/AuthTokens'

/**
 * AuthRepository interface - defines auth operations.
 *
 * NOTE: This is an INTERFACE only - no API client imports here!
 * Implementations live in the infrastructure layer.
 */
export interface AuthRepository {
  login(credentials: LoginCredentials): Promise<AuthResponse>
  register(input: RegisterInput): Promise<AuthResponse>
  refresh(refreshToken: string): Promise<AuthTokens>
  logout(): Promise<void>
}
