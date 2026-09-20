import type { EmptyResultDiagnostic } from '../types'

interface EmptyResultsStateProps {
  diagnostic?: EmptyResultDiagnostic
  onEditFilters: () => void
}

export function EmptyResultsState({ diagnostic, onEditFilters }: EmptyResultsStateProps) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-2xl border border-slate-200 bg-white px-6 py-10 text-center shadow-sm">
      <span className="text-3xl">🔍</span>
      <h3 className="text-base font-semibold text-slate-900">No profiles matched these filters</h3>
      {diagnostic ? (
        <p className="max-w-md text-sm text-slate-500">
          <span className="font-medium text-slate-700">{diagnostic.most_restrictive_filter}</span> excluded the most
          candidates. {diagnostic.explanation}
        </p>
      ) : (
        <p className="max-w-md text-sm text-slate-500">Try loosening a filter and running the search again.</p>
      )}
      <button
        type="button"
        onClick={onEditFilters}
        className="mt-1 rounded-xl bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-500"
      >
        Edit filters
      </button>
    </div>
  )
}
