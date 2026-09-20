interface ThinkingOverlayProps {
  stages: string[]
  variant?: 'fullscreen' | 'inline'
}

export function ThinkingOverlay({ stages, variant = 'fullscreen' }: ThinkingOverlayProps) {
  const current = stages[stages.length - 1] ?? 'Thinking…'
  const history = stages.slice(0, -1)

  const content = (
    <div className="flex flex-col items-center gap-3 text-center">
      <div className="relative flex h-14 w-14 items-end justify-center">
        <span className="mailbox-envelope absolute top-0 left-1/2 text-2xl">✉️</span>
        <span className="mailbox-base text-4xl">📬</span>
      </div>
      <p className="text-base font-medium text-slate-800">{current}</p>
      {history.length > 0 && (
        <ul className="space-y-0.5">
          {history.map((s, i) => (
            <li key={i} className="text-xs text-slate-400 line-through decoration-slate-300">
              {s}
            </li>
          ))}
        </ul>
      )}
    </div>
  )

  if (variant === 'inline') {
    return (
      <div className="flex items-center justify-center rounded-2xl border border-slate-200 bg-white/80 px-6 py-10 backdrop-blur-sm">
        {content}
      </div>
    )
  }

  return <div className="flex min-h-full items-center justify-center px-6">{content}</div>
}
