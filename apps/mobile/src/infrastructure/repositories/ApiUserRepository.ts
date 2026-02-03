import { User, CreateUserInput } from '@/domain/entities/User'
import { UserRepository } from '@/domain/repositories/UserRepository'
import { apiClient } from '@xenios/api-client'

/**
 * ApiUserRepository - implementation of UserRepository using Backend API.
 *
 * IMPORTANT: Mobile NEVER accesses the database directly.
 * All data operations go through the Backend API.
 */
export class ApiUserRepository implements UserRepository {
  async findById(id: string): Promise<User | null> {
    const response = await apiClient.get<User>(`/users/${id}`)
    if (!response.ok) {
      return null
    }
    return this.parseUser(response.data)
  }

  async findByEmail(email: string): Promise<User | null> {
    const response = await apiClient.get<User>(`/users/by-email/${encodeURIComponent(email)}`)
    if (!response.ok) {
      return null
    }
    return this.parseUser(response.data)
  }

  async create(input: CreateUserInput): Promise<User> {
    const response = await apiClient.post<User>('/users', input)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Failed to create user')
    }
    return this.parseUser(response.data)!
  }

  async update(id: string, user: Partial<User>): Promise<User> {
    const response = await apiClient.put<User>(`/users/${id}`, user)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Failed to update user')
    }
    return this.parseUser(response.data)!
  }

  async delete(id: string): Promise<void> {
    const response = await apiClient.delete(`/users/${id}`)
    if (!response.ok) {
      throw new Error(response.error || 'Failed to delete user')
    }
  }

  private parseUser(data: User | null): User | null {
    if (!data) return null
    return {
      ...data,
      createdAt: new Date(data.createdAt),
      updatedAt: new Date(data.updatedAt),
    }
  }
}
