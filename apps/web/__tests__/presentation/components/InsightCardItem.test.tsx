import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { InsightCardItem } from '@/presentation/components/InsightCardItem'
import { InsightCard } from '@/domain/entities/InsightCard'

const draftInsight: InsightCard = {
  id: 'i-1',
  coachId: 'coach-1',
  clientId: 'client-1',
  clientName: 'Marcus Johnson',
  title: 'Elevated LDL Cholesterol',
  body: "Marcus's LDL-C is 142 mg/dL, above the recommended <100 mg/dL threshold.",
  category: 'nutrition',
  status: 'draft',
  priority: 'high',
  evidence: [
    { measurementId: 'm-1', description: 'LDL-C: 142 mg/dL (ref: <100)' },
  ],
  createdAt: '2026-02-15T10:30:00Z',
  updatedAt: '2026-02-15T10:30:00Z',
}

const approvedInsight: InsightCard = {
  ...draftInsight,
  id: 'i-2',
  status: 'approved',
  approvedAt: '2026-02-15T11:00:00Z',
}

describe('InsightCardItem', () => {
  const mockApprove = jest.fn()
  const mockDismiss = jest.fn()
  const mockEdit = jest.fn()
  const mockShare = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders_DraftInsight_ShowsAllFields', () => {
    render(
      <InsightCardItem
        insight={draftInsight}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
      />
    )

    expect(screen.getByTestId('insight-title')).toHaveTextContent('Elevated LDL Cholesterol')
    expect(screen.getByTestId('insight-body')).toHaveTextContent("Marcus's LDL-C is 142 mg/dL")
    expect(screen.getByTestId('priority-badge')).toHaveTextContent('high')
    expect(screen.getByTestId('category-badge')).toHaveTextContent('nutrition')
    expect(screen.getByTestId('client-name')).toHaveTextContent('Marcus Johnson')
    expect(screen.getByTestId('evidence-section')).toBeInTheDocument()
  })

  test('renders_DraftInsight_ShowsActionButtons', () => {
    render(
      <InsightCardItem
        insight={draftInsight}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
      />
    )

    expect(screen.getByTestId('approve-button')).toBeInTheDocument()
    expect(screen.getByTestId('dismiss-button')).toBeInTheDocument()
    expect(screen.getByTestId('edit-button')).toBeInTheDocument()
  })

  test('clickApprove_DraftInsight_CallsOnApprove', () => {
    render(
      <InsightCardItem
        insight={draftInsight}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
      />
    )

    fireEvent.click(screen.getByTestId('approve-button'))
    expect(mockApprove).toHaveBeenCalledWith('i-1')
  })

  test('clickDismiss_DraftInsight_CallsOnDismiss', () => {
    render(
      <InsightCardItem
        insight={draftInsight}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
      />
    )

    fireEvent.click(screen.getByTestId('dismiss-button'))
    expect(mockDismiss).toHaveBeenCalledWith('i-1')
  })

  test('clickEdit_DraftInsight_ShowsEditForm', () => {
    render(
      <InsightCardItem
        insight={draftInsight}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
      />
    )

    fireEvent.click(screen.getByTestId('edit-button'))

    expect(screen.getByTestId('edit-form')).toBeInTheDocument()
    expect(screen.getByTestId('edit-title-input')).toHaveValue('Elevated LDL Cholesterol')
  })

  test('saveEdit_DraftInsight_CallsOnEdit', () => {
    render(
      <InsightCardItem
        insight={draftInsight}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
      />
    )

    fireEvent.click(screen.getByTestId('edit-button'))

    const titleInput = screen.getByTestId('edit-title-input')
    fireEvent.change(titleInput, { target: { value: 'New Title' } })

    const bodyInput = screen.getByTestId('edit-body-input')
    fireEvent.change(bodyInput, { target: { value: 'New Body' } })

    fireEvent.click(screen.getByTestId('save-edit-button'))

    expect(mockEdit).toHaveBeenCalledWith('i-1', 'New Title', 'New Body')
  })

  test('cancelEdit_DraftInsight_HidesEditForm', () => {
    render(
      <InsightCardItem
        insight={draftInsight}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
      />
    )

    fireEvent.click(screen.getByTestId('edit-button'))
    fireEvent.click(screen.getByTestId('cancel-edit-button'))

    expect(screen.queryByTestId('edit-form')).not.toBeInTheDocument()
    expect(screen.getByTestId('insight-title')).toBeInTheDocument()
  })

  test('renders_ApprovedInsight_ShowsShareButton', () => {
    render(
      <InsightCardItem
        insight={approvedInsight}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
        onShare={mockShare}
      />
    )

    expect(screen.getByTestId('share-button')).toBeInTheDocument()
    expect(screen.queryByTestId('approve-button')).not.toBeInTheDocument()
    expect(screen.queryByTestId('dismiss-button')).not.toBeInTheDocument()
  })

  test('clickShare_ApprovedInsight_CallsOnShare', () => {
    render(
      <InsightCardItem
        insight={approvedInsight}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
        onShare={mockShare}
      />
    )

    fireEvent.click(screen.getByTestId('share-button'))
    expect(mockShare).toHaveBeenCalledWith('i-2')
  })

  test('renders_InsightWithoutEvidence_NoEvidenceSection', () => {
    const noEvidence = { ...draftInsight, evidence: undefined }
    render(
      <InsightCardItem
        insight={noEvidence}
        onApprove={mockApprove}
        onDismiss={mockDismiss}
        onEdit={mockEdit}
      />
    )

    expect(screen.queryByTestId('evidence-section')).not.toBeInTheDocument()
  })
})
