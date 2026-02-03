import { User, CreateUserInput } from '../entities/User'

/**
 * UserRepository interface - defines data access operations for users.
 *
 * NOTE: This is an INTERFACE only - no API client imports here!
 * Implementations live in the infrastructure layer.
 */
export interface UserRepository {
  findById(id: string): Promise<User | null>
  findByEmail(email: string): Promise<User | null>
  create(input: CreateUserInput): Promise<User>
  update(id: string, user: Partial<User>): Promise<User>
  delete(id: string): Promise<void>
}
