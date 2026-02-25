import { InsightCard, InsightQueueResponse, InsightStatus, EditInsightInput } from '@/domain/entities/InsightCard'
import { InsightRepository } from '@/domain/repositories/InsightRepository'
import { apiClient } from '@xenios/api-client'

/**
 * ApiInsightRepository - implementation of InsightRepository using Backend API.
 *
 * IMPORTANT: Web NEVER accesses the database directly.
 * All insight operations go through the Backend API.
 */
export class ApiInsightRepository implements InsightRepository {
  async getQueue(status?: InsightStatus, limit?: number, offset?: number): Promise<InsightQueueResponse> {
    const params = new URLSearchParams()
    if (status) params.set('status', status)
    if (limit) params.set('limit', String(limit))
    if (offset) params.set('offset', String(offset))

    const query = params.toString()
    const url = `/v1/insights/queue${query ? `?${query}` : ''}`

    const response = await apiClient.get<InsightQueueResponse>(url)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Failed to fetch insight queue')
    }
    return response.data
  }

  async getClientInsights(clientId: string, status?: InsightStatus, limit?: number, offset?: number): Promise<InsightQueueResponse> {
    const params = new URLSearchParams()
    if (status) params.set('status', status)
    if (limit) params.set('limit', String(limit))
    if (offset) params.set('offset', String(offset))

    const query = params.toString()
    const url = `/v1/clients/${clientId}/insights${query ? `?${query}` : ''}`

    const response = await apiClient.get<InsightQueueResponse>(url)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Failed to fetch client insights')
    }
    return response.data
  }

  async approve(insightId: string): Promise<InsightCard> {
    const response = await apiClient.put<InsightCard>(`/v1/insights/${insightId}/approve`)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Failed to approve insight')
    }
    return response.data
  }

  async dismiss(insightId: string): Promise<InsightCard> {
    const response = await apiClient.put<InsightCard>(`/v1/insights/${insightId}/dismiss`)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Failed to dismiss insight')
    }
    return response.data
  }

  async edit(insightId: string, input: EditInsightInput): Promise<InsightCard> {
    const response = await apiClient.put<InsightCard>(`/v1/insights/${insightId}`, input)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Failed to edit insight')
    }
    return response.data
  }

  async share(insightId: string): Promise<InsightCard> {
    const response = await apiClient.put<InsightCard>(`/v1/insights/${insightId}/share`)
    if (!response.ok || !response.data) {
      throw new Error(response.error || 'Failed to share insight')
    }
    return response.data
  }
}
