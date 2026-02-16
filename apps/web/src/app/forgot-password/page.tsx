'use client'

import Link from 'next/link'

export default function ForgotPasswordPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <h1 className="text-2xl font-bold text-gray-900">Xenios</h1>
          <p className="mt-2 text-sm text-gray-600">Reset your password</p>
        </div>
        <div className="rounded-lg bg-white p-8 shadow-sm">
          <p className="mb-4 text-sm text-gray-600">
            Password reset functionality is coming soon. Please contact support
            if you need to reset your password.
          </p>
          <Link
            href="/login"
            className="block text-center text-sm text-blue-600 hover:text-blue-800"
          >
            Back to sign in
          </Link>
        </div>
      </div>
    </div>
  )
}
