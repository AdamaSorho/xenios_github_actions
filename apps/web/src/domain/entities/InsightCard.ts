export type InsightStatus = 'draft' | 'approved' | 'dismissed' | 'shared'
export type InsightCategory = 'general' | 'nutrition' | 'recovery' | 'performance' | 'behavior' | 'safety'
export type InsightPriority = 'low' | 'medium' | 'high' | 'urgent'

export interface InsightEvidence {
  measurementId: string
  description: string
}

export interface InsightCard {
  id: string
  coachId: string
  clientId: string
  clientName?: string
  sessionId?: string
  title: string
  body: string
  category: InsightCategory
  status: InsightStatus
  priority: InsightPriority
  evidence?: InsightEvidence[]
  approvedAt?: string
  sharedAt?: string
  createdAt: string
  updatedAt: string
}

export interface InsightQueueResponse {
  insights: InsightCard[]
  page: number
  limit: number
  total: number
}

export interface EditInsightInput {
  title: string
  body: string
}
