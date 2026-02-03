'use client'

import { useState, useEffect } from 'react'
import { User } from '@/domain/entities/User'
import { getUserUseCase } from '@/infrastructure/container'

interface UseUserResult {
  user: User | null
  loading: boolean
  error: Error | null
}

/**
 * Hook for fetching a user by ID.
 */
export function useUser(id: string | null): UseUserResult {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    if (!id) {
      setUser(null)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    getUserUseCase
      .execute(id)
      .then((result) => {
        if (!cancelled) {
          setUser(result)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)))
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [id])

  return { user, loading, error }
}
