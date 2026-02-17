import { InsightCard, InsightQueueResponse, EditInsightInput } from '../entities/InsightCard'

/**
 * InsightRepository interface - defines data access operations for insights.
 *
 * NOTE: This is an INTERFACE only - no API client imports here!
 * Implementations live in the infrastructure layer.
 */
export interface InsightRepository {
  getQueue(status?: string, page?: number, limit?: number): Promise<InsightQueueResponse>
  getClientInsights(clientId: string, status?: string, limit?: number, offset?: number): Promise<InsightQueueResponse>
  approve(insightId: string): Promise<InsightCard>
  dismiss(insightId: string): Promise<InsightCard>
  edit(insightId: string, input: EditInsightInput): Promise<InsightCard>
  share(insightId: string): Promise<InsightCard>
}
