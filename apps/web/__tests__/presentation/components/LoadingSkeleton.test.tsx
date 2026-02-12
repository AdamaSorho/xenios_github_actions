import React from 'react'
import { render, screen } from '@testing-library/react'
import {
  LoadingSkeleton,
  CardSkeleton,
  DashboardSkeleton,
  TableSkeleton,
} from '@/presentation/components/LoadingSkeleton'

describe('LoadingSkeleton', () => {
  test('renders_WithDefaultClass', () => {
    render(<LoadingSkeleton />)
    expect(screen.getByTestId('loading-skeleton')).toBeInTheDocument()
  })

  test('renders_WithCustomClass', () => {
    render(<LoadingSkeleton className="h-8 w-48" />)
    const skeleton = screen.getByTestId('loading-skeleton')
    expect(skeleton.className).toContain('h-8')
    expect(skeleton.className).toContain('w-48')
  })
})

describe('CardSkeleton', () => {
  test('renders_CardSkeleton', () => {
    render(<CardSkeleton />)
    expect(screen.getByTestId('card-skeleton')).toBeInTheDocument()
  })
})

describe('DashboardSkeleton', () => {
  test('renders_DashboardSkeleton', () => {
    render(<DashboardSkeleton />)
    expect(screen.getByTestId('dashboard-skeleton')).toBeInTheDocument()
  })
})

describe('TableSkeleton', () => {
  test('renders_DefaultRows', () => {
    render(<TableSkeleton />)
    expect(screen.getByTestId('table-skeleton')).toBeInTheDocument()
  })

  test('renders_CustomRowCount', () => {
    render(<TableSkeleton rows={3} />)
    const skeleton = screen.getByTestId('table-skeleton')
    // 1 header + 3 rows = 4 child divs
    expect(skeleton.children.length).toBe(4)
  })
})
