/**
 * Dependency Injection Container
 *
 * Wires up the application by injecting infrastructure implementations
 * into use cases. This is the composition root.
 */
import { ApiUserRepository } from './repositories/ApiUserRepository'
import { GetUserUseCase } from '@/application/usecases/GetUserUseCase'
import { CreateUserUseCase } from '@/application/usecases/CreateUserUseCase'

// Infrastructure: API client repositories
const userRepository = new ApiUserRepository()

// Application: Use cases with injected repositories
export const getUserUseCase = new GetUserUseCase(userRepository)
export const createUserUseCase = new CreateUserUseCase(userRepository)
