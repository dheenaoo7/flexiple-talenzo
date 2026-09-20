import { useEffect, useRef, useState } from 'react'
import {
  createSession,
  evolveSession,
  freezeSession,
  getProfile,
  getSession,
  likeSignal,
  listCompanyTypes,
  listLocations,
  listSessions,
  listSkills,
  listTitles,
  runSearch,
  unlikeSignal,
} from './api/client'
import { ChatPanel } from './components/ChatPanel'
import { EmptyResultsState } from './components/EmptyResultsState'
import { ErrorBanner } from './components/ErrorBanner'
import { FiltersRubricPanel } from './components/FiltersRubricPanel'
import { FrozenSummary } from './components/FrozenSummary'
import { LandingSearch } from './components/LandingSearch'
import { ResultsGrid } from './components/ResultsGrid'
import { SessionHistorySidebar } from './components/SessionHistorySidebar'
import { ThinkingOverlay } from './components/ThinkingOverlay'
import { VersionHistoryStrip } from './components/VersionHistoryStrip'
import type {
  ChatMessage,
  Filters,
  LikedSignal,
  MatchedPoint,
  PipelineEvent,
  Profile,
  Reaction,
  Rubric,
  ScoredResult,
  SessionDetail,
  SessionStatus,
  SessionSummary,
  Version,
} from './types'

type BusyKind = 'draft' | 'run' | 'evolve' | 'freeze' | null

interface PipelineHandlers {
  onDraft?: (payload: { session_id: string; filters: Filters; rubric: Rubric }) => void
  onVersion?: (version: Version) => void
}

interface AppError {
  stage: string
  message: string
  retryable: boolean
}

