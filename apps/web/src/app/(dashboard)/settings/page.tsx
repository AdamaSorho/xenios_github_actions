'use client'

import { useAuth } from '@/presentation/hooks/useAuth'

export default function SettingsPage() {
  const { user } = useAuth()

  return (
    <div data-testid="settings-page">
      <h1 className="mb-6 text-2xl font-bold text-gray-900">Settings</h1>

      <div className="space-y-6">
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">Profile</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Name
              </label>
              <p className="mt-1 text-sm text-gray-900">
                {user?.name || '--'}
              </p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Email
              </label>
              <p className="mt-1 text-sm text-gray-900">
                {user?.email || '--'}
              </p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Role
              </label>
              <p className="mt-1 text-sm capitalize text-gray-900">
                {user?.role || '--'}
              </p>
            </div>
          </div>
        </div>

        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Account
          </h2>
          <p className="text-sm text-gray-500">
            Account management features are coming soon.
          </p>
        </div>
      </div>
    </div>
  )
}
