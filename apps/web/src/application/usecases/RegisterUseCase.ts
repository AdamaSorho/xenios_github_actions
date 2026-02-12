import { AuthResponse, RegisterInput } from '@/domain/entities/AuthUser'
import { AuthRepository } from '@/domain/repositories/AuthRepository'

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const VALID_ROLES = ['coach', 'client', 'admin']

export class RegisterUseCase {
  constructor(private readonly authRepo: AuthRepository) {}

  async execute(input: RegisterInput): Promise<AuthResponse> {
    if (!input.email) {
      throw new Error('Email is required')
    }
    if (!EMAIL_REGEX.test(input.email)) {
      throw new Error('Invalid email format')
    }
    if (!input.password || input.password.length < 8) {
      throw new Error('Password must be at least 8 characters')
    }
    if (!input.name) {
      throw new Error('Name is required')
    }
    if (!VALID_ROLES.includes(input.role)) {
      throw new Error('Role must be one of: coach, client, admin')
    }

    return this.authRepo.register(input)
  }
}
