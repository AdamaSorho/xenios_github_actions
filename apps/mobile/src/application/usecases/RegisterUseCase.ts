import { AuthUser } from '@/domain/entities/AuthUser'
import {
  RegisterCredentials,
  validateRegisterCredentials,
} from '@/domain/entities/AuthCredentials'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { ValidationError } from '@/application/errors/ValidationError'

export interface RegisterOutput {
  user: AuthUser
  accessToken: string
}

export class RegisterUseCase {
  constructor(
    private readonly authRepo: AuthRepository,
    private readonly tokenStorage: TokenStorageRepository
  ) {}

  async execute(credentials: RegisterCredentials): Promise<RegisterOutput> {
    const validation = validateRegisterCredentials(credentials)
    if (!validation.valid) {
      throw new ValidationError(validation.errors)
    }

    const response = await this.authRepo.register(credentials)
    await this.tokenStorage.saveTokens(response.tokens)

    return { user: response.user, accessToken: response.tokens.accessToken }
  }
}
