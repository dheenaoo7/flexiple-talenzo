import type {
  Filters,
  LikedSignal,
  PipelineEvent,
  Profile,
  Reaction,
  Rubric,
  Session,
  SessionDetail,
  SessionSummary,
} from '../types'
import { streamPipeline } from './sse'

async function json<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error ?? `request failed with status ${res.status}`)
  }
  return res.json() as Promise<T>
}

export function createSession(
  query: string,
  onEvent: (event: PipelineEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  return streamPipeline('/api/sessions', { query }, onEvent, signal)
}

export function runSearch(
  sessionId: string,
  filters: Filters,
  rubric: Rubric,
  onEvent: (event: PipelineEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  return streamPipeline(`/api/sessions/${sessionId}/run`, { filters, rubric }, onEvent, signal)
}

export function evolveSession(
  sessionId: string,
  message: string,
  reactions: Reaction[],
  onEvent: (event: PipelineEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  return streamPipeline(`/api/sessions/${sessionId}/evolve`, { message, reactions }, onEvent, signal)
}

export async function getSession(sessionId: string): Promise<SessionDetail> {
  return json(await fetch(`/api/sessions/${sessionId}`))
}

export async function listSessions(): Promise<SessionSummary[]> {
  return json(await fetch('/api/sessions'))
}

export async function freezeSession(sessionId: string): Promise<Session> {
  return json(
    await fetch(`/api/sessions/${sessionId}/freeze`, { method: 'POST' }),
  )
}

export async function likeSignal(
  sessionId: string,
  signal: Omit<LikedSignal, 'id' | 'session_id' | 'created_at'>,
): Promise<LikedSignal> {
  return json(
    await fetch(`/api/sessions/${sessionId}/liked-signals`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(signal),
    }),
  )
}

export async function unlikeSignal(sessionId: string, signalId: string): Promise<void> {
  const res = await fetch(`/api/sessions/${sessionId}/liked-signals/${signalId}`, {
    method: 'DELETE',
  })
  if (!res.ok && res.status !== 204) {
    throw new Error(`failed to remove liked signal (${res.status})`)
  }
}

export async function getProfile(profileId: string): Promise<Profile> {
  return json(await fetch(`/api/profiles/${profileId}`))
}

export async function listSkills(): Promise<string[]> {
  return json(await fetch('/api/skills'))
}

export async function listTitles(): Promise<string[]> {
  return json(await fetch('/api/titles'))
}

export async function listLocations(): Promise<string[]> {
  return json(await fetch('/api/locations'))
}

export async function listCompanyTypes(): Promise<string[]> {
  return json(await fetch('/api/company-types'))
}

