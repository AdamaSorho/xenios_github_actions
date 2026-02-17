import { ShareInsightUseCase } from '@/application/usecases/ShareInsightUseCase'
import { InsightRepository } from '@/domain/repositories/InsightRepository'
import { InsightCard } from '@/domain/entities/InsightCard'

describe('ShareInsightUseCase', () => {
  let mockRepo: jest.Mocked<InsightRepository>
  let useCase: ShareInsightUseCase

  const mockInsight: InsightCard = {
    id: 'i-1',
    coachId: 'coach-1',
    clientId: 'client-1',
    title: 'Test',
    body: 'Body',
    category: 'general',
    status: 'shared',
    priority: 'medium',
    createdAt: '2026-02-15T10:30:00Z',
    updatedAt: '2026-02-15T10:30:00Z',
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
    useCase = new ShareInsightUseCase(mockRepo)
  })

  test('execute_ValidInsightId_ReturnsSharedInsight', async () => {
    mockRepo.share.mockResolvedValue(mockInsight)

    const result = await useCase.execute('i-1')

    expect(result).toEqual(mockInsight)
    expect(mockRepo.share).toHaveBeenCalledWith('i-1')
  })

  test('execute_EmptyInsightId_ThrowsError', async () => {
    await expect(useCase.execute('')).rejects.toThrow('Insight ID is required')
    expect(mockRepo.share).not.toHaveBeenCalled()
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockRepo.share.mockRejectedValue(new Error('Invalid transition'))

    await expect(useCase.execute('i-1')).rejects.toThrow('Invalid transition')
  })
})
