'use client'

import { LoginForm } from '@/presentation/components/LoginForm'

export default function LoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <h1 className="text-2xl font-bold text-gray-900">Xenios</h1>
          <p className="mt-2 text-sm text-gray-600">
            Sign in to your coaching dashboard
          </p>
        </div>
        <div className="rounded-lg bg-white p-8 shadow-sm">
          <LoginForm />
        </div>
      </div>
    </div>
  )
}
