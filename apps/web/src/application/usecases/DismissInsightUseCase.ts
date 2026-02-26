import { InsightMutationUseCase } from './InsightMutationUseCase'
import { InsightRepository } from '@/domain/repositories/InsightRepository'

export class DismissInsightUseCase extends InsightMutationUseCase {
  constructor(insightRepo: InsightRepository) {
    super(insightRepo, 'dismiss')
  }
}
