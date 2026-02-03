import { GetUserUseCase } from '@/application/usecases/GetUserUseCase'
import { User } from '@/domain/entities/User'
import { UserRepository } from '@/domain/repositories/UserRepository'

// Mock repository for testing
class MockUserRepository implements UserRepository {
  private users: Map<string, User> = new Map()

  setUser(user: User) {
    this.users.set(user.id, user)
  }

  async findById(id: string): Promise<User | null> {
    return this.users.get(id) || null
  }

  async findByEmail(_email: string): Promise<User | null> {
    return null
  }

  async create(): Promise<User> {
    throw new Error('Not implemented')
  }

  async update(): Promise<User> {
    throw new Error('Not implemented')
  }

  async delete(): Promise<void> {
    throw new Error('Not implemented')
  }
}

describe('GetUserUseCase', () => {
  let mockRepo: MockUserRepository
  let useCase: GetUserUseCase

  beforeEach(() => {
    mockRepo = new MockUserRepository()
    useCase = new GetUserUseCase(mockRepo)
  })

  it('should return user when found', async () => {
    const user: User = {
      id: 'test-id',
      email: 'test@example.com',
      name: 'Test User',
      createdAt: new Date(),
      updatedAt: new Date(),
    }
    mockRepo.setUser(user)

    const result = await useCase.execute('test-id')

    expect(result).toEqual(user)
  })

  it('should return null when user not found', async () => {
    const result = await useCase.execute('non-existent-id')

    expect(result).toBeNull()
  })
})
