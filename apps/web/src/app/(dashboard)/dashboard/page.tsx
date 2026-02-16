'use client'

import { useAuth } from '@/presentation/hooks/useAuth'

export default function DashboardPage() {
  const { user } = useAuth()

  return (
    <div data-testid="dashboard-page">
      <h1 className="mb-6 text-2xl font-bold text-gray-900">Dashboard</h1>

      <p className="mb-8 text-gray-600">
        Welcome back{user ? `, ${user.name}` : ''}
      </p>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <p className="text-sm font-medium text-gray-500">Upcoming Sessions</p>
          <p className="mt-2 text-3xl font-bold text-gray-900">--</p>
          <p className="mt-1 text-sm text-gray-500">This week</p>
        </div>
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <p className="text-sm font-medium text-gray-500">Active Clients</p>
          <p className="mt-2 text-3xl font-bold text-gray-900">--</p>
          <p className="mt-1 text-sm text-gray-500">Total</p>
        </div>
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <p className="text-sm font-medium text-gray-500">Completed Sessions</p>
          <p className="mt-2 text-3xl font-bold text-gray-900">--</p>
          <p className="mt-1 text-sm text-gray-500">This month</p>
        </div>
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <p className="text-sm font-medium text-gray-500">Alerts</p>
          <p className="mt-2 text-3xl font-bold text-gray-900">--</p>
          <p className="mt-1 text-sm text-gray-500">Pending review</p>
        </div>
      </div>

      <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Recent Activity
          </h2>
          <p className="text-sm text-gray-500">
            No recent activity to display.
          </p>
        </div>
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Quick Actions
          </h2>
          <div className="space-y-2">
            <button className="w-full rounded-md border border-gray-300 px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50">
              Schedule a session
            </button>
            <button className="w-full rounded-md border border-gray-300 px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50">
              Add a new client
            </button>
            <button className="w-full rounded-md border border-gray-300 px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50">
              View analytics report
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
