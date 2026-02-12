import { AuthTokens } from '@/domain/entities/AuthTokens'
import { AuthRepository } from '@/domain/repositories/AuthRepository'

export class RefreshTokenUseCase {
  constructor(private readonly authRepo: AuthRepository) {}

  async execute(refreshToken: string): Promise<AuthTokens> {
    if (!refreshToken) {
      throw new Error('Refresh token is required')
    }

    return this.authRepo.refresh(refreshToken)
  }
}
