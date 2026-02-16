import { ApiAuthRepository } from '@/infrastructure/repositories/ApiAuthRepository'
import { apiClient } from '@xenios/api-client'

jest.mock('@xenios/api-client', () => ({
  apiClient: {
    post: jest.fn(),
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('ApiAuthRepository branch coverage', () => {
  let repo: ApiAuthRepository

  beforeEach(() => {
    jest.clearAllMocks()
    repo = new ApiAuthRepository()
  })

  test('login_NoData_ThrowsDefaultError', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: null,
      error: null,
      ok: false,
    })

    await expect(
      repo.login({ email: 'a@b.com', password: 'pass' })
    ).rejects.toThrow('Login failed')
  })

  test('register_NoData_ThrowsDefaultError', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: null,
      error: null,
      ok: false,
    })

    await expect(
      repo.register({
        email: 'a@b.com',
        password: 'pass',
        name: 'N',
        role: 'coach',
      })
    ).rejects.toThrow('Registration failed')
  })

  test('refresh_NoData_ThrowsDefaultError', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: null,
      error: null,
      ok: false,
    })

    await expect(repo.refresh('token')).rejects.toThrow('Token refresh failed')
  })

  test('logout_NoData_ThrowsDefaultError', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: null,
      error: null,
      ok: false,
    })

    await expect(repo.logout()).rejects.toThrow('Logout failed')
  })

  test('login_OkButNullData_ThrowsError', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: null,
      error: null,
      ok: true,
    })

    await expect(
      repo.login({ email: 'a@b.com', password: 'pass' })
    ).rejects.toThrow('Login failed')
  })

  test('register_OkButNullData_ThrowsError', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: null,
      error: null,
      ok: true,
    })

    await expect(
      repo.register({
        email: 'a@b.com',
        password: 'pass',
        name: 'N',
        role: 'coach',
      })
    ).rejects.toThrow('Registration failed')
  })

  test('refresh_OkButNullData_ThrowsError', async () => {
    mockedApiClient.post.mockResolvedValue({
      data: null,
      error: null,
      ok: true,
    })

    await expect(repo.refresh('token')).rejects.toThrow('Token refresh failed')
  })
})
