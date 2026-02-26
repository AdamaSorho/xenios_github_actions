import { InsightMutationUseCase } from './InsightMutationUseCase'
import { InsightRepository } from '@/domain/repositories/InsightRepository'

export class ApproveInsightUseCase extends InsightMutationUseCase {
  constructor(insightRepo: InsightRepository) {
    super(insightRepo, 'approve')
  }
}
