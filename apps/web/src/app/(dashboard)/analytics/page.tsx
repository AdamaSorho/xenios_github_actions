'use client'

export default function AnalyticsPage() {
  return (
    <div data-testid="analytics-page">
      <h1 className="mb-6 text-2xl font-bold text-gray-900">Analytics</h1>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Session Trends
          </h2>
          <div className="flex h-48 items-center justify-center text-sm text-gray-500">
            Chart placeholder - session trends over time
          </div>
        </div>

        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Client Progress
          </h2>
          <div className="flex h-48 items-center justify-center text-sm text-gray-500">
            Chart placeholder - client progress metrics
          </div>
        </div>

        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Coaching Insights
          </h2>
          <div className="flex h-48 items-center justify-center text-sm text-gray-500">
            Insights placeholder - AI-generated coaching insights
          </div>
        </div>

        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Performance Metrics
          </h2>
          <div className="flex h-48 items-center justify-center text-sm text-gray-500">
            Metrics placeholder - key performance indicators
          </div>
        </div>
      </div>
    </div>
  )
}
