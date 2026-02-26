import { ApiInsightRepository } from '@/infrastructure/repositories/ApiInsightRepository'
import { apiClient } from '@xenios/api-client'
import { mockInsight, mockQueueResponse } from '../helpers/insightFixtures'

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

const okResponse = <T>(data: T) => ({ data, error: null, ok: true })
const errorResponse = (error: string) => ({ data: null, error, ok: false })

describe('ApiInsightRepository', () => {
  let repo: ApiInsightRepository

  beforeEach(() => {
    jest.clearAllMocks()
    repo = new ApiInsightRepository()
  })

  describe('getQueue', () => {
    test('getQueue_NoParams_FetchesDefaultQueue', async () => {
      mockedApiClient.get.mockResolvedValue(okResponse(mockQueueResponse))

      const result = await repo.getQueue()
      expect(result).toEqual(mockQueueResponse)
      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/insights/queue')
    })

    test('getQueue_WithParams_IncludesQueryString', async () => {
      mockedApiClient.get.mockResolvedValue(okResponse(mockQueueResponse))

      await repo.getQueue('draft', 10, 5)
      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/v1/insights/queue?status=draft&limit=10&offset=5'
      )
    })

    test('getQueue_ApiError_ThrowsError', async () => {
      mockedApiClient.get.mockResolvedValue(errorResponse('Unauthorized'))
      await expect(repo.getQueue()).rejects.toThrow('Unauthorized')
    })
  })

  describe('getClientInsights', () => {
    test('getClientInsights_ValidClientId_FetchesInsights', async () => {
      mockedApiClient.get.mockResolvedValue(okResponse(mockQueueResponse))

      const result = await repo.getClientInsights('client-1')
      expect(result).toEqual(mockQueueResponse)
      expect(mockedApiClient.get).toHaveBeenCalledWith('/v1/clients/client-1/insights')
    })

    test('getClientInsights_WithStatusFilter_IncludesQueryString', async () => {
      mockedApiClient.get.mockResolvedValue(okResponse(mockQueueResponse))

      await repo.getClientInsights('client-1', 'approved')
      expect(mockedApiClient.get).toHaveBeenCalledWith(
        '/v1/clients/client-1/insights?status=approved'
      )
    })
  })

  // Table-driven tests for mutation methods (approve, dismiss, share)
  describe.each([
    { method: 'approve' as const, status: 'approved', endpoint: '/v1/insights/insight-1/approve' },
    { method: 'dismiss' as const, status: 'dismissed', endpoint: '/v1/insights/insight-1/dismiss' },
    { method: 'share' as const, status: 'shared', endpoint: '/v1/insights/insight-1/share' },
  ])('$method', ({ method, status, endpoint }) => {
    test(`${method}_ValidId_ReturnsUpdatedInsight`, async () => {
      const result = { ...mockInsight, status }
      mockedApiClient.put.mockResolvedValue(okResponse(result))

      const card = await repo[method]('insight-1')
      expect(card.status).toBe(status)
      expect(mockedApiClient.put).toHaveBeenCalledWith(endpoint)
    })
  })

  test('approve_ApiError_ThrowsError', async () => {
    mockedApiClient.put.mockResolvedValue(errorResponse('Invalid transition'))
    await expect(repo.approve('insight-1')).rejects.toThrow('Invalid transition')
  })

  describe('edit', () => {
    test('edit_ValidInput_ReturnsUpdatedInsight', async () => {
      const updated = { ...mockInsight, title: 'Updated Title' }
      mockedApiClient.put.mockResolvedValue(okResponse(updated))

      const result = await repo.edit('insight-1', { title: 'Updated Title' })
      expect(result.title).toBe('Updated Title')
      expect(mockedApiClient.put).toHaveBeenCalledWith('/v1/insights/insight-1', {
        title: 'Updated Title',
      })
    })
  })
})
