import { AuthUser } from '@/domain/entities/AuthUser'
import { LoginCredentials, validateLoginCredentials } from '@/domain/entities/AuthCredentials'
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

export interface LoginOutput {
  user: AuthUser
}

export class LoginUseCase {
  constructor(
    private readonly authRepo: AuthRepository,
    private readonly tokenStorage: TokenStorageRepository
  ) {}

  async execute(credentials: LoginCredentials): Promise<LoginOutput> {
    const validation = validateLoginCredentials(credentials)
    if (!validation.valid) {
      throw new ValidationError(validation.errors)
    }

    const response = await this.authRepo.login(credentials)
    await this.tokenStorage.saveTokens(response.tokens)

    return { user: response.user }
  }
}
