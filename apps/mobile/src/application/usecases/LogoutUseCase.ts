import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'

export class LogoutUseCase {
  constructor(
    private readonly authRepo: AuthRepository,
    private readonly tokenStorage: TokenStorageRepository
  ) {}

  async execute(): Promise<void> {
    const tokens = await this.tokenStorage.getTokens()

    if (tokens) {
      try {
        await this.authRepo.logout(tokens.accessToken)
      } catch {
        // Best-effort logout on server; always clear local tokens
      }
    }

    await this.tokenStorage.clearTokens()
  }
}
