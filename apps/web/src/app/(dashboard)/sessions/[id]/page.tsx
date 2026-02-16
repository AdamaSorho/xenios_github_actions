'use client'

import { use } from 'react'
import Link from 'next/link'

export default function SessionDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)

  return (
    <div data-testid="session-detail-page">
      <div className="mb-6">
        <Link
          href="/sessions"
          className="text-sm text-blue-600 hover:text-blue-800"
        >
          &larr; Back to Sessions
        </Link>
      </div>

      <h1 className="mb-6 text-2xl font-bold text-gray-900">
        Session Details
      </h1>

      <div className="rounded-lg border border-gray-200 bg-white p-6">
        <p className="text-sm text-gray-500">
          Session ID: {id}
        </p>
        <p className="mt-2 text-sm text-gray-500">
          Session transcript, summary, and exercises will be displayed here.
        </p>
      </div>
    </div>
  )
}
