'use client'

import { useState } from 'react'
import { InsightCard } from '@/domain/entities/InsightCard'

interface InsightCardItemProps {
  insight: InsightCard
  onApprove: (id: string) => void
  onDismiss: (id: string) => void
  onEdit: (id: string, title: string, body: string) => void
  onShare?: (id: string) => void
}

const priorityColors: Record<string, string> = {
  urgent: 'bg-red-100 text-red-800',
  high: 'bg-orange-100 text-orange-800',
  medium: 'bg-yellow-100 text-yellow-800',
  low: 'bg-green-100 text-green-800',
}

const categoryColors: Record<string, string> = {
  general: 'bg-gray-100 text-gray-700',
  nutrition: 'bg-emerald-100 text-emerald-700',
  recovery: 'bg-blue-100 text-blue-700',
  performance: 'bg-purple-100 text-purple-700',
  behavior: 'bg-indigo-100 text-indigo-700',
  safety: 'bg-red-100 text-red-700',
}

const statusLabels: Record<string, string> = {
  draft: 'Draft',
  approved: 'Approved',
  dismissed: 'Dismissed',
  shared: 'Shared',
}

export function InsightCardItem({ insight, onApprove, onDismiss, onEdit, onShare }: InsightCardItemProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editTitle, setEditTitle] = useState(insight.title)
  const [editBody, setEditBody] = useState(insight.body)

  const handleSaveEdit = () => {
    onEdit(insight.id, editTitle, editBody)
    setIsEditing(false)
  }

  const handleCancelEdit = () => {
    setEditTitle(insight.title)
    setEditBody(insight.body)
    setIsEditing(false)
  }

  return (
    <div
      className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm"
      data-testid={`insight-card-${insight.id}`}
    >
      <div className="mb-2 flex items-start justify-between">
        <div className="flex flex-wrap gap-2">
          <span
            className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${priorityColors[insight.priority] || 'bg-gray-100 text-gray-700'}`}
            data-testid="priority-badge"
          >
            {insight.priority}
          </span>
          <span
            className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${categoryColors[insight.category] || 'bg-gray-100 text-gray-700'}`}
            data-testid="category-badge"
          >
            {insight.category}
          </span>
          <span
            className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600"
            data-testid="status-badge"
          >
            {statusLabels[insight.status] || insight.status}
          </span>
        </div>
        {insight.clientName && (
          <span className="text-sm text-gray-500" data-testid="client-name">
            {insight.clientName}
          </span>
        )}
      </div>

      {isEditing ? (
        <div data-testid="edit-form">
          <input
            type="text"
            value={editTitle}
            onChange={(e) => setEditTitle(e.target.value)}
            className="mb-2 w-full rounded border border-gray-300 px-3 py-1.5 text-sm font-semibold focus:border-blue-500 focus:outline-none"
            data-testid="edit-title-input"
          />
          <textarea
            value={editBody}
            onChange={(e) => setEditBody(e.target.value)}
            rows={3}
            className="mb-2 w-full rounded border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none"
            data-testid="edit-body-input"
          />
          <div className="flex gap-2">
            <button
              onClick={handleSaveEdit}
              className="rounded bg-blue-600 px-3 py-1 text-xs font-medium text-white hover:bg-blue-700"
              data-testid="save-edit-button"
            >
              Save
            </button>
            <button
              onClick={handleCancelEdit}
              className="rounded border border-gray-300 px-3 py-1 text-xs font-medium text-gray-700 hover:bg-gray-50"
              data-testid="cancel-edit-button"
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <>
          <h3 className="mb-1 text-sm font-semibold text-gray-900" data-testid="insight-title">
            {insight.title}
          </h3>
          <p className="mb-3 text-sm text-gray-600" data-testid="insight-body">
            {insight.body}
          </p>
        </>
      )}

      {insight.evidence && insight.evidence.length > 0 && (
        <div className="mb-3 rounded bg-gray-50 p-2" data-testid="evidence-section">
          <p className="mb-1 text-xs font-medium text-gray-500">Evidence</p>
          {insight.evidence.map((ev, idx) => (
            <p key={idx} className="text-xs text-gray-600">
              {ev.description}
            </p>
          ))}
        </div>
      )}

      {!isEditing && (
        <div className="flex gap-2" data-testid="action-buttons">
          {insight.status === 'draft' && (
            <>
              <button
                onClick={() => onApprove(insight.id)}
                className="rounded bg-green-600 px-3 py-1 text-xs font-medium text-white hover:bg-green-700"
                data-testid="approve-button"
              >
                Approve
              </button>
              <button
                onClick={() => onDismiss(insight.id)}
                className="rounded border border-red-300 px-3 py-1 text-xs font-medium text-red-700 hover:bg-red-50"
                data-testid="dismiss-button"
              >
                Dismiss
              </button>
              <button
                onClick={() => setIsEditing(true)}
                className="rounded border border-gray-300 px-3 py-1 text-xs font-medium text-gray-700 hover:bg-gray-50"
                data-testid="edit-button"
              >
                Edit
              </button>
            </>
          )}
          {insight.status === 'approved' && onShare && (
            <button
              onClick={() => onShare(insight.id)}
              className="rounded bg-blue-600 px-3 py-1 text-xs font-medium text-white hover:bg-blue-700"
              data-testid="share-button"
            >
              Share with Client
            </button>
          )}
        </div>
      )}

      <div className="mt-2 text-xs text-gray-400">
        Created {new Date(insight.createdAt).toLocaleDateString()}
      </div>
    </div>
  )
}
