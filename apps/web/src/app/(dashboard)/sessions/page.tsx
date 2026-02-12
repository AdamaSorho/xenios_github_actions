'use client'

export default function SessionsPage() {
  return (
    <div data-testid="sessions-page">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Sessions</h1>
        <button className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">
          New Session
        </button>
      </div>

      <div className="rounded-lg border border-gray-200 bg-white">
        <div className="p-8 text-center text-sm text-gray-500">
          No sessions found. Start your first coaching session.
        </div>
      </div>
    </div>
  )
}
