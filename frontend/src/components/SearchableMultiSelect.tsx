import { useEffect, useRef, useState } from 'react'

interface SearchableMultiSelectProps {
  label: string
  values: string[]
  options: string[]
  onChange: (values: string[]) => void
  placeholder?: string
  tone?: 'default' | 'danger'
  disabled?: boolean
}

const toneClasses: Record<NonNullable<SearchableMultiSelectProps['tone']>, string> = {
  default: 'bg-indigo-50 text-indigo-700 ring-indigo-200',
  danger: 'bg-rose-50 text-rose-700 ring-rose-200',
}

export function SearchableMultiSelect({
  label,
  values,
  options,
  onChange,
  placeholder,
  tone = 'default',
  disabled,
}: SearchableMultiSelectProps) {
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function onClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [])

  const query_ = query.trim().toLowerCase()
  const suggestions = options.filter((o) => !values.includes(o) && o.toLowerCase().includes(query_)).slice(0, 8)

  function add(v: string) {
    const trimmed = v.trim()
    if (trimmed && !values.includes(trimmed)) onChange([...values, trimmed])
    setQuery('')
  }

  function remove(v: string) {
    onChange(values.filter((x) => x !== v))
  }

  return (
    <div ref={containerRef} className="relative">
      <div className="mb-1.5 text-xs font-medium tracking-wide text-slate-500 uppercase">{label}</div>
      <div
        className={`flex flex-wrap items-center gap-1.5 rounded-lg border border-slate-200 px-2 py-1.5 ${
          disabled ? 'bg-slate-50' : 'bg-white'
        }`}
      >
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
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              setOpen(true)
            }}
            onFocus={() => setOpen(true)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ',') {
                e.preventDefault()
                if (suggestions.length > 0) add(suggestions[0])
                else if (query.trim()) add(query)
              } else if (e.key === 'Backspace' && query === '' && values.length > 0) {
                remove(values[values.length - 1])
              } else if (e.key === 'Escape') {
                setOpen(false)
              }
            }}
            placeholder={values.length === 0 ? (placeholder ?? 'Search…') : ''}
            className="min-w-[8ch] flex-1 border-none bg-transparent px-1 py-1 text-sm text-slate-700 outline-none placeholder:text-slate-400"
          />
        )}
        {values.length === 0 && disabled && <span className="text-sm text-slate-400">None</span>}
      </div>

      {!disabled && open && suggestions.length > 0 && (
        <ul className="absolute z-10 mt-1 max-h-48 w-full overflow-y-auto rounded-lg border border-slate-200 bg-white py-1 text-sm shadow-lg">
          {suggestions.map((s) => (
            <li key={s}>
              <button
                type="button"
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => add(s)}
                className="block w-full px-3 py-1.5 text-left text-slate-700 hover:bg-indigo-50"
              >
                {s}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
