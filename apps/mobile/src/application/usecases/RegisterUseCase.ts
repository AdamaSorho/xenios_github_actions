import { AuthUser } from '@/domain/entities/AuthUser'
import {
  RegisterCredentials,
  validateRegisterCredentials,
} from '@/domain/entities/AuthCredentials'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'

export class ValidationError extends Error {
  public readonly errors: string[]
  constructor(errors: string[]) {
    super(errors.join(', '))
    this.name = 'ValidationError'
    this.errors = errors
  }
}

export interface RegisterOutput {
  user: AuthUser
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

    return { user: response.user }
  }
}
