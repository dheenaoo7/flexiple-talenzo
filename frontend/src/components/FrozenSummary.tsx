import type { Profile, Version } from '../types'
import { CompactProfileCard } from './CompactProfileCard'
import { FiltersRubricPanel } from './FiltersRubricPanel'
import { VersionHistoryStrip } from './VersionHistoryStrip'

interface FrozenSummaryProps {
  originalQuery: string
  versions: Version[]
  activeVersionNumber: number
  onSelectVersion: (versionNumber: number) => void
  frozenVersionNumber?: number
  profiles: Record<string, Profile>
  onStartNew: () => void
}

export function FrozenSummary({
  originalQuery,
  versions,
  activeVersionNumber,
  onSelectVersion,
  frozenVersionNumber,
  profiles,
  onStartNew,
}: FrozenSummaryProps) {
  const version = versions.find((v) => v.version_number === activeVersionNumber) ?? versions[versions.length - 1]
  const isFrozenVersion = version.version_number === frozenVersionNumber
  const results = version.results ?? []

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-indigo-200 bg-indigo-50 px-5 py-4">
        <div>
          <p className="text-xs font-medium tracking-wide text-indigo-600 uppercase">
            {isFrozenVersion ? 'Frozen shortlist' : 'Viewing an earlier version (read-only)'}
          </p>
          <p className="mt-0.5 text-sm text-indigo-900">
            "{originalQuery}" · v{version.version_number}
          </p>
        </div>
        <button
          type="button"
          onClick={onStartNew}
          className="rounded-xl bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-500"
        >
          Start new search
        </button>
      </div>

      <VersionHistoryStrip
        versions={versions}
        activeVersionNumber={version.version_number}
        onSelect={onSelectVersion}
        frozenVersionNumber={frozenVersionNumber}
        forceShow
      />

      <FiltersRubricPanel
        filters={version.filters}
        rubric={version.rubric}
        resetKey={version.id}
        mode="review"
        editable={false}
        applyLabel=""
        skillOptions={[]}
        titleOptions={[]}
        locationOptions={[]}
        companyTypeOptions={[]}
        onApply={() => {}}
      />

      <div>
        <h2 className="mb-3 text-sm font-semibold text-slate-800">Ranked shortlist</h2>
        <div className="space-y-2">
          {results.map((result) => (
            <CompactProfileCard key={result.profile_id} result={result} profile={profiles[result.profile_id]} />
          ))}
        </div>
      </div>
    </div>
  )
}
