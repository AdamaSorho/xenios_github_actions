'use client'

import { useState, useEffect, useCallback } from 'react'
import { InsightCard } from '@/domain/entities/InsightCard'
import { InsightCardItem } from '@/presentation/components/InsightCardItem'
import {
  getInsightQueueUseCase,
  approveInsightUseCase,
  dismissInsightUseCase,
  editInsightUseCase,
  shareInsightUseCase,
} from '@/infrastructure/container'

export default function InsightsPage() {
  const [insights, setInsights] = useState<InsightCard[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState('draft')

  const fetchInsights = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await getInsightQueueUseCase.execute(statusFilter || undefined)
      setInsights(result.insights)
      setTotal(result.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load insights')
    } finally {
      setLoading(false)
    }
  }, [statusFilter])

  useEffect(() => {
    fetchInsights()
  }, [fetchInsights])

  const handleApprove = async (id: string) => {
    try {
      await approveInsightUseCase.execute(id)
      await fetchInsights()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to approve insight')
    }
  }

  const handleDismiss = async (id: string) => {
    try {
      await dismissInsightUseCase.execute(id)
      await fetchInsights()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to dismiss insight')
    }
  }

  const handleEdit = async (id: string, title: string, body: string) => {
    try {
      await editInsightUseCase.execute(id, { title, body })
      await fetchInsights()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to edit insight')
    }
  }

  const handleShare = async (id: string) => {
    try {
      await shareInsightUseCase.execute(id)
      await fetchInsights()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to share insight')
    }
  }

  return (
    <div data-testid="insights-page">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Insights Queue</h1>
          {total > 0 && (
            <p className="mt-1 text-sm text-gray-500">
              {total} insight{total !== 1 ? 's' : ''} {statusFilter ? `(${statusFilter})` : ''}
            </p>
          )}
        </div>
        <div className="flex gap-2">
          {['draft', 'approved', 'shared', 'dismissed', ''].map((status) => (
            <button
              key={status}
              onClick={() => setStatusFilter(status)}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                statusFilter === status
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
              data-testid={`filter-${status || 'all'}`}
            >
              {status ? status.charAt(0).toUpperCase() + status.slice(1) : 'All'}
            </button>
          ))}
        </div>
      </div>

      {error && (
        <div
          className="mb-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700"
          data-testid="error-message"
        >
          {error}
        </div>
      )}

      {loading ? (
        <div className="space-y-4" data-testid="loading-skeleton">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-32 animate-pulse rounded-lg bg-gray-100" />
          ))}
        </div>
      ) : insights.length === 0 ? (
        <div
          className="rounded-lg border border-gray-200 bg-white p-8 text-center"
          data-testid="empty-state"
        >
          <p className="text-sm text-gray-500">
            {statusFilter === 'draft'
              ? 'All caught up! No draft insights to review.'
              : 'No insights found with the selected filter.'}
          </p>
        </div>
      ) : (
        <div className="space-y-4" data-testid="insights-list">
          {insights.map((insight) => (
            <InsightCardItem
              key={insight.id}
              insight={insight}
              onApprove={handleApprove}
              onDismiss={handleDismiss}
              onEdit={handleEdit}
              onShare={handleShare}
            />
          ))}
        </div>
      )}
    </div>
  )
}
