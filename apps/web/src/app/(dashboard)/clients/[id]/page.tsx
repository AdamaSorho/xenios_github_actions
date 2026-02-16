'use client'

import { use } from 'react'
import Link from 'next/link'

export default function ClientDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)

  return (
    <div data-testid="client-detail-page">
      <div className="mb-6">
        <Link
          href="/clients"
          className="text-sm text-blue-600 hover:text-blue-800"
        >
          &larr; Back to Clients
        </Link>
      </div>

      <h1 className="mb-6 text-2xl font-bold text-gray-900">
        Client Details
      </h1>

      <div className="rounded-lg border border-gray-200 bg-white p-6">
        <p className="text-sm text-gray-500">
          Client ID: {id}
        </p>
        <p className="mt-2 text-sm text-gray-500">
          Client profile details will be displayed here.
        </p>
      </div>
    </div>
  )
}
