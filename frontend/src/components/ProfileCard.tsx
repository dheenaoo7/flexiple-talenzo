import type { MatchedPoint, Profile, ScoredResult } from '../types'

interface ProfileCardProps {
  result: ScoredResult
  profile: Profile | undefined
  likedCriterionIds: Set<string>
  onToggleLike: (point: MatchedPoint) => void
  reaction: 'fit' | 'not_fit' | null
  onSetReaction: (verdict: 'fit' | 'not_fit' | null) => void
  readOnly?: boolean
}

export function ProfileCard({
  result,
  profile,
  likedCriterionIds,
  onToggleLike,
  reaction,
  onSetReaction,
  readOnly,
}: ProfileCardProps) {
  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-indigo-100 text-xs font-semibold text-indigo-700">
              {result.rank}
            </span>
            <h3 className="text-base font-semibold text-slate-900">{profile?.name ?? result.profile_id}</h3>
          </div>
          {profile && (
            <p className="mt-0.5 text-sm text-slate-500">
              {profile.current_title} · {profile.current_company} · {profile.years_experience} yrs · {profile.location}
            </p>
          )}
        </div>
        <div className="shrink-0 rounded-full bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-600">
          {Math.round(result.score)}/100
        </div>
      </div>

      <p className="text-sm text-slate-700">{result.summary_line}</p>

      {profile && profile.skills.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {profile.skills.map((s) => (
            <span key={s} className="rounded-full bg-slate-50 px-2 py-0.5 text-xs text-slate-500 ring-1 ring-inset ring-slate-200">
              {s}
            </span>
          ))}
        </div>
      )}

      {result.matched_points.length > 0 && (
        <ul className="space-y-1.5 border-t border-slate-100 pt-3">
          {result.matched_points.map((mp) => {
            const liked = likedCriterionIds.has(mp.criterion_id)
            return (
              <li key={mp.criterion_id} className="flex items-start gap-2 text-sm">
                <button
                  type="button"
                  disabled={readOnly}
                  onClick={() => onToggleLike(mp)}
                  aria-label={liked ? 'Unlike this match' : 'Like this match'}
                  className={`mt-0.5 shrink-0 text-base leading-none transition ${
                    liked ? 'text-indigo-600' : 'text-slate-300 hover:text-slate-400'
                  } ${readOnly ? 'cursor-default' : 'cursor-pointer'}`}
                >
                  {liked ? '★' : '☆'}
                </button>
                <span className="text-slate-600">
                  <span className="font-medium text-slate-800">{mp.criterion_label}:</span> {mp.evidence_text}
                </span>
              </li>
            )
          })}
        </ul>
      )}

      {!readOnly && (
        <div className="mt-1 flex gap-2 border-t border-slate-100 pt-3">
          <button
            type="button"
            onClick={() => onSetReaction(reaction === 'fit' ? null : 'fit')}
            className={`flex-1 rounded-lg px-3 py-1.5 text-xs font-medium transition ${
              reaction === 'fit'
                ? 'bg-emerald-600 text-white'
                : 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100'
            }`}
          >
            Good fit
          </button>
          <button
            type="button"
            onClick={() => onSetReaction(reaction === 'not_fit' ? null : 'not_fit')}
            className={`flex-1 rounded-lg px-3 py-1.5 text-xs font-medium transition ${
              reaction === 'not_fit' ? 'bg-rose-600 text-white' : 'bg-rose-50 text-rose-700 hover:bg-rose-100'
            }`}
          >
            Not a fit
          </button>
        </div>
      )}
    </div>
  )
}
