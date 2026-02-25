import { InsightCard, EditInsightInput } from '@/domain/entities/InsightCard'
import { InsightRepository } from '@/domain/repositories/InsightRepository'

export class EditInsightUseCase {
  constructor(private readonly insightRepo: InsightRepository) {}

  async execute(insightId: string, input: EditInsightInput): Promise<InsightCard> {
    if (!insightId) {
      throw new Error('insight_id is required')
    }
    if (!input.title && !input.body) {
      throw new Error('title or body is required')
    }
    return this.insightRepo.edit(insightId, input)
  }
}
