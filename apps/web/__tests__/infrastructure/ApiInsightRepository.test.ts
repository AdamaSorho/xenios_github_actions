import { ApiInsightRepository } from '@/infrastructure/repositories/ApiInsightRepository'
import { apiClient } from '@xenios/api-client'

jest.mock('@xenios/api-client', () => ({
  apiClient: {
    get: jest.fn(),
    put: jest.fn(),
  },
}))

const mockedApiClient = jest.mocked(apiClient)

describe('ApiInsightRepository', () => {
  let repo: ApiInsightRepository

  beforeEach(() => {
    jest.clearAllMocks()
    repo = new ApiInsightRepository()
  })

  describe('getQueue', () => {
    test('getQueue_WithStatusFilter_ReturnsInsights', async () => {
      const response = {
        insights: [{ id: 'i-1', status: 'draft' }],
        page: 1,
        limit: 20,
        total: 1,
      }

      mockedApiClient.get.mockResolvedValue({
        data: response,
        error: null,
        ok: true,
      })

      const result = await repo.getQueue('draft', 1, 20)

      expect(result).toEqual(response)
      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/v1/insights/queue?status=draft&page=1&limit=20'
      )
    })

    test('getQueue_NoParams_FetchesAll', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: { insights: [], page: 1, limit: 20, total: 0 },
        error: null,
        ok: true,
      })

      await repo.getQueue()

      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/insights/queue')
    })

    test('getQueue_ApiError_ThrowsError', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: null,
        error: 'Unauthorized',
        ok: false,
      })

      await expect(repo.getQueue('draft')).rejects.toThrow('Unauthorized')
    })
  })

  describe('getClientInsights', () => {
    test('getClientInsights_ValidClientId_ReturnsInsights', async () => {
      const response = {
        insights: [{ id: 'i-1', clientId: 'client-1' }],
        page: 1,
        limit: 20,
        total: 1,
      }

      mockedApiClient.get.mockResolvedValue({
        data: response,
        error: null,
        ok: true,
      })

      const result = await repo.getClientInsights('client-1', 'approved', 20, 0)

      expect(result).toEqual(response)
      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/v1/clients/client-1/insights?status=approved&limit=20&offset=0'
      )
    })

    test('getClientInsights_ApiError_ThrowsError', async () => {
      mockedApiClient.get.mockResolvedValue({
        data: null,
        error: 'Not found',
        ok: false,
      })

      await expect(repo.getClientInsights('client-1')).rejects.toThrow('Not found')
    })
  })

  describe('approve', () => {
    test('approve_ValidId_ReturnsApprovedInsight', async () => {
      const insight = { id: 'i-1', status: 'approved' }

      mockedApiClient.put.mockResolvedValue({
        data: insight,
        error: null,
        ok: true,
      })

      const result = await repo.approve('i-1')

      expect(result).toEqual(insight)
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/i-1/approve')
    })

    test('approve_ApiError_ThrowsError', async () => {
      mockedApiClient.put.mockResolvedValue({
        data: null,
        error: 'Forbidden',
        ok: false,
      })

      await expect(repo.approve('i-1')).rejects.toThrow('Forbidden')
    })
  })

  describe('dismiss', () => {
    test('dismiss_ValidId_ReturnsDismissedInsight', async () => {
      const insight = { id: 'i-1', status: 'dismissed' }

      mockedApiClient.put.mockResolvedValue({
        data: insight,
        error: null,
        ok: true,
      })

      const result = await repo.dismiss('i-1')

      expect(result).toEqual(insight)
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/i-1/dismiss')
    })
  })

  describe('edit', () => {
    test('edit_ValidInput_ReturnsUpdatedInsight', async () => {
      const insight = { id: 'i-1', title: 'New', body: 'New body' }

      mockedApiClient.put.mockResolvedValue({
        data: insight,
        error: null,
        ok: true,
      })

      const result = await repo.edit('i-1', { title: 'New', body: 'New body' })

      expect(result).toEqual(insight)
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/i-1', {
        title: 'New',
        body: 'New body',
      })
    })

    test('edit_ApiError_ThrowsError', async () => {
      mockedApiClient.put.mockResolvedValue({
        data: null,
        error: 'Validation error',
        ok: false,
      })

      await expect(repo.edit('i-1', { title: '', body: '' })).rejects.toThrow('Validation error')
    })
  })

  describe('share', () => {
    test('share_ValidId_ReturnsSharedInsight', async () => {
      const insight = { id: 'i-1', status: 'shared' }

      mockedApiClient.put.mockResolvedValue({
        data: insight,
        error: null,
        ok: true,
      })

      const result = await repo.share('i-1')

      expect(result).toEqual(insight)
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/i-1/share')
    })

    test('share_ApiError_ThrowsError', async () => {
      mockedApiClient.put.mockResolvedValue({
        data: null,
        error: 'Invalid transition',
        ok: false,
      })

      await expect(repo.share('i-1')).rejects.toThrow('Invalid transition')
    })
  })
})
