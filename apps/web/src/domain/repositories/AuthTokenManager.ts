/**
 * AuthTokenManager interface - abstracts HTTP client auth token management.
 *
 * NOTE: This is an INTERFACE only - no API client details here!
 * Implementations live in the infrastructure layer.
 * This allows the presentation layer to manage auth headers
 * without directly depending on the API client.
 */
export interface AuthTokenManager {
  setAuthToken(token: string): void
  clearAuthToken(): void
  restoreToken(token: string | null): void
}
