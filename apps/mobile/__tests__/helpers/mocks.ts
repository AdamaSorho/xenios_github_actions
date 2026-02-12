import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { AuthUser } from '@/domain/entities/AuthUser'
import { AuthTokens } from '@/domain/entities/AuthTokens'

export function createMockAuthRepo(): jest.Mocked<AuthRepository> {
  return {
    login: jest.fn(),
    register: jest.fn(),
    refreshToken: jest.fn(),
    logout: jest.fn(),
    getCurrentUser: jest.fn(),
  }
}

export function createMockTokenStorage(): jest.Mocked<TokenStorageRepository> {
  return {
    saveTokens: jest.fn(),
    getTokens: jest.fn(),
    clearTokens: jest.fn(),
  }
}

export const mockUser: AuthUser = {
  id: 'user-1',
  email: 'test@example.com',
  name: 'Test User',
  role: 'coach',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

export const mockTokens: AuthTokens = {
  accessToken: 'access-token-123',
  refreshToken: 'refresh-token-456',
}
