import { InsightRepository } from '@/domain/repositories/InsightRepository'
import { InsightCard, InsightQueueResponse } from '@/domain/entities/InsightCard'
import { GetInsightQueueUseCase } from '@/application/usecases/GetInsightQueueUseCase'
import { GetClientInsightsUseCase } from '@/application/usecases/GetClientInsightsUseCase'
import { ApproveInsightUseCase } from '@/application/usecases/ApproveInsightUseCase'
import { DismissInsightUseCase } from '@/application/usecases/DismissInsightUseCase'
import { EditInsightUseCase } from '@/application/usecases/EditInsightUseCase'
import { ShareInsightUseCase } from '@/application/usecases/ShareInsightUseCase'

function createMockInsightRepo(): jest.Mocked<InsightRepository> {
  return {
    getQueue: jest.fn(),
    getClientInsights: jest.fn(),
    approve: jest.fn(),
    dismiss: jest.fn(),
    edit: jest.fn(),
    share: jest.fn(),
  }
}

const mockInsight: InsightCard = {
  id: 'insight-1',
  coachId: 'coach-1',
  clientId: 'client-1',
  title: 'Test Insight',
  body: 'Test body',
  category: 'nutrition',
  status: 'draft',
  priority: 'high',
  createdAt: '2026-02-15T10:00:00Z',
  updatedAt: '2026-02-15T10:00:00Z',
}

const mockQueueResponse: InsightQueueResponse = {
  insights: [mockInsight],
  pagination: { page: 1, limit: 20, total: 1 },
}

describe('GetInsightQueueUseCase', () => {
  let repo: jest.Mocked<InsightRepository>
  let uc: GetInsightQueueUseCase

  beforeEach(() => {
    repo = createMockInsightRepo()
    uc = new GetInsightQueueUseCase(repo)
  })

  test('execute_DefaultParams_ReturnsQueue', async () => {
    repo.getQueue.mockResolvedValue(mockQueueResponse)
    const result = await uc.execute()
    expect(result).toEqual(mockQueueResponse)
    expect(repo.getQueue).toHaveBeenCalledWith(undefined, undefined, undefined)
  })

  test('execute_WithStatus_PassesStatusToRepo', async () => {
    repo.getQueue.mockResolvedValue(mockQueueResponse)
    await uc.execute('draft', 10, 0)
    expect(repo.getQueue).toHaveBeenCalledWith('draft', 10, 0)
  })

  test('execute_RepoError_PropagatesError', async () => {
    repo.getQueue.mockRejectedValue(new Error('network error'))
    await expect(uc.execute()).rejects.toThrow('network error')
  })
})

describe('GetClientInsightsUseCase', () => {
  let repo: jest.Mocked<InsightRepository>
  let uc: GetClientInsightsUseCase

  beforeEach(() => {
    repo = createMockInsightRepo()
    uc = new GetClientInsightsUseCase(repo)
  })

  test('execute_ValidClientId_ReturnsInsights', async () => {
    repo.getClientInsights.mockResolvedValue(mockQueueResponse)
    const result = await uc.execute('client-1')
    expect(result).toEqual(mockQueueResponse)
    expect(repo.getClientInsights).toHaveBeenCalledWith('client-1', undefined, undefined, undefined)
  })

  test('execute_EmptyClientId_ThrowsError', async () => {
    await expect(uc.execute('')).rejects.toThrow('client_id is required')
    expect(repo.getClientInsights).not.toHaveBeenCalled()
  })
})

describe('ApproveInsightUseCase', () => {
  let repo: jest.Mocked<InsightRepository>
  let uc: ApproveInsightUseCase

  beforeEach(() => {
    repo = createMockInsightRepo()
    uc = new ApproveInsightUseCase(repo)
  })

  test('execute_ValidId_ReturnsApprovedInsight', async () => {
    const approved = { ...mockInsight, status: 'approved' as const }
    repo.approve.mockResolvedValue(approved)
    const result = await uc.execute('insight-1')
    expect(result.status).toBe('approved')
    expect(repo.approve).toHaveBeenCalledWith('insight-1')
  })

  test('execute_EmptyId_ThrowsError', async () => {
    await expect(uc.execute('')).rejects.toThrow('insight_id is required')
    expect(repo.approve).not.toHaveBeenCalled()
  })
})

describe('DismissInsightUseCase', () => {
  let repo: jest.Mocked<InsightRepository>
  let uc: DismissInsightUseCase

  beforeEach(() => {
    repo = createMockInsightRepo()
    uc = new DismissInsightUseCase(repo)
  })

  test('execute_ValidId_ReturnsDismissedInsight', async () => {
    const dismissed = { ...mockInsight, status: 'dismissed' as const }
    repo.dismiss.mockResolvedValue(dismissed)
    const result = await uc.execute('insight-1')
    expect(result.status).toBe('dismissed')
  })

  test('execute_EmptyId_ThrowsError', async () => {
    await expect(uc.execute('')).rejects.toThrow('insight_id is required')
  })
})

describe('EditInsightUseCase', () => {
  let repo: jest.Mocked<InsightRepository>
  let uc: EditInsightUseCase

  beforeEach(() => {
    repo = createMockInsightRepo()
    uc = new EditInsightUseCase(repo)
  })

  test('execute_ValidInput_ReturnsUpdatedInsight', async () => {
    const updated = { ...mockInsight, title: 'Updated' }
    repo.edit.mockResolvedValue(updated)
    const result = await uc.execute('insight-1', { title: 'Updated' })
    expect(result.title).toBe('Updated')
    expect(repo.edit).toHaveBeenCalledWith('insight-1', { title: 'Updated' })
  })

  test('execute_EmptyId_ThrowsError', async () => {
    await expect(uc.execute('', { title: 'x' })).rejects.toThrow('insight_id is required')
  })

  test('execute_EmptyTitleAndBody_ThrowsError', async () => {
    await expect(uc.execute('insight-1', {})).rejects.toThrow('title or body is required')
  })
})

describe('ShareInsightUseCase', () => {
  let repo: jest.Mocked<InsightRepository>
  let uc: ShareInsightUseCase

  beforeEach(() => {
    repo = createMockInsightRepo()
    uc = new ShareInsightUseCase(repo)
  })

  test('execute_ValidId_ReturnsSharedInsight', async () => {
    const shared = { ...mockInsight, status: 'shared' as const }
    repo.share.mockResolvedValue(shared)
    const result = await uc.execute('insight-1')
    expect(result.status).toBe('shared')
  })

  test('execute_EmptyId_ThrowsError', async () => {
    await expect(uc.execute('')).rejects.toThrow('insight_id is required')
  })
})
