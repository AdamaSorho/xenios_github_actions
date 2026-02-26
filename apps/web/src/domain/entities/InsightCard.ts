export type InsightStatus = 'draft' | 'approved' | 'dismissed' | 'shared'
export type InsightPriority = 'urgent' | 'high' | 'medium' | 'low'
export type InsightCategory = 'nutrition' | 'exercise' | 'sleep' | 'stress' | 'general'

export interface InsightEvidence {
  measurement_id: string
  description: string
}

export interface InsightCard {
  id: string
  client_id: string
  coach_id: string
  client_name: string
  title: string
  body: string
  category: InsightCategory
  priority: InsightPriority
  status: InsightStatus
  evidence: InsightEvidence[]
  created_at: string
  updated_at: string
  approved_at: string | null
  dismissed_at: string | null
  shared_at: string | null
}

export interface InsightQueueResponse {
  insights: InsightCard[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface EditInsightInput {
  title?: string
  body?: string
}
