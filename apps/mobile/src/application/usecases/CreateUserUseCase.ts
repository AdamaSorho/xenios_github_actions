import { User, CreateUserInput } from '@/domain/entities/User'
import { UserRepository } from '@/domain/repositories/UserRepository'

export class EmailAlreadyExistsError extends Error {
  constructor() {
    super('Email already exists')
    this.name = 'EmailAlreadyExistsError'
  }
}

export class InvalidInputError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'InvalidInputError'
  }
}

/**
 * CreateUserUseCase - application business logic for creating a user.
 */
export class CreateUserUseCase {
  constructor(private readonly userRepo: UserRepository) {}

  async execute(input: CreateUserInput): Promise<User> {
    if (!input.email || input.email.trim() === '') {
      throw new InvalidInputError('Email is required')
    }

    if (!input.name || input.name.trim() === '') {
      throw new InvalidInputError('Name is required')
    }

    // Check if email already exists
    const existing = await this.userRepo.findByEmail(input.email)
    if (existing) {
      throw new EmailAlreadyExistsError()
    }

    return this.userRepo.create(input)
  }
}
