import { InsightCard, InsightQueueResponse } from '@/domain/entities/InsightCard'
import { InsightRepository } from '@/domain/repositories/InsightRepository'

export const mockInsight: InsightCard = {
  id: 'insight-1',
  coachId: 'coach-1',
  clientId: 'client-1',
  title: 'Test Insight',
  body: 'Test body',
  category: 'nutrition',
  status: 'draft',
  priority: 'high',
  createdAt: '2026-02-15T10:00:00Z',
  updatedAt: '2026-02-15T10:00:00Z',
}

export const mockQueueResponse: InsightQueueResponse = {
  insights: [mockInsight],
  pagination: { page: 1, limit: 20, total: 1 },
}

export function createMockInsightRepo(): jest.Mocked<InsightRepository> {
  return {
    getQueue: jest.fn(),
    getClientInsights: jest.fn(),
    approve: jest.fn(),
    dismiss: jest.fn(),
    edit: jest.fn(),
    share: jest.fn(),
  }
}
