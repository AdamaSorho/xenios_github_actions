import { InsightCard, InsightQueueResponse, EditInsightInput } from '@/domain/entities/InsightCard'

export interface InsightRepository {
  getQueue(status?: string, page?: number, limit?: number): Promise<InsightQueueResponse>
  getClientInsights(clientId: string, status?: string, page?: number, limit?: number): Promise<InsightQueueResponse>
  approve(insightId: string): Promise<InsightCard>
  dismiss(insightId: string): Promise<InsightCard>
  edit(insightId: string, input: EditInsightInput): Promise<InsightCard>
  share(insightId: string): Promise<InsightCard>
}
