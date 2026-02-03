import { User } from '@/domain/entities/User'
import { UserRepository } from '@/domain/repositories/UserRepository'

/**
 * GetUserUseCase - application business logic for retrieving a user.
 */
export class GetUserUseCase {
  constructor(private readonly userRepo: UserRepository) {}

  async execute(id: string): Promise<User | null> {
    return this.userRepo.findById(id)
  }
}
