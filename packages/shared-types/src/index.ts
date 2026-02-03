/**
 * Shared types for the Xenios platform.
 * Used across Backend, Web, and Mobile apps.
 */

// User types
export interface User {
  id: string
  email: string
  name: string
  createdAt: string // ISO string for JSON serialization
  updatedAt: string
}

export interface CreateUserInput {
  email: string
  name: string
}

export interface UpdateUserInput {
  email?: string
  name?: string
}

// API response types
export interface ApiResponse<T> {
  data: T | null
  error: string | null
  ok: boolean
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
  hasMore: boolean
}

// Error types
export interface ApiError {
  code: string
  message: string
  details?: Record<string, unknown>
}
