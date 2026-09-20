import { useState } from 'react'
import type { ChatMessage, Reaction } from '../types'

interface ChatPanelProps {
  messages: ChatMessage[]
  pendingReactions: Reaction[]
  onRemovePendingReaction: (profileId: string) => void
  onSend: (message: string) => void
  busy?: boolean
  disabled?: boolean
  onFocusChange?: (focused: boolean) => void
}

export function ChatPanel({
  messages,
  pendingReactions,
  onRemovePendingReaction,
  onSend,
  busy,
  disabled,
  onFocusChange,
}: ChatPanelProps) {
  const [draft, setDraft] = useState('')

  function submit() {
    const trimmed = draft.trim()
    if (!trimmed && pendingReactions.length === 0) return
    onSend(trimmed)
    setDraft('')
  }

  return (
    <div className="flex flex-col rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div className="border-b border-slate-100 px-5 py-3">
        <h3 className="text-sm font-semibold text-slate-800">Feedback</h3>
        <p className="text-xs text-slate-500">Tell us what's wrong or right, and we'll update the search.</p>
      </div>

      <div className="flex max-h-80 flex-col gap-3 overflow-y-auto px-5 py-4">
        {messages.length === 0 && (
          <p className="text-sm text-slate-400">No feedback yet — mark profiles as fit/not-fit or type below.</p>
        )}
        {messages.map((m) => (
          <div key={m.id} className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div
              className={`max-w-[85%] rounded-2xl px-3.5 py-2 text-sm ${
                m.role === 'user' ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-700'
              }`}
            >
              {m.content}
            </div>
          </div>
        ))}
        {busy && (
          <div className="flex justify-start">
            <div className="max-w-[85%] rounded-2xl bg-slate-100 px-3.5 py-2 text-sm text-slate-400">Thinking…</div>
          </div>
        )}
      </div>

      {!disabled && (
        <div className="border-t border-slate-100 px-5 py-3">
          {pendingReactions.length > 0 && (
            <div className="mb-2 flex flex-wrap gap-1.5">
              {pendingReactions.map((r) => (
                <span
                  key={r.profile_id}
                  className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs ring-1 ring-inset ${
                    r.verdict === 'fit' ? 'bg-emerald-50 text-emerald-700 ring-emerald-200' : 'bg-rose-50 text-rose-700 ring-rose-200'
                  }`}
                >
                  {r.profile_id}: {r.verdict === 'fit' ? 'fit' : 'not fit'}
                  <button
                    type="button"
                    onClick={() => onRemovePendingReaction(r.profile_id)}
                    className="text-current/60 hover:text-current"
                    aria-label={`Remove reaction for ${r.profile_id}`}
                  >
                    ×
                  </button>
                </span>
              ))}
            </div>
          )}
          <form
            onSubmit={(e) => {
              e.preventDefault()
              submit()
            }}
            className="flex items-center gap-2 rounded-xl border border-slate-200 bg-slate-50 p-1.5 focus-within:border-indigo-300 focus-within:ring-4 focus-within:ring-indigo-50"
          >
            <input
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onFocus={() => onFocusChange?.(true)}
              onBlur={() => onFocusChange?.(false)}
              disabled={busy}
              placeholder="e.g. 1 is too junior, 2 and 4 are right"
              className="flex-1 border-none bg-transparent px-2 py-1.5 text-sm outline-none placeholder:text-slate-400 disabled:opacity-60"
            />
            <button
              type="submit"
              disabled={busy || (!draft.trim() && pendingReactions.length === 0)}
              className="rounded-lg bg-indigo-600 px-4 py-1.5 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-40"
            >
              Send
            </button>
          </form>
        </div>
      )}
    </div>
  )
}
