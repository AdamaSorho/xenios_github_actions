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

describe('ApiInsightRepository', () => {
  let repo: ApiInsightRepository

  beforeEach(() => {
    jest.clearAllMocks()
    repo = new ApiInsightRepository()
  })

  describe('getQueue', () => {
    test('getQueue_Success_ReturnsQueueResponse', async () => {
      const response = {
        insights: [mockInsight],
        pagination: { page: 1, limit: 20, total: 1 },
      }
      mockedApiClient.get.mockResolvedValue({ data: response, error: null, ok: true })

      const result = await repo.getQueue('draft', 1, 20)

      expect(result).toEqual(response)
      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/insights/queue?status=draft&page=1&limit=20')
    })

    test('getQueue_NoParams_CallsWithoutQuery', async () => {
      const response = {
        insights: [],
        pagination: { page: 1, limit: 20, total: 0 },
      }
      mockedApiClient.get.mockResolvedValue({ data: response, error: null, ok: true })

      await repo.getQueue()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/insights/queue')
    })

    test('getQueue_ApiError_ThrowsError', async () => {
      mockedApiClient.get.mockResolvedValue({ data: null, error: 'Unauthorized', ok: false })

      await expect(repo.getQueue()).rejects.toThrow('Unauthorized')
    })
  })

  describe('getClientInsights', () => {
    test('getClientInsights_Success_ReturnsInsights', async () => {
      const response = {
        insights: [mockInsight],
        pagination: { page: 1, limit: 20, total: 1 },
      }
      mockedApiClient.get.mockResolvedValue({ data: response, error: null, ok: true })

      const result = await repo.getClientInsights('client-1', 'draft')

      expect(result).toEqual(response)
      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/clients/client-1/insights?status=draft')
    })

    test('getClientInsights_ApiError_ThrowsError', async () => {
      mockedApiClient.get.mockResolvedValue({ data: null, error: 'Not found', ok: false })

      await expect(repo.getClientInsights('client-1')).rejects.toThrow('Not found')
    })
  })

  describe('approve', () => {
    test('approve_Success_ReturnsApprovedInsight', async () => {
      const approved = { ...mockInsight, status: 'approved' }
      mockedApiClient.put.mockResolvedValue({ data: approved, error: null, ok: true })

      const result = await repo.approve('insight-1')

      expect(result.status).toBe('approved')
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/insight-1/approve')
    })

    test('approve_ApiError_ThrowsError', async () => {
      mockedApiClient.put.mockResolvedValue({ data: null, error: 'Cannot approve', ok: false })

      await expect(repo.approve('insight-1')).rejects.toThrow('Cannot approve')
    })
  })

  describe('dismiss', () => {
    test('dismiss_Success_ReturnsDismissedInsight', async () => {
      const dismissed = { ...mockInsight, status: 'dismissed' }
      mockedApiClient.put.mockResolvedValue({ data: dismissed, error: null, ok: true })

      const result = await repo.dismiss('insight-1')

      expect(result.status).toBe('dismissed')
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/insight-1/dismiss')
    })

    test('dismiss_ApiError_ThrowsError', async () => {
      mockedApiClient.put.mockResolvedValue({ data: null, error: 'Cannot dismiss', ok: false })

      await expect(repo.dismiss('insight-1')).rejects.toThrow('Cannot dismiss')
    })
  })

  describe('edit', () => {
    test('edit_Success_ReturnsUpdatedInsight', async () => {
      const updated = { ...mockInsight, title: 'Updated' }
      mockedApiClient.put.mockResolvedValue({ data: updated, error: null, ok: true })

      const result = await repo.edit('insight-1', { title: 'Updated' })

      expect(result.title).toBe('Updated')
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/insight-1', { title: 'Updated' })
    })

    test('edit_ApiError_ThrowsError', async () => {
      mockedApiClient.put.mockResolvedValue({ data: null, error: 'Cannot edit', ok: false })

      await expect(repo.edit('insight-1', { title: 'X' })).rejects.toThrow('Cannot edit')
    })
  })

  describe('share', () => {
    test('share_Success_ReturnsSharedInsight', async () => {
      const shared = { ...mockInsight, status: 'shared' }
      mockedApiClient.put.mockResolvedValue({ data: shared, error: null, ok: true })

      const result = await repo.share('insight-1')

      expect(result.status).toBe('shared')
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/insight-1/share')
    })

    test('share_ApiError_ThrowsError', async () => {
      mockedApiClient.put.mockResolvedValue({ data: null, error: 'Cannot share', ok: false })

      await expect(repo.share('insight-1')).rejects.toThrow('Cannot share')
    })
  })
})
