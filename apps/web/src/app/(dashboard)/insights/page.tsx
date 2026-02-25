'use client'

import { useState, useEffect, useCallback } from 'react'
import { InsightCard, InsightStatus } from '@/domain/entities/InsightCard'
import { InsightCardItem } from '@/presentation/components/InsightCardItem'
import {
  getInsightQueueUseCase,
  approveInsightUseCase,
  dismissInsightUseCase,
  editInsightUseCase,
  shareInsightUseCase,
} from '@/infrastructure/container'

type StatusFilter = InsightStatus | 'all'

export default function InsightsPage() {
  const [insights, setInsights] = useState<InsightCard[]>([])
  const [total, setTotal] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('draft')

  const fetchInsights = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const status = statusFilter === 'all' ? undefined : statusFilter
      const result = await getInsightQueueUseCase.execute(status)
      setInsights(result.insights)
      setTotal(result.pagination.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load insights')
    } finally {
      setIsLoading(false)
    }
  }, [statusFilter])

  useEffect(() => {
    fetchInsights()
  }, [fetchInsights])

  const handleApprove = async (id: string) => {
    await approveInsightUseCase.execute(id)
    await fetchInsights()
  }

  const handleDismiss = async (id: string) => {
    await dismissInsightUseCase.execute(id)
    await fetchInsights()
  }

  const handleEdit = async (id: string, title: string, body: string) => {
    await editInsightUseCase.execute(id, { title, body })
    await fetchInsights()
  }

  const handleShare = async (id: string) => {
    await shareInsightUseCase.execute(id)
    await fetchInsights()
  }

  const statusTabs: { label: string; value: StatusFilter }[] = [
    { label: 'Draft', value: 'draft' },
    { label: 'Approved', value: 'approved' },
    { label: 'Shared', value: 'shared' },
    { label: 'Dismissed', value: 'dismissed' },
    { label: 'All', value: 'all' },
  ]

  return (
    <div data-testid="insights-page">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Insight Queue</h1>
          <p className="mt-1 text-sm text-gray-500">
            Review AI-generated insights before sharing with clients
          </p>
        </div>
        {total > 0 && statusFilter === 'draft' && (
          <span
            className="inline-flex items-center rounded-full bg-blue-100 px-3 py-1 text-sm font-medium text-blue-800"
            data-testid="queue-count-badge"
          >
            {total} pending
          </span>
        )}
      </div>

      <div className="mb-6 flex gap-2" data-testid="status-tabs">
        {statusTabs.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setStatusFilter(tab.value)}
            className={`rounded-md px-4 py-2 text-sm font-medium transition-colors ${
              statusFilter === tab.value
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
            data-testid={`tab-${tab.value}`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {error && (
        <div className="mb-4 rounded-md bg-red-50 p-4 text-sm text-red-700" role="alert" data-testid="insights-error">
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="space-y-4" data-testid="insights-loading">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-32 animate-pulse rounded-lg bg-gray-200" />
          ))}
        </div>
      ) : insights.length === 0 ? (
        <div
          className="rounded-lg border border-gray-200 bg-white py-12 text-center"
          data-testid="insights-empty"
        >
          <p className="text-lg font-medium text-gray-900">All caught up!</p>
          <p className="mt-1 text-sm text-gray-500">
            No {statusFilter === 'all' ? '' : statusFilter + ' '}insights to review.
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
