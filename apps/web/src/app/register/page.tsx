'use client'

import { RegisterForm } from '@/presentation/components/RegisterForm'

export default function RegisterPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <h1 className="text-2xl font-bold text-gray-900">Xenios</h1>
          <p className="mt-2 text-sm text-gray-600">
            Create your coaching account
          </p>
        </div>
        <div className="rounded-lg bg-white p-8 shadow-sm">
          <RegisterForm />
        </div>
      </div>
    </div>
  )
}
