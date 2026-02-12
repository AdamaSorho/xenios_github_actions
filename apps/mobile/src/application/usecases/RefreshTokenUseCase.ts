import { AuthTokens } from '@/domain/entities/AuthTokens'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'

export class RefreshTokenUseCase {
  constructor(
    private readonly authRepo: AuthRepository,
    private readonly tokenStorage: TokenStorageRepository
  ) {}

  async execute(): Promise<AuthTokens> {
    const currentTokens = await this.tokenStorage.getTokens()
    if (!currentTokens) {
      throw new Error('No refresh token available')
    }

    const newTokens = await this.authRepo.refreshToken(
      currentTokens.refreshToken
    )
    await this.tokenStorage.saveTokens(newTokens)

    return newTokens
  }
}
