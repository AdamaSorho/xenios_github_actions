import { ApiInsightRepository } from '@/infrastructure/repositories/ApiInsightRepository'
import { apiClient } from '@xenios/api-client'

jest.mock('@xenios/api-client', () => ({
  apiClient: {
    get: jest.fn(),
    put: jest.fn(),
    post: jest.fn(),
    setAuthToken: jest.fn(),
    clearAuthToken: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

const mockInsight = {
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

const mockQueueResponse = {
  insights: [mockInsight],
  pagination: { page: 1, limit: 20, total: 1 },
}

describe('ApiInsightRepository', () => {
  let repo: ApiInsightRepository

  beforeEach(() => {
    jest.clearAllMocks()
    repo = new ApiInsightRepository()
  })

  describe('getQueue', () => {
    test('getQueue_NoParams_FetchesDefaultQueue', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: mockQueueResponse,
        error: null,
        ok: true,
      })

      const result = await repo.getQueue()
      expect(result).toEqual(mockQueueResponse)
      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/insights/queue')
    })

    test('getQueue_WithParams_IncludesQueryString', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: mockQueueResponse,
        error: null,
        ok: true,
      })

      await repo.getQueue('draft', 10, 5)
      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/v1/insights/queue?status=draft&limit=10&offset=5'
      )
    })

    test('getQueue_ApiError_ThrowsError', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: null,
        error: 'Unauthorized',
        ok: false,
      })

      await expect(repo.getQueue()).rejects.toThrow('Unauthorized')
    })
  })

  describe('getClientInsights', () => {
    test('getClientInsights_ValidClientId_FetchesInsights', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: mockQueueResponse,
        error: null,
        ok: true,
      })

      const result = await repo.getClientInsights('client-1')
      expect(result).toEqual(mockQueueResponse)
      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/clients/client-1/insights')
    })

    test('getClientInsights_WithStatusFilter_IncludesQueryString', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: mockQueueResponse,
        error: null,
        ok: true,
      })

      await repo.getClientInsights('client-1', 'approved')
      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/v1/clients/client-1/insights?status=approved'
      )
    })
  })

  describe('approve', () => {
    test('approve_ValidId_ReturnsApprovedInsight', async () => {
      const approved = { ...mockInsight, status: 'approved' }
      mockedApiClient.put.mockResolvedValue({
        data: approved,
        error: null,
        ok: true,
      })

      const result = await repo.approve('insight-1')
      expect(result.status).toBe('approved')
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/insight-1/approve')
    })

    test('approve_ApiError_ThrowsError', async () => {
      mockedApiClient.put.mockResolvedValue({
        data: null,
        error: 'Invalid transition',
        ok: false,
      })

      await expect(repo.approve('insight-1')).rejects.toThrow('Invalid transition')
    })
  })

  describe('dismiss', () => {
    test('dismiss_ValidId_ReturnsDismissedInsight', async () => {
      const dismissed = { ...mockInsight, status: 'dismissed' }
      mockedApiClient.put.mockResolvedValue({
        data: dismissed,
        error: null,
        ok: true,
      })

      const result = await repo.dismiss('insight-1')
      expect(result.status).toBe('dismissed')
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/insight-1/dismiss')
    })
  })

  describe('edit', () => {
    test('edit_ValidInput_ReturnsUpdatedInsight', async () => {
      const updated = { ...mockInsight, title: 'Updated Title' }
      mockedApiClient.put.mockResolvedValue({
        data: updated,
        error: null,
        ok: true,
      })

      const result = await repo.edit('insight-1', { title: 'Updated Title' })
      expect(result.title).toBe('Updated Title')
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/insight-1', {
        title: 'Updated Title',
      })
    })
  })

  describe('share', () => {
    test('share_ValidId_ReturnsSharedInsight', async () => {
      const shared = { ...mockInsight, status: 'shared' }
      mockedApiClient.put.mockResolvedValue({
        data: shared,
        error: null,
        ok: true,
      })

      const result = await repo.share('insight-1')
      expect(result.status).toBe('shared')
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/insight-1/share')
    })
  })
})
