import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'
import { AuthRepository } from '@/domain/repositories/AuthRepository'

describe('LogoutUseCase', () => {
  let mockAuthRepo: jest.Mocked<AuthRepository>
  let useCase: LogoutUseCase

  beforeEach(() => {
    mockAuthRepo = {
      login: jest.fn(),
      register: jest.fn(),
      refresh: jest.fn(),
      logout: jest.fn(),
    }
    useCase = new LogoutUseCase(mockAuthRepo)
  })

  test('execute_Success_CallsLogout', async () => {
    mockAuthRepo.logout.mockResolvedValue(undefined)

    await useCase.execute()

    expect(mockAuthRepo.logout).toHaveBeenCalledTimes(1)
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockAuthRepo.logout.mockRejectedValue(new Error('Network error'))

    await expect(useCase.execute()).rejects.toThrow('Network error')
  })
})
