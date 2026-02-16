'use client'

import { useState, FormEvent } from 'react'
import { useAuth } from '@/presentation/hooks/useAuth'
import Link from 'next/link'
import { useRouter } from 'next/navigation'

export function LoginForm() {
  const { login, isLoading, error, clearError } = useAuth()
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    clearError()
    try {
      await login({ email, password })
      router.push('/dashboard')
    } catch {
      // Error is handled by the auth context
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4" data-testid="login-form">
      {error && (
        <div
          className="rounded-md bg-red-50 p-3 text-sm text-red-700"
          role="alert"
          data-testid="login-error"
        >
          {error}
        </div>
      )}

      <div>
        <label
          htmlFor="email"
          className="mb-1 block text-sm font-medium text-gray-700"
        >
          Email
        </label>
        <input
          id="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          placeholder="coach@example.com"
          data-testid="login-email"
        />
      </div>

      <div>
        <label
          htmlFor="password"
          className="mb-1 block text-sm font-medium text-gray-700"
        >
          Password
        </label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          placeholder="Enter your password"
          data-testid="login-password"
        />
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="w-full rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
        data-testid="login-submit"
      >
        {isLoading ? 'Signing in...' : 'Sign in'}
      </button>

      <div className="flex items-center justify-between text-sm">
        <Link
          href="/forgot-password"
          className="text-blue-600 hover:text-blue-800"
        >
          Forgot password?
        </Link>
        <Link href="/register" className="text-blue-600 hover:text-blue-800">
          Create account
        </Link>
      </div>
    </form>
  )
}
