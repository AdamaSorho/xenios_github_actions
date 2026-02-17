import { GetClientInsightsUseCase } from '@/application/usecases/GetClientInsightsUseCase'
import { InsightRepository } from '@/domain/repositories/InsightRepository'
import { InsightQueueResponse } from '@/domain/entities/InsightCard'

describe('GetClientInsightsUseCase', () => {
  let mockRepo: jest.Mocked<InsightRepository>
  let useCase: GetClientInsightsUseCase

  const mockResponse: InsightQueueResponse = {
    insights: [
      {
        id: 'i-1',
        coachId: 'coach-1',
        clientId: 'client-1',
        title: 'Test',
        body: 'Body',
        category: 'nutrition',
        status: 'approved',
        priority: 'medium',
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
    useCase = new GetClientInsightsUseCase(mockRepo)
  })

  test('execute_ValidClientId_ReturnsInsights', async () => {
    mockRepo.getClientInsights.mockResolvedValue(mockResponse)

    const result = await useCase.execute('client-1', 'approved', 20, 0)

    expect(result).toEqual(mockResponse)
    expect(mockRepo.getClientInsights).toHaveBeenCalledWith('client-1', 'approved', 20, 0)
  })

  test('execute_EmptyClientId_ThrowsError', async () => {
    await expect(useCase.execute('')).rejects.toThrow('Client ID is required')
    expect(mockRepo.getClientInsights).not.toHaveBeenCalled()
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockRepo.getClientInsights.mockRejectedValue(new Error('Failed'))

    await expect(useCase.execute('client-1')).rejects.toThrow('Failed')
  })
})
