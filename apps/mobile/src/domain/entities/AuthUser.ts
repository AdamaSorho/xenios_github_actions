export type UserRole = 'coach' | 'client' | 'admin'

export const VALID_ROLES: UserRole[] = ['coach', 'client', 'admin']

export interface AuthUser {
  id: string
  email: string
  name: string
  role: UserRole
  avatarUrl?: string
  createdAt: string
  updatedAt: string
}

export function isValidRole(role: string): role is UserRole {
  return VALID_ROLES.includes(role as UserRole)
}
