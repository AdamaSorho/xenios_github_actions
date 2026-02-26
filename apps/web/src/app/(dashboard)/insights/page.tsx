'use client'

import { useCallback, useEffect, useReducer, useState } from 'react'
import type { InsightCard } from '@/domain/entities/InsightCard'
import {
  getInsightQueueUseCase,
  approveInsightUseCase,
  dismissInsightUseCase,
  editInsightUseCase,
  shareInsightUseCase,
} from '@/infrastructure/container'

interface State {
  insights: InsightCard[]
  total: number
  loading: boolean
  error: string | null
}

type Action =
  | { type: 'FETCH_START' }
  | { type: 'FETCH_SUCCESS'; insights: InsightCard[]; total: number }
  | { type: 'FETCH_ERROR'; error: string }
  | { type: 'UPDATE_INSIGHT'; insight: InsightCard }
  | { type: 'REMOVE_INSIGHT'; insightId: string }

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'FETCH_START':
      return { ...state, loading: true, error: null }
    case 'FETCH_SUCCESS':
      return { ...state, loading: false, insights: action.insights, total: action.total }
    case 'FETCH_ERROR':
      return { ...state, loading: false, error: action.error }
    case 'UPDATE_INSIGHT':
      return {
        ...state,
        insights: state.insights.map((i) =>
          i.id === action.insight.id ? action.insight : i
        ),
      }
    case 'REMOVE_INSIGHT':
      return {
        ...state,
        insights: state.insights.filter((i) => i.id !== action.insightId),
        total: state.total - 1,
      }
    default:
      return state
  }
}

