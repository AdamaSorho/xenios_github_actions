import { InsightCard } from '@/domain/entities/InsightCard'
import { InsightRepository } from '@/domain/repositories/InsightRepository'

export class ShareInsightUseCase {
  constructor(private readonly insightRepo: InsightRepository) {}

  async execute(insightId: string): Promise<InsightCard> {
    if (!insightId) {
      throw new Error('insight_id is required')
    }
    return this.insightRepo.share(insightId)
  }
}
