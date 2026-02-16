import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { ErrorBoundary } from '@/presentation/components/ErrorBoundary'

let shouldThrow = false

function ThrowingComponent() {
  if (shouldThrow) {
    throw new Error('Test error')
  }
  return <div>Content rendered</div>
}

describe('ErrorBoundary', () => {
  const originalError = console.error
  beforeAll(() => {
    console.error = jest.fn()
  })
  afterAll(() => {
    console.error = originalError
  })

  beforeEach(() => {
    shouldThrow = false
  })

  test('renders_Children_WhenNoError', () => {
    render(
      <ErrorBoundary>
        <ThrowingComponent />
      </ErrorBoundary>
    )

    expect(screen.getByText('Content rendered')).toBeInTheDocument()
  })

  test('renders_ErrorFallback_WhenChildThrows', () => {
    shouldThrow = true
    render(
      <ErrorBoundary>
        <ThrowingComponent />
      </ErrorBoundary>
    )

    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    expect(screen.getByText('Test error')).toBeInTheDocument()
    expect(screen.getByTestId('error-retry-button')).toBeInTheDocument()
  })

  test('renders_CustomFallback_WhenProvided', () => {
    shouldThrow = true
    render(
      <ErrorBoundary fallback={<div>Custom error UI</div>}>
        <ThrowingComponent />
      </ErrorBoundary>
    )

    expect(screen.getByText('Custom error UI')).toBeInTheDocument()
  })

  test('retry_ResetsError_RendersChildren', () => {
    shouldThrow = true
    render(
      <ErrorBoundary>
        <ThrowingComponent />
      </ErrorBoundary>
    )

    expect(screen.getByText('Something went wrong')).toBeInTheDocument()

    // Fix the component before clicking retry
    shouldThrow = false

    // Click retry - this resets the error boundary state
    fireEvent.click(screen.getByTestId('error-retry-button'))

    // After reset, children should render successfully
    expect(screen.getByText('Content rendered')).toBeInTheDocument()
  })
})
