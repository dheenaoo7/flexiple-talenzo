import { useState } from 'react'
import type { SessionSummary } from '../types'

interface SessionHistorySidebarProps {
  sessions: SessionSummary[]
  activeSessionId?: string
  onSelect: (sessionId: string) => void
  onNew: () => void
}

export function SessionHistorySidebar({ sessions, activeSessionId, onSelect, onNew }: SessionHistorySidebarProps) {
  const [collapsed, setCollapsed] = useState(false)

  if (collapsed) {
    return (
      <div className="flex w-10 shrink-0 flex-col items-center gap-3 border-r border-slate-200 bg-white/60 py-4">
        <button
          type="button"
          onClick={() => setCollapsed(false)}
          aria-label="Expand history"
          title="Expand history"
          className="flex h-7 w-7 items-center justify-center rounded-full text-slate-400 transition hover:bg-slate-100 hover:text-slate-600"
        >
          ›
        </button>
        {sessions.length > 0 && (
          <span className="rounded-full bg-slate-100 px-1.5 py-0.5 text-[10px] font-semibold text-slate-500">
            {sessions.length}
          </span>
        )}
      </div>
    )
  }

  return (
    <div className="flex w-64 shrink-0 flex-col gap-3 border-r border-slate-200 bg-white/60 px-3 py-4">
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onNew}
          className="flex-1 rounded-xl border border-dashed border-indigo-300 px-3 py-2 text-sm font-medium text-indigo-600 transition hover:bg-indigo-50"
        >
          + New search
        </button>
        <button
          type="button"
          onClick={() => setCollapsed(true)}
          aria-label="Collapse history"
          title="Collapse history"
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-slate-400 transition hover:bg-slate-100 hover:text-slate-600"
        >
          ‹
        </button>
      </div>
      <div className="flex-1 space-y-1 overflow-y-auto">
        {sessions.length === 0 && <p className="px-2 text-xs text-slate-400">No past searches yet.</p>}
        {sessions.map((s) => (
          <button
            key={s.id}
            type="button"
            onClick={() => onSelect(s.id)}
            className={`block w-full rounded-lg px-3 py-2 text-left text-sm transition ${
              s.id === activeSessionId ? 'bg-indigo-50 text-indigo-900' : 'text-slate-600 hover:bg-slate-100'
            }`}
          >
            <div className="truncate font-medium">{s.original_query}</div>
            <div className="mt-0.5 flex items-center gap-1.5 text-xs text-slate-400">
              <span>{new Date(s.created_at).toLocaleDateString()}</span>
              {s.status === 'frozen' && <span className="text-indigo-500">❄ frozen</span>}
            </div>
          </button>
        ))}
      </div>
    </div>
  )
}
