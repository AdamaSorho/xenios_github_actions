import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { AuthResponse } from '@/domain/entities/AuthUser'

describe('LoginUseCase', () => {
  const mockAuthResponse: AuthResponse = {
    user: {
      id: 'user-1',
      email: 'coach@example.com',
      name: 'Test Coach',
      role: 'coach',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    tokens: {
      access_token: 'access-token-123',
      refresh_token: 'refresh-token-456',
    },
  }

  let mockAuthRepo: jest.Mocked<AuthRepository>
  let useCase: LoginUseCase

  beforeEach(() => {
    mockAuthRepo = {
      login: jest.fn(),
      register: jest.fn(),
      refresh: jest.fn(),
      logout: jest.fn(),
    }
    useCase = new LoginUseCase(mockAuthRepo)
  })

  test('execute_ValidCredentials_ReturnsAuthResponse', async () => {
    mockAuthRepo.login.mockResolvedValue(mockAuthResponse)

    const result = await useCase.execute({
      email: 'coach@example.com',
      password: 'password123',
    })

    expect(result).toEqual(mockAuthResponse)
    expect(mockAuthRepo.login).toHaveBeenCalledWith({
      email: 'coach@example.com',
      password: 'password123',
    })
  })

  test('execute_EmptyEmail_ThrowsValidationError', async () => {
    await expect(
      useCase.execute({ email: '', password: 'password123' })
    ).rejects.toThrow('Email is required')
    expect(mockAuthRepo.login).not.toHaveBeenCalled()
  })

  test('execute_EmptyPassword_ThrowsValidationError', async () => {
    await expect(
      useCase.execute({ email: 'coach@example.com', password: '' })
    ).rejects.toThrow('Password is required')
    expect(mockAuthRepo.login).not.toHaveBeenCalled()
  })

  test('execute_InvalidEmail_ThrowsValidationError', async () => {
    await expect(
      useCase.execute({ email: 'not-an-email', password: 'password123' })
    ).rejects.toThrow('Invalid email format')
    expect(mockAuthRepo.login).not.toHaveBeenCalled()
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockAuthRepo.login.mockRejectedValue(new Error('Invalid credentials'))

    await expect(
      useCase.execute({
        email: 'coach@example.com',
        password: 'wrong-password',
      })
    ).rejects.toThrow('Invalid credentials')
  })
})
