/**
 * Dependency Injection Container
 *
 * Wires up the application by injecting infrastructure implementations
 * into use cases. This is the composition root.
 */
import { ApiUserRepository } from './repositories/ApiUserRepository'
import { ApiAuthRepository } from './repositories/ApiAuthRepository'
import { LocalTokenStorage } from './auth/LocalTokenStorage'
import { GetUserUseCase } from '@/application/usecases/GetUserUseCase'
import { CreateUserUseCase } from '@/application/usecases/CreateUserUseCase'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'
import { RefreshTokenUseCase } from '@/application/usecases/RefreshTokenUseCase'

// Infrastructure: API client repositories
const userRepository = new ApiUserRepository()
const authRepository = new ApiAuthRepository()
export const tokenStorage = new LocalTokenStorage()

// Application: Use cases with injected repositories
export const getUserUseCase = new GetUserUseCase(userRepository)
export const createUserUseCase = new CreateUserUseCase(userRepository)
export const loginUseCase = new LoginUseCase(authRepository)
export const registerUseCase = new RegisterUseCase(authRepository)
export const logoutUseCase = new LogoutUseCase(authRepository)
export const refreshTokenUseCase = new RefreshTokenUseCase(authRepository)
