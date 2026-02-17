import { EditInsightUseCase } from '@/application/usecases/EditInsightUseCase'
import { InsightRepository } from '@/domain/repositories/InsightRepository'
import { InsightCard } from '@/domain/entities/InsightCard'

describe('EditInsightUseCase', () => {
  let mockRepo: jest.Mocked<InsightRepository>
  let useCase: EditInsightUseCase

  const mockInsight: InsightCard = {
    id: 'i-1',
    coachId: 'coach-1',
    clientId: 'client-1',
    title: 'Updated Title',
    body: 'Updated Body',
    category: 'general',
    status: 'draft',
    priority: 'medium',
    createdAt: '2026-02-15T10:30:00Z',
    updatedAt: '2026-02-15T11:00:00Z',
  }

  beforeEach(() => {
    mockRepo = {
      getQueue: jest.fn(),
      getClientInsights: jest.fn(),
      approve: jest.fn(),
      dismiss: jest.fn(),
      edit: jest.fn(),
      share: jest.fn(),
    }
    useCase = new EditInsightUseCase(mockRepo)
  })

  test('execute_ValidInput_ReturnsUpdatedInsight', async () => {
    mockRepo.edit.mockResolvedValue(mockInsight)

    const result = await useCase.execute('i-1', { title: 'Updated Title', body: 'Updated Body' })

    expect(result).toEqual(mockInsight)
    expect(mockRepo.edit).toHaveBeenCalledWith('i-1', { title: 'Updated Title', body: 'Updated Body' })
  })

  test('execute_EmptyInsightId_ThrowsError', async () => {
    await expect(useCase.execute('', { title: 'a', body: 'b' })).rejects.toThrow('Insight ID is required')
    expect(mockRepo.edit).not.toHaveBeenCalled()
  })

  test('execute_EmptyTitle_ThrowsError', async () => {
    await expect(useCase.execute('i-1', { title: '', body: 'body' })).rejects.toThrow('Title is required')
    expect(mockRepo.edit).not.toHaveBeenCalled()
  })

  test('execute_EmptyBody_ThrowsError', async () => {
    await expect(useCase.execute('i-1', { title: 'title', body: '' })).rejects.toThrow('Body is required')
    expect(mockRepo.edit).not.toHaveBeenCalled()
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockRepo.edit.mockRejectedValue(new Error('Failed to edit'))

    await expect(useCase.execute('i-1', { title: 'a', body: 'b' })).rejects.toThrow('Failed to edit')
  })
})
