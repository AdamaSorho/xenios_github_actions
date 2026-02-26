'use client'

import { useState, useCallback } from 'react'
import { InsightCard } from '@/domain/entities/InsightCard'

interface InsightCardItemProps {
  insight: InsightCard
  onApprove: (id: string) => Promise<void>
  onDismiss: (id: string) => Promise<void>
  onEdit: (id: string, title: string, body: string) => Promise<void>
  onShare?: (id: string) => Promise<void>
}

const priorityColors: Record<string, string> = {
  urgent: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-green-100 text-green-800',
}

const categoryColors: Record<string, string> = {
  general: 'bg-gray-100 text-gray-800',
  nutrition: 'bg-emerald-100 text-emerald-800',
  recovery: 'bg-blue-100 text-blue-800',
  performance: 'bg-purple-100 text-purple-800',
  behavior: 'bg-pink-100 text-pink-800',
  safety: 'bg-red-100 text-red-800',
}

export function InsightCardItem({ insight, onApprove, onDismiss, onEdit, onShare }: InsightCardItemProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editTitle, setEditTitle] = useState(insight.title)
  const [editBody, setEditBody] = useState(insight.body)
  const [isLoading, setIsLoading] = useState(false)

  const withLoading = useCallback(
    (fn: () => Promise<void>) => async () => {
      setIsLoading(true)
      try {
        await fn()
      } finally {
        setIsLoading(false)
      }
    },
    []
  )

  const handleApprove = withLoading(() => onApprove(insight.id))
  const handleDismiss = withLoading(() => onDismiss(insight.id))
  const handleShare = withLoading(async () => {
    if (onShare) await onShare(insight.id)
  })
  const handleSave = withLoading(async () => {
    await onEdit(insight.id, editTitle, editBody)
    setIsEditing(false)
  })

  const handleCancel = () => {
    setEditTitle(insight.title)
    setEditBody(insight.body)
    setIsEditing(false)
  }

  return (
    <div
      className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm"
      data-testid={`insight-card-${insight.id}`}
    >
      <div className="mb-3 flex items-start justify-between">
        <div className="flex flex-wrap gap-2">
          <span
            className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${priorityColors[insight.priority] || ''}`}
            data-testid="insight-priority"
          >
            {insight.priority}
          </span>
          <span
            className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${categoryColors[insight.category] || ''}`}
            data-testid="insight-category"
          >
            {insight.category}
          </span>
          <span
            className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600"
            data-testid="insight-status"
          >
            {insight.status}
          </span>
        </div>
        {insight.clientName && (
          <span className="text-sm text-gray-500" data-testid="insight-client-name">
            {insight.clientName}
          </span>
        )}
      </div>

      {isEditing ? (
        <div data-testid="insight-edit-form">
          <input
            type="text"
            value={editTitle}
            onChange={(e) => setEditTitle(e.target.value)}
            className="mb-2 w-full rounded border border-gray-300 px-3 py-2 text-sm font-semibold"
            data-testid="insight-edit-title"
          />
          <textarea
            value={editBody}
            onChange={(e) => setEditBody(e.target.value)}
            className="mb-3 w-full rounded border border-gray-300 px-3 py-2 text-sm"
            rows={4}
            data-testid="insight-edit-body"
          />
          <div className="flex gap-2">
            <button
              onClick={handleSave}
              disabled={isLoading}
              className="rounded bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              data-testid="insight-save-btn"
            >
              Save
            </button>
            <button
              onClick={handleCancel}
              className="rounded border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50"
              data-testid="insight-cancel-btn"
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <>
          <h3 className="mb-1 text-base font-semibold text-gray-900" data-testid="insight-title">
            {insight.title}
          </h3>
          <p className="mb-3 text-sm text-gray-600" data-testid="insight-body">
            {insight.body}
          </p>

          {insight.evidence && insight.evidence.length > 0 && (
            <div className="mb-3 rounded bg-gray-50 p-2" data-testid="insight-evidence">
              <p className="mb-1 text-xs font-medium text-gray-500">Evidence</p>
              {insight.evidence.map((e, idx) => (
                <p key={idx} className="text-xs text-gray-600">
                  {e.description}
                </p>
              ))}
            </div>
          )}

          <div className="flex items-center justify-between">
            <span className="text-xs text-gray-400">
              {new Date(insight.createdAt).toLocaleDateString()}
            </span>
            <div className="flex gap-2">
              {insight.status === 'draft' && (
                <>
                  <button
                    onClick={() => setIsEditing(true)}
                    className="rounded border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50"
                    data-testid="insight-edit-btn"
                  >
                    Edit
                  </button>
                  <button
                    onClick={handleDismiss}
                    disabled={isLoading}
                    className="rounded border border-red-300 px-3 py-1.5 text-sm font-medium text-red-700 hover:bg-red-50 disabled:opacity-50"
                    data-testid="insight-dismiss-btn"
                  >
                    Dismiss
                  </button>
                  <button
                    onClick={handleApprove}
                    disabled={isLoading}
                    className="rounded bg-green-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-green-700 disabled:opacity-50"
                    data-testid="insight-approve-btn"
                  >
                    Approve
                  </button>
                </>
              )}
              {insight.status === 'approved' && onShare && (
                <button
                  onClick={handleShare}
                  disabled={isLoading}
                  className="rounded bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                  data-testid="insight-share-btn"
                >
                  Share with Client
                </button>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
