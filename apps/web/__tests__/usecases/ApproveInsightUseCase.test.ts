import { ApproveInsightUseCase } from '@/application/usecases/ApproveInsightUseCase'
import { InsightRepository } from '@/domain/repositories/InsightRepository'
import { InsightCard } from '@/domain/entities/InsightCard'

describe('ApproveInsightUseCase', () => {
  let mockRepo: jest.Mocked<InsightRepository>
  let useCase: ApproveInsightUseCase

  const mockInsight: InsightCard = {
    id: 'i-1',
    coachId: 'coach-1',
    clientId: 'client-1',
    title: 'Test Insight',
    body: 'Test body',
    category: 'general',
    status: 'approved',
    priority: 'high',
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
    useCase = new ApproveInsightUseCase(mockRepo)
  })

  test('execute_ValidInsightId_ReturnsApprovedInsight', async () => {
    mockRepo.approve.mockResolvedValue(mockInsight)

    const result = await useCase.execute('i-1')

    expect(result).toEqual(mockInsight)
    expect(mockRepo.approve).toHaveBeenCalledWith('i-1')
  })

  test('execute_EmptyInsightId_ThrowsError', async () => {
    await expect(useCase.execute('')).rejects.toThrow('Insight ID is required')
    expect(mockRepo.approve).not.toHaveBeenCalled()
  })

  test('execute_RepositoryError_PropagatesError', async () => {
    mockRepo.approve.mockRejectedValue(new Error('Forbidden'))

    await expect(useCase.execute('i-1')).rejects.toThrow('Forbidden')
  })
})
