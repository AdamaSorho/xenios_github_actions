import { isValidRole, VALID_ROLES } from '@/domain/entities/AuthUser'

describe('AuthUser', () => {
  describe('isValidRole', () => {
    it('should return true for coach role', () => {
      expect(isValidRole('coach')).toBe(true)
    })

    it('should return true for client role', () => {
      expect(isValidRole('client')).toBe(true)
    })

    it('should return true for admin role', () => {
      expect(isValidRole('admin')).toBe(true)
    })

    it('should return false for invalid role', () => {
      expect(isValidRole('superuser')).toBe(false)
    })

    it('should return false for empty string', () => {
      expect(isValidRole('')).toBe(false)
    })
  })

  describe('VALID_ROLES', () => {
    it('should contain exactly three roles', () => {
      expect(VALID_ROLES).toHaveLength(3)
    })

    it('should contain coach, client, and admin', () => {
      expect(VALID_ROLES).toEqual(['coach', 'client', 'admin'])
    })
  })
})
