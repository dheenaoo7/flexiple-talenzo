interface ErrorBannerProps {
  stage: string
  message: string
  retryable: boolean
  onRetry: () => void
  onDismiss: () => void
}

export function ErrorBanner({ stage, message, retryable, onRetry, onDismiss }: ErrorBannerProps) {
  return (
    <div className="flex items-start gap-3 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-rose-900">
      <span className="mt-0.5 text-lg">⚠</span>
      <div className="flex-1">
        <p className="text-sm font-medium">Something went wrong while {stage}.</p>
        <p className="mt-0.5 text-sm text-rose-700">{message}</p>
        <p className="mt-1 text-xs text-rose-500">Nothing you've done so far was lost.</p>
      </div>
      <div className="flex shrink-0 gap-2">
        {retryable && (
          <button
            type="button"
            onClick={onRetry}
            className="rounded-lg bg-rose-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-rose-500"
          >
            Retry
          </button>
        )}
        <button
          type="button"
          onClick={onDismiss}
          className="rounded-lg border border-rose-200 px-3 py-1.5 text-xs font-medium text-rose-700 hover:bg-rose-100"
        >
          Dismiss
        </button>
      </div>
    </div>
  )
}
