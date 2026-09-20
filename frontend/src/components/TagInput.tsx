import { useState } from 'react'

interface TagInputProps {
  label: string
  values: string[]
  onChange: (values: string[]) => void
  placeholder?: string
  tone?: 'default' | 'danger'
  disabled?: boolean
}

const toneClasses: Record<NonNullable<TagInputProps['tone']>, string> = {
  default: 'bg-indigo-50 text-indigo-700 ring-indigo-200',
  danger: 'bg-rose-50 text-rose-700 ring-rose-200',
}

export function TagInput({ label, values, onChange, placeholder, tone = 'default', disabled }: TagInputProps) {
  const [draft, setDraft] = useState('')

  function commit() {
    const v = draft.trim()
    if (v && !values.includes(v)) onChange([...values, v])
    setDraft('')
  }

  function remove(v: string) {
    onChange(values.filter((x) => x !== v))
  }

  return (
    <div>
      <div className="mb-1.5 text-xs font-medium tracking-wide text-slate-500 uppercase">{label}</div>
      <div className="flex flex-wrap items-center gap-1.5">
        {values.map((v) => (
          <span
            key={v}
            className={`inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-sm ring-1 ring-inset ${toneClasses[tone]}`}
          >
            {v}
            {!disabled && (
              <button
                type="button"
                onClick={() => remove(v)}
                className="rounded-full text-current/60 hover:text-current"
                aria-label={`Remove ${v}`}
              >
                ×
              </button>
            )}
          </span>
        ))}
        {!disabled && (
          <input
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ',') {
                e.preventDefault()
                commit()
              } else if (e.key === 'Backspace' && draft === '' && values.length > 0) {
                remove(values[values.length - 1])
              }
            }}
            onBlur={commit}
            placeholder={placeholder ?? 'Add…'}
            className="min-w-[8ch] flex-1 border-none bg-transparent px-1 py-1 text-sm text-slate-700 outline-none placeholder:text-slate-400"
          />
        )}
        {values.length === 0 && disabled && <span className="text-sm text-slate-400">None</span>}
      </div>
    </div>
  )
}
