/**
 * AuthTokens - represents the JWT token pair returned by the backend.
 * Pure domain type with no external dependencies.
 */
export interface AuthTokens {
  access_token: string
  refresh_token: string
}
