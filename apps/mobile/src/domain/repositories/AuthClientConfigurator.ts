export interface AuthClientConfigurator {
  configureAuth(): Promise<void>
  setAccessToken(token: string): void
  clearAuth(): void
}
