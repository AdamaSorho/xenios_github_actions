import { DismissInsightUseCase } from '@/application/usecases/DismissInsightUseCase'
import { InsightRepository } from '@/domain/repositories/InsightRepository'
import { InsightCard } from '@/domain/entities/InsightCard'

describe('DismissInsightUseCase', () => {
  let mockRepo: jest.Mocked<InsightRepository>
  let useCase: DismissInsightUseCase

  const mockInsight: InsightCard = {
    id: 'i-1',
    coachId: 'coach-1',
    clientId: 'client-1',
    title: 'Test Insight',
    body: 'Test body',
    category: 'general',
    status: 'dismissed',
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
    useCase = new DismissInsightUseCase(mockRepo)
  })

  test('execute_ValidInsightId_ReturnsDismissedInsight', async () => {
    mockRepo.dismiss.mockResolvedValue(mockInsight)

    const result = await useCase.execute('i-1')

    expect(result).toEqual(mockInsight)
    expect(mockRepo.dismiss).toHaveBeenCalledWith('i-1')
  })

  test('execute_EmptyInsightId_ThrowsError', async () => {
    await expect(useCase.execute('')).rejects.toThrow('Insight ID is required')
    expect(mockRepo.dismiss).not.toHaveBeenCalled()
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockRepo.dismiss.mockRejectedValue(new Error('Not found'))

    await expect(useCase.execute('i-1')).rejects.toThrow('Not found')
  })
})
