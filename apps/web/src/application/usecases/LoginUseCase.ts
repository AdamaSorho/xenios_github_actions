import { AuthResponse, LoginCredentials } from '@/domain/entities/AuthUser'
import { AuthRepository } from '@/domain/repositories/AuthRepository'

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export class LoginUseCase {
  constructor(private readonly authRepo: AuthRepository) {}

  async execute(credentials: LoginCredentials): Promise<AuthResponse> {
    if (!credentials.email) {
      throw new Error('Email is required')
    }
    if (!EMAIL_REGEX.test(credentials.email)) {
      throw new Error('Invalid email format')
    }
    if (!credentials.password) {
      throw new Error('Password is required')
    }

    return this.authRepo.login(credentials)
  }
}
