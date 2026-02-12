'use client'

export function LoadingSkeleton({ className = '' }: { className?: string }) {
  return (
    <div
      className={`animate-pulse rounded bg-gray-200 ${className}`}
      data-testid="loading-skeleton"
    />
  )
}

export function CardSkeleton() {
  return (
    <div className="rounded-lg border border-gray-200 p-6" data-testid="card-skeleton">
      <LoadingSkeleton className="mb-4 h-4 w-1/3" />
      <LoadingSkeleton className="mb-2 h-8 w-1/2" />
      <LoadingSkeleton className="h-4 w-2/3" />
    </div>
  )
}

export function DashboardSkeleton() {
  return (
    <div className="space-y-6" data-testid="dashboard-skeleton">
      <LoadingSkeleton className="h-8 w-48" />
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        <CardSkeleton />
        <CardSkeleton />
        <CardSkeleton />
        <CardSkeleton />
      </div>
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="rounded-lg border border-gray-200 p-6">
          <LoadingSkeleton className="mb-4 h-6 w-40" />
          <div className="space-y-3">
            <LoadingSkeleton className="h-12 w-full" />
            <LoadingSkeleton className="h-12 w-full" />
            <LoadingSkeleton className="h-12 w-full" />
          </div>
        </div>
        <div className="rounded-lg border border-gray-200 p-6">
          <LoadingSkeleton className="mb-4 h-6 w-40" />
          <LoadingSkeleton className="h-48 w-full" />
        </div>
      </div>
    </div>
  )
}

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-3" data-testid="table-skeleton">
      <LoadingSkeleton className="h-10 w-full" />
      {Array.from({ length: rows }).map((_, i) => (
        <LoadingSkeleton key={i} className="h-14 w-full" />
      ))}
    </div>
  )
}
