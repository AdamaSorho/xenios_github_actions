import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'

export interface AuthState {
  isAuthenticated: boolean
  accessToken: string | null
}

export class GetAuthStateUseCase {
  constructor(private readonly tokenStorage: TokenStorageRepository) {}

  async execute(): Promise<AuthState> {
    const tokens = await this.tokenStorage.getTokens()

    if (!tokens) {
      return { isAuthenticated: false, accessToken: null }
    }

    return {
      isAuthenticated: true,
      accessToken: tokens.accessToken,
    }
  }
}
