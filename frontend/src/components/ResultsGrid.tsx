import type { LikedSignal, MatchedPoint, Profile, ScoredResult } from '../types'
import { ProfileCard } from './ProfileCard'

interface ResultsGridProps {
  results: ScoredResult[]
  profiles: Record<string, Profile>
  likedSignals: LikedSignal[]
  onToggleLike: (result: ScoredResult, point: MatchedPoint) => void
  reactions: Record<string, 'fit' | 'not_fit'>
  onSetReaction: (profileId: string, verdict: 'fit' | 'not_fit' | null) => void
  readOnly?: boolean
}

export function ResultsGrid({
  results,
  profiles,
  likedSignals,
  onToggleLike,
  reactions,
  onSetReaction,
  readOnly,
}: ResultsGridProps) {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
      {results.map((result) => {
        const likedCriterionIds = new Set(
          likedSignals.filter((s) => s.profile_id === result.profile_id).map((s) => s.criterion_id),
        )
        return (
          <ProfileCard
            key={result.profile_id}
            result={result}
            profile={profiles[result.profile_id]}
            likedCriterionIds={likedCriterionIds}
            onToggleLike={(point) => onToggleLike(result, point)}
            reaction={reactions[result.profile_id] ?? null}
            onSetReaction={(verdict) => onSetReaction(result.profile_id, verdict)}
            readOnly={readOnly}
          />
        )
      })}
    </div>
  )
}
