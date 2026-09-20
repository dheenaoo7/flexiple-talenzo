import type { Version } from '../types'

interface VersionHistoryStripProps {
  versions: Version[]
  activeVersionNumber: number
  onSelect: (versionNumber: number) => void
  frozenVersionNumber?: number
  forceShow?: boolean
}

const triggerLabel: Record<Version['trigger_type'], string> = {
  initial: 'Initial',
  manual_edit: 'Manual edit',
  chat: 'Feedback',
}

export function VersionHistoryStrip({
  versions,
  activeVersionNumber,
  onSelect,
  frozenVersionNumber,
  forceShow,
}: VersionHistoryStripProps) {
  if (versions.length <= 1 && !forceShow) return null

  const sorted = [...versions].sort((a, b) => a.version_number - b.version_number)

  return (
    <div className="flex flex-wrap items-center gap-2 rounded-2xl border border-slate-200 bg-white px-4 py-3 shadow-sm">
      <span className="text-xs font-medium tracking-wide text-slate-500 uppercase">Versions</span>
      {sorted.map((v) => {
        const active = v.version_number === activeVersionNumber
        const frozen = v.version_number === frozenVersionNumber
        return (
          <button
            key={v.id}
            type="button"
            onClick={() => onSelect(v.version_number)}
            title={triggerLabel[v.trigger_type]}
            className={`flex items-center gap-1 rounded-full px-3 py-1 text-xs font-medium transition ${
              active ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
            }`}
          >
            v{v.version_number}
            {frozen && <span title="Frozen">❄</span>}
          </button>
        )
      })}
    </div>
  )
}