function App() {
  const [sessions, setSessions] = useState<SessionSummary[]>([])
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [originalQuery, setOriginalQuery] = useState('')
  const [sessionStatus, setSessionStatus] = useState<SessionStatus>('active')
  const [frozenVersionNumber, setFrozenVersionNumber] = useState<number | undefined>(undefined)
  const [versions, setVersions] = useState<Version[]>([])
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([])
  const [likedSignals, setLikedSignals] = useState<LikedSignal[]>([])
  const [activeVersionNumber, setActiveVersionNumber] = useState(0)
  const [draftFilters, setDraftFilters] = useState<Filters | null>(null)
  const [draftRubric, setDraftRubric] = useState<Rubric | null>(null)
  const [profiles, setProfiles] = useState<Record<string, Profile>>({})
  const [pendingReactions, setPendingReactions] = useState<Record<string, 'fit' | 'not_fit'>>({})
  const [busyKind, setBusyKind] = useState<BusyKind>(null)
  const [stages, setStages] = useState<string[]>([])
  const [error, setError] = useState<AppError | null>(null)
  const [chatFocused, setChatFocused] = useState(false)
  const [skillOptions, setSkillOptions] = useState<string[]>([])
  const [titleOptions, setTitleOptions] = useState<string[]>([])
  const [locationOptions, setLocationOptions] = useState<string[]>([])
  const [companyTypeOptions, setCompanyTypeOptions] = useState<string[]>([])

  const lastActionRef = useRef<(() => void) | null>(null)
  const fetchedProfileIds = useRef(new Set<string>())
  const filtersRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    listSessions().then(setSessions).catch(() => {})
    listSkills().then(setSkillOptions).catch(() => {})
    listTitles().then(setTitleOptions).catch(() => {})
    listLocations().then(setLocationOptions).catch(() => {})
    listCompanyTypes().then(setCompanyTypeOptions).catch(() => {})
  }, [])

  useEffect(() => {
    const ids = Array.from(new Set(versions.flatMap((v) => (v.results ?? []).map((r) => r.profile_id))))
    const toFetch = ids.filter((id) => !fetchedProfileIds.current.has(id))
    if (toFetch.length === 0) return
    toFetch.forEach((id) => fetchedProfileIds.current.add(id))
    Promise.all(toFetch.map((id) => getProfile(id).catch(() => null))).then((fetched) => {
      setProfiles((prev) => {
        const next = { ...prev }
        toFetch.forEach((id, i) => {
          const p = fetched[i]
          if (p) next[id] = p
        })
        return next
      })
    })
  }, [versions])

  function applySessionDetail(detail: SessionDetail) {
    setSessionStatus(detail.session.status)
    setFrozenVersionNumber(detail.session.frozen_version_number)
    setVersions(detail.versions)
    setChatMessages(detail.chat_messages)
    setLikedSignals(detail.liked_signals)
  }

  function runPipeline(kind: Exclude<BusyKind, null>, call: (onEvent: (e: PipelineEvent) => void) => Promise<void>, handlers: PipelineHandlers) {
    const action = () => {
      setError(null)
      setBusyKind(kind)
      setStages([])
      call((event) => {
        if (event.type === 'stage') {
          setStages((s) => [...s, event.message])
        } else if (event.type === 'draft') {
          handlers.onDraft?.(event.payload)
        } else if (event.type === 'version') {
          handlers.onVersion?.(event.payload)
        } else if (event.type === 'error') {
          setError({ stage: event.stage, message: event.message, retryable: event.retryable })
          setBusyKind(null)
        }
      }).catch((err) => {
        setError({ stage: kind, message: err instanceof Error ? err.message : String(err), retryable: true })
        setBusyKind(null)
      })
    }
    lastActionRef.current = action
    action()
  }

  function resetToLanding() {
    setSessionId(null)
    setOriginalQuery('')
    setSessionStatus('active')
    setFrozenVersionNumber(undefined)
    setVersions([])
    setChatMessages([])
    setLikedSignals([])
    setActiveVersionNumber(0)
    setDraftFilters(null)
    setDraftRubric(null)
    setPendingReactions({})
    setError(null)
    setStages([])
    setBusyKind(null)
  }

  function handleLandingSubmit(query: string) {
    runPipeline('draft', (onEvent) => createSession(query, onEvent), {
      onDraft: (payload) => {
        setSessionId(payload.session_id)
        setOriginalQuery(query)
        setDraftFilters(payload.filters)
        setDraftRubric(payload.rubric)
        setBusyKind(null)
        listSessions().then(setSessions).catch(() => {})
      },
    })
  }

  function runSearchAction(filters: Filters, rubric: Rubric) {
    if (!sessionId) return
    const id = sessionId
    runPipeline('run', (onEvent) => runSearch(id, filters, rubric, onEvent), {
      onVersion: (version) => {
        setBusyKind(null)
        setDraftFilters(null)
        setDraftRubric(null)
        setActiveVersionNumber(version.version_number)
        getSession(id)
          .then(applySessionDetail)
          .catch(() => setVersions((v) => [...v, version]))
        listSessions().then(setSessions).catch(() => {})
      },
    })
  }

  function evolveAction(message: string) {
    if (!sessionId) return
    const id = sessionId
    const reactions: Reaction[] = Object.entries(pendingReactions).map(([profile_id, verdict]) => ({
      profile_id,
      verdict,
    }))
    runPipeline('evolve', (onEvent) => evolveSession(id, message, reactions, onEvent), {
      onVersion: (version) => {
        setBusyKind(null)
        setActiveVersionNumber(version.version_number)
        setPendingReactions({})
        getSession(id)
          .then(applySessionDetail)
          .catch(() => setVersions((v) => [...v, version]))
      },
    })
  }

  async function freeze() {
    if (!sessionId) return
    setBusyKind('freeze')
    try {
      const updated = await freezeSession(sessionId)
      setSessionStatus(updated.status)
      setFrozenVersionNumber(updated.frozen_version_number)
      if (updated.frozen_version_number !== undefined) setActiveVersionNumber(updated.frozen_version_number)
      listSessions().then(setSessions).catch(() => {})
    } catch (err) {
      setError({
        stage: 'freezing the search',
        message: err instanceof Error ? err.message : String(err),
        retryable: true,
      })
    } finally {
      setBusyKind(null)
    }
  }

  async function toggleLike(result: ScoredResult, point: MatchedPoint) {
    if (!sessionId) return
    const activeVersion = versions.find((v) => v.version_number === activeVersionNumber)
    if (!activeVersion) return
    const existing = likedSignals.find(
      (s) => s.profile_id === result.profile_id && s.criterion_id === point.criterion_id,
    )
    if (existing) {
      setLikedSignals((s) => s.filter((x) => x.id !== existing.id))
      try {
        await unlikeSignal(sessionId, existing.id)
      } catch {
        setLikedSignals((s) => [...s, existing])
      }
      return
    }
    try {
      const created = await likeSignal(sessionId, {
        version_id: activeVersion.id,
        profile_id: result.profile_id,
        criterion_id: point.criterion_id,
        criterion_label: point.criterion_label,
        evidence_text: point.evidence_text,
      })
      setLikedSignals((s) => [...s, created])
    } catch (err) {
      setError({
        stage: 'saving your like',
        message: err instanceof Error ? err.message : String(err),
        retryable: false,
      })
    }
  }

  async function openSession(id: string) {
    setError(null)
    try {
      const detail = await getSession(id)
      setSessionId(id)
      setOriginalQuery(detail.session.original_query)
      setDraftFilters(null)
      setDraftRubric(null)
      setPendingReactions({})
      applySessionDetail(detail)
      const latest = detail.versions.length ? Math.max(...detail.versions.map((v) => v.version_number)) : 0
      setActiveVersionNumber(detail.session.frozen_version_number ?? latest)
    } catch (err) {
      setError({
        stage: 'loading that session',
        message: err instanceof Error ? err.message : String(err),
        retryable: false,
      })
    }
  }

  function handleRetry() {
    const action = lastActionRef.current
    setError(null)
    action?.()
  }

  function setReaction(profileId: string, verdict: 'fit' | 'not_fit' | null) {
    setPendingReactions((prev) => {
      const next = { ...prev }
      if (verdict) next[profileId] = verdict
      else delete next[profileId]
      return next
    })
  }


  let mainContent: React.ReactNode

  if (!sessionId) {
    mainContent = busyKind === 'draft' ? <ThinkingOverlay stages={stages} /> : <LandingSearch onSubmit={handleLandingSubmit} />
  } else if (sessionStatus === 'frozen') {
    mainContent =
      versions.length > 0 ? (
        <FrozenSummary
          originalQuery={originalQuery}
          versions={versions}
          activeVersionNumber={activeVersionNumber}
          onSelectVersion={setActiveVersionNumber}
          frozenVersionNumber={frozenVersionNumber}
          profiles={profiles}
          onStartNew={resetToLanding}
        />
      ) : (
        <ThinkingOverlay stages={['Loading…']} />
      )
  } else if (versions.length === 0) {
    mainContent =
      busyKind === 'draft' || !draftFilters || !draftRubric ? (
        <ThinkingOverlay stages={stages.length ? stages : ['Thinking…']} />
      ) : (
        <div className="space-y-5">
          <FiltersRubricPanel
            filters={draftFilters}
            rubric={draftRubric}
            resetKey={sessionId}
            mode="review"
            editable
            busy={busyKind === 'run'}
            applyLabel="Run search →"
            skillOptions={skillOptions}
            titleOptions={titleOptions}
            locationOptions={locationOptions}
            companyTypeOptions={companyTypeOptions}
            onApply={(f, r) => runSearchAction(f, r)}
          />
          {busyKind === 'run' && <ThinkingOverlay stages={stages} variant="inline" />}
        </div>
      )
  } else {
    const latestVersionNumber = Math.max(...versions.map((v) => v.version_number))
    const activeVersion = versions.find((v) => v.version_number === activeVersionNumber) ?? versions[versions.length - 1]
    const isLatest = activeVersion.version_number === latestVersionNumber
    const results = activeVersion.results ?? []
    const pendingReactionList: Reaction[] = Object.entries(pendingReactions).map(([profile_id, verdict]) => ({
      profile_id,
      verdict,
    }))

    mainContent = (
      <div className="space-y-5">
        <div className="flex items-center justify-between gap-3">
          <VersionHistoryStrip
            versions={versions}
            activeVersionNumber={activeVersion.version_number}
            onSelect={setActiveVersionNumber}
          />
          <button
            type="button"
            onClick={freeze}
            disabled={busyKind === 'freeze'}
            className="shrink-0 rounded-xl border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 transition hover:bg-slate-100 disabled:opacity-50"
          >
            {busyKind === 'freeze' ? 'Freezing…' : 'Freeze shortlist'}
          </button>
        </div>

        {!chatFocused && (
          <div ref={filtersRef}>
            <FiltersRubricPanel
              filters={activeVersion.filters}
              rubric={activeVersion.rubric}
              resetKey={activeVersion.id}
              mode="header"
              editable={isLatest}
              busy={busyKind === 'run'}
              applyLabel="Re-run with edits"
              skillOptions={skillOptions}
              titleOptions={titleOptions}
              locationOptions={locationOptions}
              companyTypeOptions={companyTypeOptions}
              onApply={(f, r) => runSearchAction(f, r)}
            />
          </div>
        )}

        {busyKind === 'run' && <ThinkingOverlay stages={stages} variant="inline" />}

        {results.length === 0 ? (
          <EmptyResultsState
            diagnostic={activeVersion.empty_diagnostic}
            onEditFilters={() => filtersRef.current?.scrollIntoView({ behavior: 'smooth' })}
          />
        ) : (
          <ResultsGrid
            results={results}
            profiles={profiles}
            likedSignals={likedSignals}
            onToggleLike={toggleLike}
            reactions={pendingReactions}
            onSetReaction={setReaction}
            readOnly={!isLatest}
          />
        )}

        {isLatest && (
          <ChatPanel
            messages={chatMessages}
            pendingReactions={pendingReactionList}
            onRemovePendingReaction={(profileId) => setReaction(profileId, null)}
            onSend={evolveAction}
            busy={busyKind === 'evolve'}
            onFocusChange={setChatFocused}
          />
        )}
      </div>
    )
  }

  return (
    <div className="flex h-screen overflow-hidden">
      {!chatFocused && (
        <SessionHistorySidebar
          sessions={sessions}
          activeSessionId={sessionId ?? undefined}
          onSelect={openSession}
          onNew={resetToLanding}
        />
      )}
      <div className="flex-1 overflow-y-auto">
        <div className="mx-auto max-w-5xl px-6 py-8">
          {error && (
            <div className="mb-5">
              <ErrorBanner
                stage={error.stage}
                message={error.message}
                retryable={error.retryable}
                onRetry={handleRetry}
                onDismiss={() => setError(null)}
              />
            </div>
          )}
          {mainContent}
        </div>
      </div>
    </div>
  )
}

export default App
