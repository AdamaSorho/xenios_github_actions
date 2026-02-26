import { InsightCard, InsightQueueResponse, InsightStatus, EditInsightInput } from '@/domain/entities/InsightCard'
import { InsightRepository } from '@/domain/repositories/InsightRepository'
import { apiClient } from '@xenios/api-client'

/**
 * Validates an API response and returns its data, or throws on error.
 */
function unwrap<T>(response: { ok: boolean; data: T | null; error?: string | null }, fallbackMsg: string): T {
  if (!response.ok || !response.data) {
    throw new Error(response.error || fallbackMsg)
  }
  return response.data
}

/**
 * Builds a query string from optional pagination parameters.
 */
function buildPaginationQuery(status?: string, limit?: number, offset?: number): string {
  const params = new URLSearchParams()
  if (status) params.set('status', status)
  if (limit) params.set('limit', String(limit))
  if (offset) params.set('offset', String(offset))
  const query = params.toString()
  return query ? `?${query}` : ''
}

/**
 * ApiInsightRepository - implementation of InsightRepository using Backend API.
 *
 * IMPORTANT: Web NEVER accesses the database directly.
 * All insight operations go through the Backend API.
 */
export class ApiInsightRepository implements InsightRepository {
  async getQueue(status?: InsightStatus, limit?: number, offset?: number): Promise<InsightQueueResponse> {
    const url = `/v1/insights/queue${buildPaginationQuery(status, limit, offset)}`
    return unwrap(await apiClient.get<InsightQueueResponse>(url), 'Failed to fetch insight queue')
  }

  async getClientInsights(clientId: string, status?: InsightStatus, limit?: number, offset?: number): Promise<InsightQueueResponse> {
    const url = `/v1/clients/${clientId}/insights${buildPaginationQuery(status, limit, offset)}`
    return unwrap(await apiClient.get<InsightQueueResponse>(url), 'Failed to fetch client insights')
  }

  async approve(insightId: string): Promise<InsightCard> {
    return unwrap(await apiClient.put<InsightCard>(`/v1/insights/${insightId}/approve`), 'Failed to approve insight')
  }

  async dismiss(insightId: string): Promise<InsightCard> {
    return unwrap(await apiClient.put<InsightCard>(`/v1/insights/${insightId}/dismiss`), 'Failed to dismiss insight')
  }

  async edit(insightId: string, input: EditInsightInput): Promise<InsightCard> {
    return unwrap(await apiClient.put<InsightCard>(`/v1/insights/${insightId}`, input), 'Failed to edit insight')
  }

  async share(insightId: string): Promise<InsightCard> {
    return unwrap(await apiClient.put<InsightCard>(`/v1/insights/${insightId}/share`), 'Failed to share insight')
  }
}
