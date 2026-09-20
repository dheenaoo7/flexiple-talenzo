import { useEffect, useState } from 'react'
import type { Filters, Rubric, RubricCriterion } from '../types'
import { RubricWeightScale } from './RubricWeightScale'
import { SearchableMultiSelect } from './SearchableMultiSelect'

interface FiltersRubricPanelProps {
  filters: Filters
  rubric: Rubric
  resetKey: string
  mode: 'review' | 'header'
  editable: boolean
  busy?: boolean
  applyLabel: string
  skillOptions: string[]
  titleOptions: string[]
  locationOptions: string[]
  companyTypeOptions: string[]
  onApply: (filters: Filters, rubric: Rubric) => void
}

function emptyCriterion(): RubricCriterion {
  return { id: `custom_${Date.now()}`, label: '', description: '', weight: 3 }
}

export function FiltersRubricPanel({
  filters,
  rubric,
  resetKey,
  mode,
  editable,
  busy,
  applyLabel,
  skillOptions,
  titleOptions,
  locationOptions,
  companyTypeOptions,
  onApply,
}: FiltersRubricPanelProps) {
  const [draftFilters, setDraftFilters] = useState(filters)
  const [draftRubric, setDraftRubric] = useState(rubric)
  const [expanded, setExpanded] = useState(mode === 'review')

  useEffect(() => {
    setDraftFilters(filters)
    setDraftRubric(rubric)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [resetKey])

  const dirty = JSON.stringify(draftFilters) !== JSON.stringify(filters) || JSON.stringify(draftRubric) !== JSON.stringify(rubric)

  function updateFilter<K extends keyof Filters>(key: K, value: Filters[K]) {
    setDraftFilters((f) => ({ ...f, [key]: value }))
  }

  function updateCriterion(id: string, patch: Partial<RubricCriterion>) {
    setDraftRubric((r) => ({ criteria: r.criteria.map((c) => (c.id === id ? { ...c, ...patch } : c)) }))
  }

  function removeCriterion(id: string) {
    setDraftRubric((r) => ({ criteria: r.criteria.filter((c) => c.id !== id) }))
  }

  function addCriterion() {
    setDraftRubric((r) => ({ criteria: [...r.criteria, emptyCriterion()] }))
  }

  const body = (
    <div className="space-y-5">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <SearchableMultiSelect
          label="Required skills"
          values={draftFilters.required_skills}
          options={skillOptions}
          onChange={(v) => updateFilter('required_skills', v)}
          disabled={!editable}
        />
        <SearchableMultiSelect
          label="Nice to have"
          values={draftFilters.nice_to_have_skills}
          options={skillOptions}
          onChange={(v) => updateFilter('nice_to_have_skills', v)}
          disabled={!editable}
        />
        <SearchableMultiSelect
          label="Excluded skills"
          values={draftFilters.excluded_skills}
          options={skillOptions}
          onChange={(v) => updateFilter('excluded_skills', v)}
          tone="danger"
          disabled={!editable}
        />
        <SearchableMultiSelect
          label="Title keywords"
          values={draftFilters.title_keywords}
          options={titleOptions}
          onChange={(v) => updateFilter('title_keywords', v)}
          disabled={!editable}
        />
        <SearchableMultiSelect
          label="Locations"
          values={draftFilters.locations}
          options={locationOptions}
          onChange={(v) => updateFilter('locations', v)}
          disabled={!editable}
        />
        <SearchableMultiSelect
          label="Company types"
          values={draftFilters.company_types}
          options={companyTypeOptions}
          onChange={(v) => updateFilter('company_types', v)}
          disabled={!editable}
        />
      </div>

      <div className="flex flex-wrap items-end gap-4">
        <label className="text-sm text-slate-600">
          <span className="mb-1 block text-xs font-medium tracking-wide text-slate-500 uppercase">Min years</span>
          <input
            type="number"
            min={0}
            disabled={!editable}
            value={draftFilters.min_years_experience ?? ''}
            onChange={(e) => updateFilter('min_years_experience', e.target.value === '' ? null : Number(e.target.value))}
            className="w-24 rounded-lg border border-slate-200 px-2 py-1.5 text-sm disabled:bg-slate-50"
          />
        </label>
        <label className="text-sm text-slate-600">
          <span className="mb-1 block text-xs font-medium tracking-wide text-slate-500 uppercase">Max years</span>
          <input
            type="number"
            min={0}
            disabled={!editable}
            value={draftFilters.max_years_experience ?? ''}
            onChange={(e) => updateFilter('max_years_experience', e.target.value === '' ? null : Number(e.target.value))}
            className="w-24 rounded-lg border border-slate-200 px-2 py-1.5 text-sm disabled:bg-slate-50"
          />
        </label>
        <label className="min-w-[16rem] flex-1 text-sm text-slate-600">
          <span className="mb-1 block text-xs font-medium tracking-wide text-slate-500 uppercase">Extra context</span>
          <input
            type="text"
            disabled={!editable}
            value={draftFilters.keywords}
            onChange={(e) => updateFilter('keywords', e.target.value)}
            className="w-full rounded-lg border border-slate-200 px-2 py-1.5 text-sm disabled:bg-slate-50"
            placeholder="e.g. fintech domain"
          />
        </label>
      </div>

      <div>
        <div className="mb-2 flex items-center justify-between">
          <span className="text-xs font-medium tracking-wide text-slate-500 uppercase">Rubric</span>
          {editable && (
            <button type="button" onClick={addCriterion} className="text-xs font-medium text-indigo-600 hover:text-indigo-500">
              + Add criterion
            </button>
          )}
        </div>
        <ul className="space-y-2">
          {draftRubric.criteria.map((c) => (
            <li key={c.id} className="rounded-xl border border-slate-200 bg-slate-50/60 p-3">
              <div className="flex items-center gap-3">
                <div className="flex-1 space-y-1">
                  <input
                    value={c.label}
                    disabled={!editable}
                    onChange={(e) => updateCriterion(c.id, { label: e.target.value })}
                    className="w-full border-none bg-transparent p-0 text-sm font-medium text-slate-900 outline-none disabled:opacity-90"
                    placeholder="Criterion label"
                  />
                  <input
                    value={c.description}
                    disabled={!editable}
                    onChange={(e) => updateCriterion(c.id, { description: e.target.value })}
                    className="w-full border-none bg-transparent p-0 text-xs text-slate-500 outline-none disabled:opacity-90"
                    placeholder="What to look for"
                  />
                </div>
                <RubricWeightScale
                  weight={c.weight}
                  onChange={(w) => updateCriterion(c.id, { weight: w })}
                  disabled={!editable}
                />
                {editable && (
                  <button
                    type="button"
                    onClick={() => removeCriterion(c.id)}
                    className="text-slate-400 hover:text-rose-500"
                    aria-label="Remove criterion"
                  >
                    ×
                  </button>
                )}
              </div>
            </li>
          ))}
        </ul>
      </div>

      {editable && (
        <div className="flex justify-end">
          <button
            type="button"
            disabled={busy || (mode === 'header' && !dirty)}
            onClick={() => onApply(draftFilters, draftRubric)}
            className="rounded-xl bg-indigo-600 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {busy ? 'Working…' : applyLabel}
          </button>
        </div>
      )}
    </div>
  )

  if (mode === 'review') {
    return (
      <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <h2 className="mb-1 text-lg font-semibold text-slate-900">Filters & Rubric</h2>
        <p className="mb-5 text-sm text-slate-500">Edit anything before we run the search.</p>
        {body}
      </div>
    )
  }

  return (
    <div className="rounded-2xl border border-slate-200 bg-white shadow-sm">
      <button
        type="button"
        onClick={() => setExpanded((e) => !e)}
        className="flex w-full items-center justify-between px-5 py-3 text-left"
      >
        <span className="text-sm font-semibold text-slate-800">
          Filters & Rubric{' '}
          <span className="font-normal text-slate-400">
            ({draftFilters.required_skills.length} required skills · {draftRubric.criteria.length} criteria)
          </span>
        </span>
        <span className="text-slate-400">{expanded ? '▲' : '▼'}</span>
      </button>
      {expanded && <div className="border-t border-slate-100 px-5 py-5">{body}</div>}
    </div>
  )
}
