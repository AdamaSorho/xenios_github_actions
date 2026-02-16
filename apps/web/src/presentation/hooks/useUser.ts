'use client'

import { useEffect, useReducer } from 'react'
import { User } from '@/domain/entities/User'
import { getUserUseCase } from '@/infrastructure/container'

interface UseUserResult {
  user: User | null
  loading: boolean
  error: Error | null
}

type Action =
  | { type: 'FETCH_START' }
  | { type: 'FETCH_SUCCESS'; user: User | null }
  | { type: 'FETCH_ERROR'; error: Error }
  | { type: 'RESET' }

function reducer(_state: UseUserResult, action: Action): UseUserResult {
  switch (action.type) {
    case 'FETCH_START':
      return { user: null, loading: true, error: null }
    case 'FETCH_SUCCESS':
      return { user: action.user, loading: false, error: null }
    case 'FETCH_ERROR':
      return { user: null, loading: false, error: action.error }
    case 'RESET':
      return { user: null, loading: false, error: null }
  }
}

const initialState: UseUserResult = { user: null, loading: false, error: null }

/**
 * Hook for fetching a user by ID.
 * When id is null, returns null user without fetching.
 */
export function useUser(id: string | null): UseUserResult {
  const [state, dispatch] = useReducer(reducer, initialState)

  useEffect(() => {
    if (!id) {
      dispatch({ type: 'RESET' })
      return
    }

    let cancelled = false
    dispatch({ type: 'FETCH_START' })

    getUserUseCase
      .execute(id)
      .then((result) => {
        if (!cancelled) {
          dispatch({ type: 'FETCH_SUCCESS', user: result })
        }
      })
      .catch((err) => {
        if (!cancelled) {
          dispatch({
            type: 'FETCH_ERROR',
            error: err instanceof Error ? err : new Error(String(err)),
          })
        }
      })

    return () => {
      cancelled = true
    }
  }, [id])

  return state
}
