/**
 * AuthUser - represents an authenticated user from the backend.
 * Matches the backend User entity JSON serialization.
 * Pure domain type with no external dependencies.
 */
export interface AuthUser {
  id: string
  email: string
  name: string
  role: string
  avatar_url?: string
  created_at: string
  updated_at: string
}

/**
 * LoginCredentials - input for authenticating a user.
 */
export interface LoginCredentials {
  email: string
  password: string
}

/**
 * RegisterInput - input for creating a new user account.
 */
export interface RegisterInput {
  email: string
  password: string
  name: string
  role: string
}

/**
 * AuthResponse - the response from login/register endpoints.
 */
export interface AuthResponse {
  user: AuthUser
  tokens: {
    access_token: string
    refresh_token: string
  }
}
