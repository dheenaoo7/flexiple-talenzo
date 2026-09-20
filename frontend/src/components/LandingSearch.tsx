import { useState } from 'react'

interface LandingSearchProps {
  onSubmit: (query: string) => void
}

const EXAMPLES = [
  'Senior backend engineer with deep AWS RDS experience, startup pace',
  'Frontend engineer strong in accessibility and design systems',
  'DevOps engineer with Terraform and Kubernetes, remote-friendly',
]

export function LandingSearch({ onSubmit }: LandingSearchProps) {
  const [query, setQuery] = useState('')

  function submit() {
    const trimmed = query.trim()
    if (trimmed) onSubmit(trimmed)
  }

  return (
    <div className="flex min-h-full flex-col items-center justify-center px-6">
      <div className="w-full max-w-2xl text-center">
        <p className="mb-2 text-sm font-medium tracking-wide text-indigo-600 uppercase">Rubric-ARM</p>
        <h1 className="mb-3 text-3xl font-semibold text-slate-900 sm:text-4xl">Who are you looking for?</h1>
        <p className="mb-8 text-slate-500">
          Describe the role in plain language. We'll turn it into editable filters and a rubric, then find the best
          matches.
        </p>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            submit()
          }}
          className="flex items-center gap-2 rounded-2xl border border-slate-200 bg-white p-2 shadow-sm focus-within:border-indigo-300 focus-within:ring-4 focus-within:ring-indigo-50"
        >
          <input
            autoFocus
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="e.g. Senior backend engineer with AWS RDS experience at a startup"
            className="flex-1 border-none bg-transparent px-3 py-2.5 text-base text-slate-900 outline-none placeholder:text-slate-400"
          />
          <button
            type="submit"
            disabled={!query.trim()}
            className="rounded-xl bg-indigo-600 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-40"
          >
            Search
          </button>
        </form>

        <div className="mt-6 flex flex-wrap justify-center gap-2">
          {EXAMPLES.map((ex) => (
            <button
              key={ex}
              type="button"
              onClick={() => onSubmit(ex)}
              className="rounded-full border border-slate-200 bg-white px-3 py-1.5 text-xs text-slate-600 transition hover:border-indigo-200 hover:bg-indigo-50 hover:text-indigo-700"
            >
              {ex}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