export default function InsightsQueuePage() {
  const [state, dispatch] = useReducer(reducer, {
    insights: [],
    total: 0,
    loading: true,
    error: null,
  })
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const [editBody, setEditBody] = useState('')
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState<string>('draft')

  const fetchQueue = useCallback(async () => {
    dispatch({ type: 'FETCH_START' })
    try {
      const result = await getInsightQueueUseCase.execute(statusFilter || undefined)
      dispatch({ type: 'FETCH_SUCCESS', insights: result.insights, total: result.pagination.total })
    } catch (err) {
      dispatch({ type: 'FETCH_ERROR', error: err instanceof Error ? err.message : 'Failed to load insights' })
    }
  }, [statusFilter])

  useEffect(() => {
    fetchQueue()
  }, [fetchQueue])

  const handleApprove = async (insightId: string) => {
    setActionLoading(insightId)
    try {
      await approveInsightUseCase.execute(insightId)
      dispatch({ type: 'REMOVE_INSIGHT', insightId })
    } catch {
      // keep card in list on error
    } finally {
      setActionLoading(null)
    }
  }

  const handleDismiss = async (insightId: string) => {
    setActionLoading(insightId)
    try {
      await dismissInsightUseCase.execute(insightId)
      dispatch({ type: 'REMOVE_INSIGHT', insightId })
    } catch {
      // keep card in list on error
    } finally {
      setActionLoading(null)
    }
  }

  const handleShare = async (insightId: string) => {
    setActionLoading(insightId)
    try {
      const updated = await shareInsightUseCase.execute(insightId)
      dispatch({ type: 'UPDATE_INSIGHT', insight: updated })
    } catch {
      // keep card in list on error
    } finally {
      setActionLoading(null)
    }
  }

  const startEdit = (insight: InsightCard) => {
    setEditingId(insight.id)
    setEditTitle(insight.title)
    setEditBody(insight.body)
  }

  const cancelEdit = () => {
    setEditingId(null)
    setEditTitle('')
    setEditBody('')
  }

  const handleSaveEdit = async (insightId: string) => {
    setActionLoading(insightId)
    try {
      const updated = await editInsightUseCase.execute(insightId, {
        title: editTitle,
        body: editBody,
      })
      dispatch({ type: 'UPDATE_INSIGHT', insight: updated })
      cancelEdit()
    } catch {
      // keep editing on error
    } finally {
      setActionLoading(null)
    }
  }

  const priorityColor = (priority: string) => {
    switch (priority) {
      case 'urgent': return 'bg-red-100 text-red-800'
      case 'high': return 'bg-orange-100 text-orange-800'
      case 'medium': return 'bg-yellow-100 text-yellow-800'
      case 'low': return 'bg-green-100 text-green-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const categoryColor = (category: string) => {
    switch (category) {
      case 'nutrition': return 'bg-emerald-100 text-emerald-800'
      case 'exercise': return 'bg-blue-100 text-blue-800'
      case 'sleep': return 'bg-indigo-100 text-indigo-800'
      case 'stress': return 'bg-purple-100 text-purple-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const statusBadge = (status: string) => {
    switch (status) {
      case 'draft': return 'bg-gray-100 text-gray-700'
      case 'approved': return 'bg-green-100 text-green-700'
      case 'dismissed': return 'bg-red-100 text-red-700'
      case 'shared': return 'bg-blue-100 text-blue-700'
      default: return 'bg-gray-100 text-gray-700'
    }
  }

  return (
    <div data-testid="insights-queue-page">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Insight Queue</h1>
          <p className="mt-1 text-sm text-gray-500">
            Review and approve AI-generated insights before sharing with clients
          </p>
        </div>
        {state.total > 0 && statusFilter === 'draft' && (
          <span
            className="inline-flex items-center rounded-full bg-blue-100 px-3 py-1 text-sm font-medium text-blue-700"
            data-testid="queue-badge"
          >
            {state.total} pending
          </span>
        )}
      </div>

      <div className="mb-4 flex gap-2" data-testid="status-filter">
        {['draft', 'approved', 'shared', 'dismissed'].map((s) => (
          <button
            key={s}
            onClick={() => setStatusFilter(s)}
            className={`rounded-md px-3 py-1.5 text-sm font-medium capitalize transition-colors ${
              statusFilter === s
                ? 'bg-gray-900 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
            data-testid={`filter-${s}`}
          >
            {s}
          </button>
        ))}
      </div>

      {state.loading && (
        <div className="py-12 text-center text-gray-500" data-testid="loading-state">
          Loading insights...
        </div>
      )}

      {state.error && (
        <div className="rounded-md border border-red-200 bg-red-50 p-4 text-red-700" data-testid="error-state">
          {state.error}
        </div>
      )}

      {!state.loading && !state.error && state.insights.length === 0 && (
        <div className="py-12 text-center" data-testid="empty-state">
          <p className="text-lg font-medium text-gray-900">All caught up!</p>
          <p className="mt-1 text-sm text-gray-500">No {statusFilter} insights to review.</p>
        </div>
      )}

      {!state.loading && state.insights.length > 0 && (
        <div className="space-y-4" data-testid="insights-list">
          {state.insights.map((insight) => (
            <div
              key={insight.id}
              className="rounded-lg border border-gray-200 bg-white p-6"
              data-testid={`insight-card-${insight.id}`}
            >
              <div className="mb-3 flex items-start justify-between">
                <div className="flex items-center gap-2">
                  <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${priorityColor(insight.priority)}`}>
                    {insight.priority}
                  </span>
                  <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${categoryColor(insight.category)}`}>
                    {insight.category}
                  </span>
                  <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${statusBadge(insight.status)}`}>
                    {insight.status}
                  </span>
                </div>
                <span className="text-xs text-gray-400">
                  {insight.client_name}
                </span>
              </div>

              {editingId === insight.id ? (
                <div className="space-y-3" data-testid="edit-form">
                  <input
                    type="text"
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                    data-testid="edit-title-input"
                  />
                  <textarea
                    value={editBody}
                    onChange={(e) => setEditBody(e.target.value)}
                    rows={3}
                    className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
                    data-testid="edit-body-input"
                  />
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleSaveEdit(insight.id)}
                      disabled={actionLoading === insight.id}
                      className="rounded-md bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
                      data-testid="save-edit-button"
                    >
                      Save
                    </button>
                    <button
                      onClick={cancelEdit}
                      className="rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50"
                      data-testid="cancel-edit-button"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <>
                  <h3 className="text-lg font-semibold text-gray-900" data-testid="insight-title">
                    {insight.title}
                  </h3>
                  <p className="mt-1 text-sm text-gray-600" data-testid="insight-body">
                    {insight.body}
                  </p>
                </>
              )}

              {insight.evidence && insight.evidence.length > 0 && (
                <div className="mt-3 border-t border-gray-100 pt-3">
                  <p className="text-xs font-medium text-gray-500">Evidence</p>
                  <ul className="mt-1 space-y-1">
                    {insight.evidence.map((ev, idx) => (
                      <li key={idx} className="text-xs text-gray-600">
                        {ev.description}
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {editingId !== insight.id && (
                <div className="mt-4 flex gap-2 border-t border-gray-100 pt-4" data-testid="insight-actions">
                  {insight.status === 'draft' && (
                    <>
                      <button
                        onClick={() => handleApprove(insight.id)}
                        disabled={actionLoading === insight.id}
                        className="rounded-md bg-green-600 px-3 py-1.5 text-sm text-white hover:bg-green-700 disabled:opacity-50"
                        data-testid="approve-button"
                      >
                        Approve
                      </button>
                      <button
                        onClick={() => handleDismiss(insight.id)}
                        disabled={actionLoading === insight.id}
                        className="rounded-md border border-red-300 px-3 py-1.5 text-sm text-red-700 hover:bg-red-50 disabled:opacity-50"
                        data-testid="dismiss-button"
                      >
                        Dismiss
                      </button>
                      <button
                        onClick={() => startEdit(insight)}
                        disabled={actionLoading === insight.id}
                        className="rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                        data-testid="edit-button"
                      >
                        Edit
                      </button>
                    </>
                  )}
                  {insight.status === 'approved' && (
                    <button
                      onClick={() => handleShare(insight.id)}
                      disabled={actionLoading === insight.id}
                      className="rounded-md bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
                      data-testid="share-button"
                    >
                      Share with Client
                    </button>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
