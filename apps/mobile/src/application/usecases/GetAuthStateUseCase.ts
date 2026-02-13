import { AuthUser } from '@/domain/entities/AuthUser'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'

export interface AuthState {
  isAuthenticated: boolean
  accessToken: string | null
  user: AuthUser | null
}

export class GetAuthStateUseCase {
  constructor(
    private readonly tokenStorage: TokenStorageRepository,
    private readonly authRepo: AuthRepository
  ) {}

  async execute(): Promise<AuthState> {
    const tokens = await this.tokenStorage.getTokens()

    if (!tokens) {
      return { isAuthenticated: false, accessToken: null, user: null }
    }

    try {
      const user = await this.authRepo.getCurrentUser()
      return {
        isAuthenticated: true,
        accessToken: tokens.accessToken,
        user,
      }
    } catch {
      return {
        isAuthenticated: false,
        accessToken: tokens.accessToken,
        user: null,
      }
    }
  }
}
