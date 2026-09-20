interface RubricWeightScaleProps {
  weight: number
  onChange: (weight: number) => void
  disabled?: boolean
}

const LEVELS = [
  { value: 1, label: 'Optional' },
  { value: 2, label: 'Helpful' },
  { value: 3, label: 'Important' },
  { value: 4, label: 'Very important' },
  { value: 5, label: 'Critical' },
]

export function RubricWeightScale({ weight, onChange, disabled }: RubricWeightScaleProps) {
  const activeIndex = Math.max(
    0,
    LEVELS.findIndex((l) => l.value === weight),
  )
  const fillPercent = (activeIndex / (LEVELS.length - 1)) * 100

  return (
    <div className="w-40 shrink-0 select-none">
      <div className="relative h-5">
        <div className="absolute top-1/2 left-0 h-1.5 w-full -translate-y-1/2 rounded-full bg-slate-200" />
        <div
          className="absolute top-1/2 left-0 h-1.5 -translate-y-1/2 rounded-full bg-gradient-to-r from-indigo-300 to-indigo-500 transition-all duration-300 ease-out"
          style={{ width: `${fillPercent}%` }}
        />
        {LEVELS.map((level, i) => {
          const isActive = level.value === weight
          const pos = (i / (LEVELS.length - 1)) * 100
          return (
            <button
              key={level.value}
              type="button"
              disabled={disabled}
              onClick={() => onChange(level.value)}
              title={level.label}
              aria-label={`Set weight ${level.value}: ${level.label}`}
              style={{ left: `${pos}%` }}
              className={`absolute top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full transition-all duration-300 ease-out ${
                isActive
                  ? 'h-3.5 w-3.5 bg-indigo-600 shadow-[0_0_0_4px_rgba(99,102,241,0.25)]'
                  : 'h-2 w-2 bg-white ring-2 ring-slate-300'
              } ${disabled ? '' : 'cursor-pointer hover:ring-2 hover:ring-indigo-400'}`}
            >
              {isActive && (
                <span className="absolute inset-0 -m-1 animate-ping rounded-full bg-indigo-400 opacity-40" />
              )}
            </button>
          )
        })}
      </div>
      <div className="mt-1.5 text-center text-[11px] font-medium tracking-wide text-indigo-600 transition-all duration-300">
        {LEVELS[activeIndex].label}
      </div>
    </div>
  )
}
