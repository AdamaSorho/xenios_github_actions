import { InsightQueueResponse } from '@/domain/entities/InsightCard'
import { InsightRepository } from '@/domain/repositories/InsightRepository'

export class GetClientInsightsUseCase {
  constructor(private readonly insightRepo: InsightRepository) {}

  async execute(clientId: string, status?: string, limit?: number, offset?: number): Promise<InsightQueueResponse> {
    if (!clientId) {
      throw new Error('Client ID is required')
    }
    return this.insightRepo.getClientInsights(clientId, status, limit, offset)
  }
}
