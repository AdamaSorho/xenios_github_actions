import { InsightQueueResponse } from '@/domain/entities/InsightCard'
import { InsightRepository } from '@/domain/repositories/InsightRepository'

export class GetInsightQueueUseCase {
  constructor(private readonly insightRepo: InsightRepository) {}

  async execute(status?: string, page?: number, limit?: number): Promise<InsightQueueResponse> {
    return this.insightRepo.getQueue(status, page, limit)
  }
}
