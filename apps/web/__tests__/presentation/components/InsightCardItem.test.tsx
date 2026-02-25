import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { InsightCardItem } from '@/presentation/components/InsightCardItem'
import { InsightCard } from '@/domain/entities/InsightCard'

const draftInsight: InsightCard = {
  id: 'insight-1',
  coachId: 'coach-1',
  clientId: 'client-1',
  clientName: 'Marcus Johnson',
  title: 'Elevated LDL Cholesterol',
  body: "Marcus's LDL-C is 142 mg/dL, above the recommended range.",
  category: 'nutrition',
  status: 'draft',
  priority: 'high',
  evidence: [
    { measurementId: 'm1', description: 'LDL-C: 142 mg/dL (ref: <100)' },
  ],
  createdAt: '2026-02-15T10:30:00Z',
  updatedAt: '2026-02-15T10:30:00Z',
}

const approvedInsight: InsightCard = {
  ...draftInsight,
  id: 'insight-2',
  status: 'approved',
  approvedAt: '2026-02-16T10:00:00Z',
}

const mockOnApprove = jest.fn().mockResolvedValue(undefined)
const mockOnDismiss = jest.fn().mockResolvedValue(undefined)
const mockOnEdit = jest.fn().mockResolvedValue(undefined)
const mockOnShare = jest.fn().mockResolvedValue(undefined)

function renderCard(insight: InsightCard = draftInsight, onShare?: (id: string) => Promise<void>) {
  return render(
    <InsightCardItem
      insight={insight}
      onApprove={mockOnApprove}
      onDismiss={mockOnDismiss}
      onEdit={mockOnEdit}
      onShare={onShare}
    />
  )
}

describe('InsightCardItem', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders_DraftInsight_ShowsTitleAndBody', () => {
    renderCard()

    expect(screen.getByTestId('insight-title')).toHaveTextContent('Elevated LDL Cholesterol')
    expect(screen.getByTestId('insight-body')).toHaveTextContent("Marcus's LDL-C is 142 mg/dL")
    expect(screen.getByTestId('insight-priority')).toHaveTextContent('high')
    expect(screen.getByTestId('insight-category')).toHaveTextContent('nutrition')
    expect(screen.getByTestId('insight-client-name')).toHaveTextContent('Marcus Johnson')
  })

  test('renders_DraftInsight_ShowsActionButtons', () => {
    renderCard()

    expect(screen.getByTestId('insight-approve-btn')).toBeInTheDocument()
    expect(screen.getByTestId('insight-dismiss-btn')).toBeInTheDocument()
    expect(screen.getByTestId('insight-edit-btn')).toBeInTheDocument()
  })

  test('renders_Evidence_WhenPresent', () => {
    renderCard()

    expect(screen.getByTestId('insight-evidence')).toBeInTheDocument()
    expect(screen.getByText('LDL-C: 142 mg/dL (ref: <100)')).toBeInTheDocument()
  })

  test('approve_ClickButton_CallsOnApprove', async () => {
    renderCard()

    fireEvent.click(screen.getByTestId('insight-approve-btn'))

    await waitFor(() => {
      expect(mockOnApprove).toHaveBeenCalledWith('insight-1')
    })
  })

  test('dismiss_ClickButton_CallsOnDismiss', async () => {
    renderCard()

    fireEvent.click(screen.getByTestId('insight-dismiss-btn'))

    await waitFor(() => {
      expect(mockOnDismiss).toHaveBeenCalledWith('insight-1')
    })
  })

  test('edit_ClickEdit_ShowsEditForm', () => {
    renderCard()

    fireEvent.click(screen.getByTestId('insight-edit-btn'))

    expect(screen.getByTestId('insight-edit-form')).toBeInTheDocument()
    expect(screen.getByTestId('insight-edit-title')).toHaveValue('Elevated LDL Cholesterol')
    expect(screen.getByTestId('insight-edit-body')).toHaveValue("Marcus's LDL-C is 142 mg/dL, above the recommended range.")
  })

  test('edit_SaveChanges_CallsOnEdit', async () => {
    renderCard()

    fireEvent.click(screen.getByTestId('insight-edit-btn'))

    const titleInput = screen.getByTestId('insight-edit-title')
    const bodyInput = screen.getByTestId('insight-edit-body')

    fireEvent.change(titleInput, { target: { value: 'Updated Title' } })
    fireEvent.change(bodyInput, { target: { value: 'Updated Body' } })
    fireEvent.click(screen.getByTestId('insight-save-btn'))

    await waitFor(() => {
      expect(mockOnEdit).toHaveBeenCalledWith('insight-1', 'Updated Title', 'Updated Body')
    })
  })

  test('edit_ClickCancel_HidesEditForm', () => {
    renderCard()

    fireEvent.click(screen.getByTestId('insight-edit-btn'))
    expect(screen.getByTestId('insight-edit-form')).toBeInTheDocument()

    fireEvent.click(screen.getByTestId('insight-cancel-btn'))
    expect(screen.queryByTestId('insight-edit-form')).not.toBeInTheDocument()
    expect(screen.getByTestId('insight-title')).toBeInTheDocument()
  })

  test('renders_ApprovedInsight_ShowsShareButton', () => {
    renderCard(approvedInsight, mockOnShare)

    expect(screen.getByTestId('insight-share-btn')).toBeInTheDocument()
    expect(screen.queryByTestId('insight-approve-btn')).not.toBeInTheDocument()
    expect(screen.queryByTestId('insight-dismiss-btn')).not.toBeInTheDocument()
  })

  test('share_ClickButton_CallsOnShare', async () => {
    renderCard(approvedInsight, mockOnShare)

    fireEvent.click(screen.getByTestId('insight-share-btn'))

    await waitFor(() => {
      expect(mockOnShare).toHaveBeenCalledWith('insight-2')
    })
  })

  test('renders_ApprovedInsight_NoShareWithoutHandler', () => {
    renderCard(approvedInsight)
    expect(screen.queryByTestId('insight-share-btn')).not.toBeInTheDocument()
  })
})
