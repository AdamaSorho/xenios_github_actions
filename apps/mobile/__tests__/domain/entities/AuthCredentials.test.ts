import {
  validateEmail,
  validatePassword,
  validateLoginCredentials,
  validateRegisterCredentials,
} from '@/domain/entities/AuthCredentials'

describe('AuthCredentials', () => {
  describe('validateEmail', () => {
    it('should accept valid email', () => {
      expect(validateEmail('user@example.com')).toBe(true)
    })

    it('should accept email with subdomain', () => {
      expect(validateEmail('user@mail.example.com')).toBe(true)
    })

    it('should reject email without @', () => {
      expect(validateEmail('userexample.com')).toBe(false)
    })

    it('should reject email without domain', () => {
      expect(validateEmail('user@')).toBe(false)
    })

    it('should reject email with spaces', () => {
      expect(validateEmail('user @example.com')).toBe(false)
    })

    it('should reject empty string', () => {
      expect(validateEmail('')).toBe(false)
    })
  })

  describe('validatePassword', () => {
    it('should accept password with 8 characters', () => {
      expect(validatePassword('12345678')).toBe(true)
    })

    it('should accept password with more than 8 characters', () => {
      expect(validatePassword('longpassword123')).toBe(true)
    })

    it('should reject password with less than 8 characters', () => {
      expect(validatePassword('short')).toBe(false)
    })

    it('should reject empty password', () => {
      expect(validatePassword('')).toBe(false)
    })
  })

  describe('validateLoginCredentials', () => {
    it('should pass with valid credentials', () => {
      const result = validateLoginCredentials({
        email: 'user@example.com',
        password: 'password123',
      })
      expect(result.valid).toBe(true)
      expect(result.errors).toHaveLength(0)
    })

    it('should fail with empty email', () => {
      const result = validateLoginCredentials({
        email: '',
        password: 'password123',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Email is required')
    })

    it('should fail with invalid email format', () => {
      const result = validateLoginCredentials({
        email: 'invalid-email',
        password: 'password123',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Email format is invalid')
    })

    it('should fail with empty password', () => {
      const result = validateLoginCredentials({
        email: 'user@example.com',
        password: '',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Password is required')
    })

    it('should fail with whitespace-only email', () => {
      const result = validateLoginCredentials({
        email: '   ',
        password: 'password123',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Email is required')
    })

    it('should return multiple errors when both fields invalid', () => {
      const result = validateLoginCredentials({
        email: '',
        password: '',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toHaveLength(2)
    })
  })

  describe('validateRegisterCredentials', () => {
    const validCredentials = {
      email: 'user@example.com',
      password: 'password123',
      name: 'John Doe',
      role: 'coach' as const,
    }

    it('should pass with valid credentials', () => {
      const result = validateRegisterCredentials(validCredentials)
      expect(result.valid).toBe(true)
      expect(result.errors).toHaveLength(0)
    })

    it('should fail with empty email', () => {
      const result = validateRegisterCredentials({
        ...validCredentials,
        email: '',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Email is required')
    })

    it('should fail with invalid email format', () => {
      const result = validateRegisterCredentials({
        ...validCredentials,
        email: 'bad-email',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Email format is invalid')
    })

    it('should fail with empty password', () => {
      const result = validateRegisterCredentials({
        ...validCredentials,
        password: '',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Password is required')
    })

    it('should fail with short password', () => {
      const result = validateRegisterCredentials({
        ...validCredentials,
        password: 'short',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Password must be at least 8 characters')
    })

    it('should fail with empty name', () => {
      const result = validateRegisterCredentials({
        ...validCredentials,
        name: '',
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Name is required')
    })

    it('should fail with missing role', () => {
      const result = validateRegisterCredentials({
        ...validCredentials,
        role: '' as any,
      })
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('Role is required')
    })

    it('should return all errors when multiple fields invalid', () => {
      const result = validateRegisterCredentials({
        email: '',
        password: '',
        name: '',
        role: '' as any,
      })
      expect(result.valid).toBe(false)
      expect(result.errors.length).toBeGreaterThanOrEqual(4)
    })
  })
})
