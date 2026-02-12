import { LoginUseCase, ValidationError } from '@/application/usecases/LoginUseCase'
import { AuthRepository, AuthResponse } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { AuthUser } from '@/domain/entities/AuthUser'
import { AuthTokens } from '@/domain/entities/AuthTokens'

const mockUser: AuthUser = {
  id: 'user-1',
  email: 'test@example.com',
  name: 'Test User',
  role: 'coach',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-01T00:00:00Z',
}

const mockTokens: AuthTokens = {
  accessToken: 'access-token-123',
  refreshToken: 'refresh-token-456',
}

const mockAuthResponse: AuthResponse = {
  user: mockUser,
  tokens: mockTokens,
}

function createMockAuthRepo(): jest.Mocked<AuthRepository> {
  return {
    login: jest.fn(),
    register: jest.fn(),
    refreshToken: jest.fn(),
    logout: jest.fn(),
  }
}

function createMockTokenStorage(): jest.Mocked<TokenStorageRepository> {
  return {
    saveTokens: jest.fn(),
    getTokens: jest.fn(),
    clearTokens: jest.fn(),
  }
}

describe('LoginUseCase', () => {
  let useCase: LoginUseCase
  let authRepo: jest.Mocked<AuthRepository>
  let tokenStorage: jest.Mocked<TokenStorageRepository>

  beforeEach(() => {
    authRepo = createMockAuthRepo()
    tokenStorage = createMockTokenStorage()
    useCase = new LoginUseCase(authRepo, tokenStorage)
  })

  it('should login successfully with valid credentials', async () => {
    authRepo.login.mockResolvedValue(mockAuthResponse)
    tokenStorage.saveTokens.mockResolvedValue(undefined)

    const result = await useCase.execute({
      email: 'test@example.com',
      password: 'password123',
    })

    expect(result.user).toEqual(mockUser)
    expect(authRepo.login).toHaveBeenCalledWith({
      email: 'test@example.com',
      password: 'password123',
    })
    expect(tokenStorage.saveTokens).toHaveBeenCalledWith(mockTokens)
  })

  it('should throw ValidationError for empty email', async () => {
    await expect(
      useCase.execute({ email: '', password: 'password123' })
    ).rejects.toThrow(ValidationError)
  })

  it('should throw ValidationError for invalid email format', async () => {
    await expect(
      useCase.execute({ email: 'invalid', password: 'password123' })
    ).rejects.toThrow(ValidationError)
  })

  it('should throw ValidationError for empty password', async () => {
    await expect(
      useCase.execute({ email: 'test@example.com', password: '' })
    ).rejects.toThrow(ValidationError)
  })

  it('should not call authRepo when validation fails', async () => {
    try {
      await useCase.execute({ email: '', password: '' })
    } catch {
      // expected
    }
    expect(authRepo.login).not.toHaveBeenCalled()
    expect(tokenStorage.saveTokens).not.toHaveBeenCalled()
  })

  it('should propagate auth repo errors', async () => {
    authRepo.login.mockRejectedValue(new Error('Invalid credentials'))

    await expect(
      useCase.execute({ email: 'test@example.com', password: 'password123' })
    ).rejects.toThrow('Invalid credentials')
  })

  it('should include validation errors in ValidationError', async () => {
    try {
      await useCase.execute({ email: '', password: '' })
    } catch (error) {
      expect(error).toBeInstanceOf(ValidationError)
      expect((error as ValidationError).errors).toContain('Email is required')
      expect((error as ValidationError).errors).toContain('Password is required')
    }
  })
})
