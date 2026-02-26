import { InsightCard } from '@/domain/entities/InsightCard'
import { InsightRepository } from '@/domain/repositories/InsightRepository'

type InsightMutationMethod = 'approve' | 'dismiss' | 'share'

/**
 * InsightMutationUseCase handles single-insight mutation operations
 * (approve, dismiss, share) that share the same validate-and-delegate pattern.
 */
export class InsightMutationUseCase {
  constructor(
    private readonly insightRepo: InsightRepository,
    private readonly method: InsightMutationMethod
  ) {}

  async execute(insightId: string): Promise<InsightCard> {
    if (!insightId) {
      throw new Error('insight_id is required')
    }
    return this.insightRepo[this.method](insightId)
  }
}
