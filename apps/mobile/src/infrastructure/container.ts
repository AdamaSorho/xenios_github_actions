/**
 * Dependency Injection Container
 *
 * Wires up the application by injecting infrastructure implementations
 * into use cases. This is the composition root.
 */
import { ApiUserRepository } from './repositories/ApiUserRepository'
import { GetUserUseCase } from '@/application/usecases/GetUserUseCase'
import { CreateUserUseCase } from '@/application/usecases/CreateUserUseCase'
import { SecureTokenStorage } from './storage/SecureTokenStorage'
import { ApiAuthRepository } from './api/ApiAuthRepository'
import { AuthenticatedApiClient } from './api/AuthenticatedApiClient'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'
import { RefreshTokenUseCase } from '@/application/usecases/RefreshTokenUseCase'
import { GetAuthStateUseCase } from '@/application/usecases/GetAuthStateUseCase'
import { AuthProviderDeps } from '@/presentation/context/AuthContext'

// Infrastructure: Repositories
const userRepository = new ApiUserRepository()
const tokenStorage = new SecureTokenStorage()
const authRepository = new ApiAuthRepository()
const authenticatedApiClient = new AuthenticatedApiClient(
  tokenStorage,
  authRepository
)

// Application: User use cases
export const getUserUseCase = new GetUserUseCase(userRepository)
export const createUserUseCase = new CreateUserUseCase(userRepository)

// Application: Auth use cases
export const loginUseCase = new LoginUseCase(authRepository, tokenStorage)
export const registerUseCase = new RegisterUseCase(authRepository, tokenStorage)
export const logoutUseCase = new LogoutUseCase(authRepository, tokenStorage)
export const refreshTokenUseCase = new RefreshTokenUseCase(
  authRepository,
  tokenStorage
)
export const getAuthStateUseCase = new GetAuthStateUseCase(tokenStorage)

// Auth provider dependencies bundle
export const authProviderDeps: AuthProviderDeps = {
  loginUseCase,
  registerUseCase,
  logoutUseCase,
  getAuthStateUseCase,
  authenticatedApiClient,
}
