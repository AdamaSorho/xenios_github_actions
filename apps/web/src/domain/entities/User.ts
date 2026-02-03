/**
 * User entity - core business object.
 * This is a pure domain type with no external dependencies.
 */
export interface User {
  id: string
  email: string
  name: string
  createdAt: Date
  updatedAt: Date
}

/**
 * Input for creating a new user.
 */
export interface CreateUserInput {
  email: string
  name: string
}
