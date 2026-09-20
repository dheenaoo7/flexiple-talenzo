import { useState } from 'react'
import type { Profile, ScoredResult } from '../types'
import { ProfileCard } from './ProfileCard'

interface CompactProfileCardProps {
  result: ScoredResult
  profile: Profile | undefined
  defaultExpanded?: boolean
}

export function CompactProfileCard({ result, profile, defaultExpanded }: CompactProfileCardProps) {
  const [expanded, setExpanded] = useState(Boolean(defaultExpanded))

  if (expanded) {
    return (
      <div>
        <button
          type="button"
          onClick={() => setExpanded(false)}
          className="mb-2 text-xs font-medium text-indigo-600 hover:text-indigo-500"
        >
          ← Collapse
        </button>
        <ProfileCard
          result={result}
          profile={profile}
          likedCriterionIds={new Set()}
          onToggleLike={() => {}}
          reaction={null}
          onSetReaction={() => {}}
          readOnly
        />
      </div>
    )
  }

  return (
    <button
      type="button"
      onClick={() => setExpanded(true)}
      className="flex w-full items-center justify-between gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 text-left shadow-sm transition hover:border-indigo-200 hover:bg-indigo-50/40"
    >
      <div className="flex min-w-0 items-center gap-3">
        <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-indigo-100 text-xs font-semibold text-indigo-700">
          {result.rank}
        </span>
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-slate-900">{profile?.name ?? result.profile_id}</p>
          {profile && (
            <p className="truncate text-xs text-slate-500">
              {profile.current_title} · {profile.location}
            </p>
          )}
        </div>
      </div>
      <div className="flex shrink-0 items-center gap-2 text-xs text-slate-400">
        <span className="rounded-full bg-slate-100 px-2 py-0.5 font-semibold text-slate-600">
          {Math.round(result.score)}/100
        </span>
        <span>▸</span>
      </div>
    </button>
  )
}
