import { InsightCard, InsightQueueResponse, InsightStatus, EditInsightInput } from '../entities/InsightCard'

/**
 * InsightRepository interface - defines data access operations for insight cards.
 *
 * NOTE: This is an INTERFACE only - no API client imports here!
 * Implementations live in the infrastructure layer.
 */
export interface InsightRepository {
  getQueue(status?: InsightStatus, limit?: number, offset?: number): Promise<InsightQueueResponse>
  getClientInsights(clientId: string, status?: InsightStatus, limit?: number, offset?: number): Promise<InsightQueueResponse>
  approve(insightId: string): Promise<InsightCard>
  dismiss(insightId: string): Promise<InsightCard>
  edit(insightId: string, input: EditInsightInput): Promise<InsightCard>
  share(insightId: string): Promise<InsightCard>
}
