import {
  CreateUserUseCase,
  EmailAlreadyExistsError,
  InvalidInputError,
} from '@/application/usecases/CreateUserUseCase'
import { User, CreateUserInput } from '@/domain/entities/User'
import { UserRepository } from '@/domain/repositories/UserRepository'

// Mock repository for testing
class MockUserRepository implements UserRepository {
  private users: Map<string, User> = new Map()

  async findById(id: string): Promise<User | null> {
    return this.users.get(id) || null
  }

  async findByEmail(email: string): Promise<User | null> {
    for (const user of this.users.values()) {
      if (user.email === email) {
        return user
      }
    }
    return null
  }

  async create(input: CreateUserInput): Promise<User> {
    const user: User = {
      id: `generated-${Date.now()}`,
      ...input,
      createdAt: new Date(),
      updatedAt: new Date(),
    }
    this.users.set(user.id, user)
    return user
  }

  async update(): Promise<User> {
    throw new Error('Not implemented')
  }

  async delete(): Promise<void> {
    throw new Error('Not implemented')
  }

  setUser(user: User) {
    this.users.set(user.id, user)
  }
}

describe('CreateUserUseCase', () => {
  let mockRepo: MockUserRepository
  let useCase: CreateUserUseCase

  beforeEach(() => {
    mockRepo = new MockUserRepository()
    useCase = new CreateUserUseCase(mockRepo)
  })

  it('should create a new user successfully', async () => {
    const input: CreateUserInput = {
      email: 'new@example.com',
      name: 'New User',
    }

    const result = await useCase.execute(input)

    expect(result.email).toBe(input.email)
    expect(result.name).toBe(input.name)
    expect(result.id).toBeDefined()
  })

  it('should throw EmailAlreadyExistsError when email exists', async () => {
    const existingUser: User = {
      id: 'existing-id',
      email: 'existing@example.com',
      name: 'Existing User',
      createdAt: new Date(),
      updatedAt: new Date(),
    }
    mockRepo.setUser(existingUser)

    const input: CreateUserInput = {
      email: 'existing@example.com',
      name: 'New User',
    }

    await expect(useCase.execute(input)).rejects.toThrow(EmailAlreadyExistsError)
  })

  it('should throw InvalidInputError when email is empty', async () => {
    const input: CreateUserInput = {
      email: '',
      name: 'New User',
    }

    await expect(useCase.execute(input)).rejects.toThrow(InvalidInputError)
  })

  it('should throw InvalidInputError when name is empty', async () => {
    const input: CreateUserInput = {
      email: 'test@example.com',
      name: '',
    }

    await expect(useCase.execute(input)).rejects.toThrow(InvalidInputError)
  })
})
