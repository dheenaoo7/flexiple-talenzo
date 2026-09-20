export interface Filters {
  title_keywords: string[]
  min_years_experience: number | null
  max_years_experience: number | null
  locations: string[]
  company_types: string[]
  required_skills: string[]
  nice_to_have_skills: string[]
  excluded_skills: string[]
  keywords: string
}

export interface RubricCriterion {
  id: string
  label: string
  description: string
  weight: number
}

export interface Rubric {
  criteria: RubricCriterion[]
}

export interface MatchedPoint {
  criterion_id: string
  criterion_label: string
  evidence_text: string
}

export interface ScoredResult {
  profile_id: string
  rank: number
  score: number
  summary_line: string
  matched_points: MatchedPoint[]
}

export interface EmptyResultDiagnostic {
  most_restrictive_filter: string
  explanation: string
}

export type TriggerType = 'initial' | 'manual_edit' | 'chat'

export interface Version {
  id: string
  session_id: string
  version_number: number
  trigger_type: TriggerType
  trigger_text?: string
  filters: Filters
  rubric: Rubric
  change_summary?: string
  results: ScoredResult[] | null
  empty_diagnostic?: EmptyResultDiagnostic
  created_at: string
}

export type SessionStatus = 'active' | 'frozen'

export interface Session {
  id: string
  original_query: string
  created_at: string
  status: SessionStatus
  frozen_version_number?: number
}

export interface ChatMessage {
  id: string
  session_id: string
  version_id?: string
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

export interface LikedSignal {
  id: string
  session_id: string
  version_id: string
  profile_id: string
  criterion_id: string
  criterion_label: string
  evidence_text: string
  created_at: string
}

export interface SessionDetail {
  session: Session
  versions: Version[]
  chat_messages: ChatMessage[]
  liked_signals: LikedSignal[]
}

export interface SessionSummary {
  id: string
  original_query: string
  created_at: string
  status: SessionStatus
}

export interface PastCompany {
  company: string
  company_type: string
  title: string
  years: number
}

export interface Profile {
  id: string
  name: string
  current_title: string
  years_experience: number
  location: string
  current_company: string
  current_company_type: string
  skills: string[]
  past_companies: PastCompany[]
  education: string
  summary: string
}

export interface Reaction {
  profile_id: string
  verdict: 'fit' | 'not_fit'
  note?: string
}

export interface StageEvent {
  type: 'stage'
  stage: string
  message: string
}

export interface DraftEvent {
  type: 'draft'
  payload: {
    session_id: string
    filters: Filters
    rubric: Rubric
  }
}

export interface VersionEvent {
  type: 'version'
  payload: Version
}

export interface ErrorEvent {
  type: 'error'
  stage: string
  message: string
  retryable: boolean
}

export type PipelineEvent = StageEvent | DraftEvent | VersionEvent | ErrorEvent
