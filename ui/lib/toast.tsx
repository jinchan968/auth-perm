'use client'

import { toast as sonnerToast } from 'sonner'
import { X } from 'lucide-react'

export function showError(message: string) {
  sonnerToast.custom((t) => (
    <div
      className="group flex items-start w-full rounded-lg"
      style={{
        background: 'hsl(var(--card) / 0.9)',
        border: '2px solid #ef4444',
        boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
        padding: '16px 20px',
      }}
    >
      <p className="flex-1 min-w-0 text-sm leading-relaxed" style={{ color: 'hsl(var(--foreground) / 0.85)' }}>
        {message}
      </p>
      <div className="flex items-center shrink-0 pl-3">
        <button
          onClick={() => sonnerToast.dismiss(t)}
          className="opacity-0 group-hover:opacity-100 transition-opacity h-7 w-7 flex items-center justify-center rounded-sm"
          style={{ color: 'hsl(var(--muted-foreground))' }}
          onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'hsl(var(--muted))' }}
          onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  ), { unstyled: true })
}

export function showSuccess(message: string) {
  sonnerToast.custom((t) => (
    <div
      className="group flex items-start w-full rounded-lg"
      style={{
        background: 'hsl(var(--card) / 0.9)',
        border: '1px solid hsl(var(--border))',
        boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
        padding: '16px 20px',
      }}
    >
      <p className="flex-1 min-w-0 text-sm leading-relaxed" style={{ color: 'hsl(var(--foreground) / 0.85)' }}>
        {message}
      </p>
      <div className="flex items-center shrink-0 pl-3">
        <button
          onClick={() => sonnerToast.dismiss(t)}
          className="opacity-0 group-hover:opacity-100 transition-opacity h-7 w-7 flex items-center justify-center rounded-sm"
          style={{ color: 'hsl(var(--muted-foreground))' }}
          onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = 'hsl(var(--muted))' }}
          onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent' }}
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  ), { unstyled: true })
}
