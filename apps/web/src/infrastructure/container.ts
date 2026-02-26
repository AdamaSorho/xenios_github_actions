/**
 * Dependency Injection Container
 *
 * Wires up the application by injecting infrastructure implementations
 * into use cases. This is the composition root.
 */
import { ApiUserRepository } from './repositories/ApiUserRepository'
import { ApiAuthRepository } from './repositories/ApiAuthRepository'
import { LocalTokenStorage } from './auth/LocalTokenStorage'
import { ApiAuthTokenManager } from './auth/ApiAuthTokenManager'
import { AuthInterceptor } from './auth/AuthInterceptor'
import { GetUserUseCase } from '@/application/usecases/GetUserUseCase'
import { CreateUserUseCase } from '@/application/usecases/CreateUserUseCase'
import { LoginUseCase } from '@/application/usecases/LoginUseCase'
import { RegisterUseCase } from '@/application/usecases/RegisterUseCase'
import { LogoutUseCase } from '@/application/usecases/LogoutUseCase'
import { RefreshTokenUseCase } from '@/application/usecases/RefreshTokenUseCase'
import { ApiInsightRepository } from './repositories/ApiInsightRepository'
import { GetInsightQueueUseCase } from '@/application/usecases/GetInsightQueueUseCase'
import { ApproveInsightUseCase } from '@/application/usecases/ApproveInsightUseCase'
import { DismissInsightUseCase } from '@/application/usecases/DismissInsightUseCase'
import { EditInsightUseCase } from '@/application/usecases/EditInsightUseCase'
import { ShareInsightUseCase } from '@/application/usecases/ShareInsightUseCase'

// Infrastructure: API client repositories and services
const userRepository = new ApiUserRepository()
const authRepository = new ApiAuthRepository()
export const tokenStorage = new LocalTokenStorage()
export const tokenManager = new ApiAuthTokenManager()

// Automatic 401 token refresh interceptor
const authInterceptor = new AuthInterceptor(tokenStorage, authRepository, tokenManager)
authInterceptor.install()

// Application: Use cases with injected repositories
export const getUserUseCase = new GetUserUseCase(userRepository)
export const createUserUseCase = new CreateUserUseCase(userRepository)
export const loginUseCase = new LoginUseCase(authRepository)
export const registerUseCase = new RegisterUseCase(authRepository)
export const logoutUseCase = new LogoutUseCase(authRepository)
export const refreshTokenUseCase = new RefreshTokenUseCase(authRepository)

// Insight queue use cases
const insightRepository = new ApiInsightRepository()
export const getInsightQueueUseCase = new GetInsightQueueUseCase(insightRepository)
export const approveInsightUseCase = new ApproveInsightUseCase(insightRepository)
export const dismissInsightUseCase = new DismissInsightUseCase(insightRepository)
export const editInsightUseCase = new EditInsightUseCase(insightRepository)
export const shareInsightUseCase = new ShareInsightUseCase(insightRepository)
