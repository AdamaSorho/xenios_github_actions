import { RegisterUseCase, ValidationError } from '@/application/usecases/RegisterUseCase'
import { AuthRepository, AuthResponse } from '@/domain/repositories/AuthRepository'
import { TokenStorageRepository } from '@/domain/repositories/TokenStorageRepository'
import { AuthUser } from '@/domain/entities/AuthUser'
import { AuthTokens } from '@/domain/entities/AuthTokens'

const mockUser: AuthUser = {
  id: 'user-1',
  email: 'new@example.com',
  name: 'New User',
  role: 'client',
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

describe('RegisterUseCase', () => {
  let useCase: RegisterUseCase
  let authRepo: jest.Mocked<AuthRepository>
  let tokenStorage: jest.Mocked<TokenStorageRepository>

  const validCredentials = {
    email: 'new@example.com',
    password: 'password123',
    name: 'New User',
    role: 'client' as const,
  }

  beforeEach(() => {
    authRepo = createMockAuthRepo()
    tokenStorage = createMockTokenStorage()
    useCase = new RegisterUseCase(authRepo, tokenStorage)
  })

  it('should register successfully with valid credentials', async () => {
    authRepo.register.mockResolvedValue(mockAuthResponse)
    tokenStorage.saveTokens.mockResolvedValue(undefined)

    const result = await useCase.execute(validCredentials)

    expect(result.user).toEqual(mockUser)
    expect(authRepo.register).toHaveBeenCalledWith(validCredentials)
    expect(tokenStorage.saveTokens).toHaveBeenCalledWith(mockTokens)
  })

  it('should throw ValidationError for empty email', async () => {
    await expect(
      useCase.execute({ ...validCredentials, email: '' })
    ).rejects.toThrow(ValidationError)
  })

  it('should throw ValidationError for invalid email', async () => {
    await expect(
      useCase.execute({ ...validCredentials, email: 'bad' })
    ).rejects.toThrow(ValidationError)
  })

  it('should throw ValidationError for short password', async () => {
    await expect(
      useCase.execute({ ...validCredentials, password: 'short' })
    ).rejects.toThrow(ValidationError)
  })

  it('should throw ValidationError for empty name', async () => {
    await expect(
      useCase.execute({ ...validCredentials, name: '' })
    ).rejects.toThrow(ValidationError)
  })

  it('should throw ValidationError for missing role', async () => {
    await expect(
      useCase.execute({ ...validCredentials, role: '' as any })
    ).rejects.toThrow(ValidationError)
  })

  it('should not call authRepo when validation fails', async () => {
    try {
      await useCase.execute({ ...validCredentials, email: '' })
    } catch {
      // expected
    }
    expect(authRepo.register).not.toHaveBeenCalled()
  })

  it('should propagate auth repo errors', async () => {
    authRepo.register.mockRejectedValue(new Error('Email already exists'))

    await expect(useCase.execute(validCredentials)).rejects.toThrow(
      'Email already exists'
    )
  })
})
