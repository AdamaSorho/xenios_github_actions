import { GetInsightQueueUseCase } from '@/application/usecases/GetInsightQueueUseCase'
import { InsightRepository } from '@/domain/repositories/InsightRepository'
import { InsightQueueResponse } from '@/domain/entities/InsightCard'

describe('GetInsightQueueUseCase', () => {
  let mockRepo: jest.Mocked<InsightRepository>
  let useCase: GetInsightQueueUseCase

  const mockResponse: InsightQueueResponse = {
    insights: [
      {
        id: 'i-1',
        coachId: 'coach-1',
        clientId: 'client-1',
        title: 'Test',
        body: 'Body',
        category: 'nutrition',
        status: 'draft',
        priority: 'high',
        createdAt: '2026-02-15T10:30:00Z',
        updatedAt: '2026-02-15T10:30:00Z',
      },
    ],
    page: 1,
    limit: 20,
    total: 1,
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
    useCase = new GetInsightQueueUseCase(mockRepo)
  })

  test('execute_WithStatus_ReturnsFilteredQueue', async () => {
    mockRepo.getQueue.mockResolvedValue(mockResponse)

    const result = await useCase.execute('draft', 1, 20)

    expect(result).toEqual(mockResponse)
    expect(mockRepo.getQueue).toHaveBeenCalledWith('draft', 1, 20)
  })

  test('execute_NoParams_ReturnsAllQueue', async () => {
    mockRepo.getQueue.mockResolvedValue({ insights: [], page: 1, limit: 20, total: 0 })

    const result = await useCase.execute()

    expect(result.insights).toEqual([])
    expect(mockRepo.getQueue).toHaveBeenCalledWith(undefined, undefined, undefined)
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockRepo.getQueue.mockRejectedValue(new Error('Failed'))

    await expect(useCase.execute('draft')).rejects.toThrow('Failed')
  })
})
