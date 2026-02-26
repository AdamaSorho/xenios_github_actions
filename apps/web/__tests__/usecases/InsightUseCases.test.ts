import { GetInsightQueueUseCase } from '@/application/usecases/GetInsightQueueUseCase'
import { ApproveInsightUseCase } from '@/application/usecases/ApproveInsightUseCase'
import { DismissInsightUseCase } from '@/application/usecases/DismissInsightUseCase'
import { EditInsightUseCase } from '@/application/usecases/EditInsightUseCase'
import { ShareInsightUseCase } from '@/application/usecases/ShareInsightUseCase'
import { InsightRepository } from '@/domain/repositories/InsightRepository'
import type { InsightCard, InsightQueueResponse, EditInsightInput } from '@/domain/entities/InsightCard'

class MockInsightRepository implements InsightRepository {
  getQueue = jest.fn<Promise<InsightQueueResponse>, [string?, number?, number?]>()
  getClientInsights = jest.fn<Promise<InsightQueueResponse>, [string, string?, number?, number?]>()
  approve = jest.fn<Promise<InsightCard>, [string]>()
  dismiss = jest.fn<Promise<InsightCard>, [string]>()
  edit = jest.fn<Promise<InsightCard>, [string, EditInsightInput]>()
  share = jest.fn<Promise<InsightCard>, [string]>()
}

const mockInsight: InsightCard = {
  id: 'insight-1',
  client_id: 'client-1',
  coach_id: 'coach-1',
  client_name: 'Alice',
  title: 'High LDL',
  body: 'LDL is elevated',
  category: 'nutrition',
  priority: 'high',
  status: 'draft',
  evidence: [],
  created_at: '2026-02-15T10:00:00Z',
  updated_at: '2026-02-15T10:00:00Z',
  approved_at: null,
  dismissed_at: null,
  shared_at: null,
}

describe('GetInsightQueueUseCase', () => {
  let repo: MockInsightRepository
  let useCase: GetInsightQueueUseCase

  beforeEach(() => {
    repo = new MockInsightRepository()
    useCase = new GetInsightQueueUseCase(repo)
  })

  test('execute_Success_ReturnsQueueResponse', async () => {
    const response: InsightQueueResponse = {
      insights: [mockInsight],
      pagination: { page: 1, limit: 20, total: 1 },
    }
    repo.getQueue.mockResolvedValue(response)

    const result = await useCase.execute('draft', 1, 20)

    expect(result).toEqual(response)
    expect(repo.getQueue).toHaveBeenCalledWith('draft', 1, 20)
  })

  test('execute_NoParams_CallsWithUndefined', async () => {
    const response: InsightQueueResponse = {
      insights: [],
      pagination: { page: 1, limit: 20, total: 0 },
    }
    repo.getQueue.mockResolvedValue(response)

    await useCase.execute()

    expect(repo.getQueue).toHaveBeenCalledWith(undefined, undefined, undefined)
  })

  test('execute_RepoError_PropagatesError', async () => {
    repo.getQueue.mockRejectedValue(new Error('Network error'))

    await expect(useCase.execute()).rejects.toThrow('Network error')
  })
})

describe('ApproveInsightUseCase', () => {
  let repo: MockInsightRepository
  let useCase: ApproveInsightUseCase

  beforeEach(() => {
    repo = new MockInsightRepository()
    useCase = new ApproveInsightUseCase(repo)
  })

  test('execute_ValidId_ReturnsApprovedInsight', async () => {
    const approved = { ...mockInsight, status: 'approved' as const }
    repo.approve.mockResolvedValue(approved)

    const result = await useCase.execute('insight-1')

    expect(result.status).toBe('approved')
    expect(repo.approve).toHaveBeenCalledWith('insight-1')
  })

  test('execute_EmptyId_ThrowsError', async () => {
    await expect(useCase.execute('')).rejects.toThrow('Insight ID is required')
  })
})

describe('DismissInsightUseCase', () => {
  let repo: MockInsightRepository
  let useCase: DismissInsightUseCase

  beforeEach(() => {
    repo = new MockInsightRepository()
    useCase = new DismissInsightUseCase(repo)
  })

  test('execute_ValidId_ReturnsDismissedInsight', async () => {
    const dismissed = { ...mockInsight, status: 'dismissed' as const }
    repo.dismiss.mockResolvedValue(dismissed)

    const result = await useCase.execute('insight-1')

    expect(result.status).toBe('dismissed')
    expect(repo.dismiss).toHaveBeenCalledWith('insight-1')
  })

  test('execute_EmptyId_ThrowsError', async () => {
    await expect(useCase.execute('')).rejects.toThrow('Insight ID is required')
  })
})

describe('EditInsightUseCase', () => {
  let repo: MockInsightRepository
  let useCase: EditInsightUseCase

  beforeEach(() => {
    repo = new MockInsightRepository()
    useCase = new EditInsightUseCase(repo)
  })

  test('execute_ValidInput_ReturnsUpdatedInsight', async () => {
    const updated = { ...mockInsight, title: 'Updated Title' }
    repo.edit.mockResolvedValue(updated)

    const result = await useCase.execute('insight-1', { title: 'Updated Title' })

    expect(result.title).toBe('Updated Title')
    expect(repo.edit).toHaveBeenCalledWith('insight-1', { title: 'Updated Title' })
  })

  test('execute_EmptyId_ThrowsError', async () => {
    await expect(useCase.execute('', { title: 'X' })).rejects.toThrow('Insight ID is required')
  })

  test('execute_EmptyTitleAndBody_ThrowsError', async () => {
    await expect(useCase.execute('insight-1', {})).rejects.toThrow('Title or body is required')
  })
})

describe('ShareInsightUseCase', () => {
  let repo: MockInsightRepository
  let useCase: ShareInsightUseCase

  beforeEach(() => {
    repo = new MockInsightRepository()
    useCase = new ShareInsightUseCase(repo)
  })

  test('execute_ValidId_ReturnsSharedInsight', async () => {
    const shared = { ...mockInsight, status: 'shared' as const }
    repo.share.mockResolvedValue(shared)

    const result = await useCase.execute('insight-1')

    expect(result.status).toBe('shared')
    expect(repo.share).toHaveBeenCalledWith('insight-1')
  })

  test('execute_EmptyId_ThrowsError', async () => {
    await expect(useCase.execute('')).rejects.toThrow('Insight ID is required')
  })
})
