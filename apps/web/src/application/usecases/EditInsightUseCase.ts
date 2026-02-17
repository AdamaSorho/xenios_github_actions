import { InsightCard, EditInsightInput } from '@/domain/entities/InsightCard'
import { InsightRepository } from '@/domain/repositories/InsightRepository'

export class EditInsightUseCase {
  constructor(private readonly insightRepo: InsightRepository) {}

  async execute(insightId: string, input: EditInsightInput): Promise<InsightCard> {
    if (!insightId) {
      throw new Error('Insight ID is required')
    }
    if (!input.title) {
      throw new Error('Title is required')
    }
    if (!input.body) {
      throw new Error('Body is required')
    }
    return this.insightRepo.edit(insightId, input)
  }
}
