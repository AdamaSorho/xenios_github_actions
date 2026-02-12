import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { AuthRepository } from '@/domain/repositories/AuthRepository'
import { AuthResponse } from '@/domain/entities/AuthUser'

describe('RegisterUseCase', () => {
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
  let useCase: RegisterUseCase

  beforeEach(() => {
    mockAuthRepo = {
      login: jest.fn(),
      register: jest.fn(),
      refresh: jest.fn(),
      logout: jest.fn(),
    }
    useCase = new RegisterUseCase(mockAuthRepo)
  })

  test('execute_ValidInput_ReturnsAuthResponse', async () => {
    mockAuthRepo.register.mockResolvedValue(mockAuthResponse)

    const result = await useCase.execute({
      email: 'coach@example.com',
      password: 'Password123!',
      name: 'Test Coach',
      role: 'coach',
    })

    expect(result).toEqual(mockAuthResponse)
    expect(mockAuthRepo.register).toHaveBeenCalledWith({
      email: 'coach@example.com',
      password: 'Password123!',
      name: 'Test Coach',
      role: 'coach',
    })
  })

  test('execute_EmptyEmail_ThrowsValidationError', async () => {
    await expect(
      useCase.execute({
        email: '',
        password: 'Password123!',
        name: 'Test',
        role: 'coach',
      })
    ).rejects.toThrow('Email is required')
    expect(mockAuthRepo.register).not.toHaveBeenCalled()
  })

  test('execute_InvalidEmail_ThrowsValidationError', async () => {
    await expect(
      useCase.execute({
        email: 'invalid',
        password: 'Password123!',
        name: 'Test',
        role: 'coach',
      })
    ).rejects.toThrow('Invalid email format')
    expect(mockAuthRepo.register).not.toHaveBeenCalled()
  })

  test('execute_ShortPassword_ThrowsValidationError', async () => {
    await expect(
      useCase.execute({
        email: 'coach@example.com',
        password: '12345',
        name: 'Test',
        role: 'coach',
      })
    ).rejects.toThrow('Password must be at least 8 characters')
    expect(mockAuthRepo.register).not.toHaveBeenCalled()
  })

  test('execute_EmptyName_ThrowsValidationError', async () => {
    await expect(
      useCase.execute({
        email: 'coach@example.com',
        password: 'Password123!',
        name: '',
        role: 'coach',
      })
    ).rejects.toThrow('Name is required')
    expect(mockAuthRepo.register).not.toHaveBeenCalled()
  })

  test('execute_InvalidRole_ThrowsValidationError', async () => {
    await expect(
      useCase.execute({
        email: 'coach@example.com',
        password: 'Password123!',
        name: 'Test',
        role: 'invalid',
      })
    ).rejects.toThrow('Role must be one of: coach, client')
    expect(mockAuthRepo.register).not.toHaveBeenCalled()
  })

  test('execute_AdminRole_ThrowsValidationError', async () => {
    await expect(
      useCase.execute({
        email: 'coach@example.com',
        password: 'Password123!',
        name: 'Test',
        role: 'admin',
      })
    ).rejects.toThrow('Role must be one of: coach, client')
    expect(mockAuthRepo.register).not.toHaveBeenCalled()
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockAuthRepo.register.mockRejectedValue(
      new Error('Email already registered')
    )

    await expect(
      useCase.execute({
        email: 'coach@example.com',
        password: 'Password123!',
        name: 'Test',
        role: 'coach',
      })
    ).rejects.toThrow('Email already registered')
  })
})
