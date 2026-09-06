import { useMemo } from 'react'
import { cn } from '@/utils/cn'
import type { ResponseMode } from '@/types/expert'

export interface ExpertAvatarProps {
  domain: string
  status: 'idle' | 'analyzing' | 'responded'
  mode?: ResponseMode
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

const SIZE_PX: Record<NonNullable<ExpertAvatarProps['size']>, number> = {
  sm: 32,
  md: 48,
  lg: 72,
}

/** Response mode -> Tailwind text-color utility, for the responded-state ring. */
const MODE_RING_CLASS: Record<ResponseMode, string> = {
  ADVISE: 'text-mode-advise',
  ASK: 'text-mode-ask',
  WARN: 'text-mode-warn',
  PUSH_BACK: 'text-mode-pushback',
  REFUSE: 'text-mode-refuse',
}

type BaseShape = 'hexagon' | 'cylinder' | 'shield' | 'blueprint' | 'core'

/**
 * domain is backend free text, not an enum (same finding already
 * documented for ExpertBadge's color-hash gap) - matched by keyword
 * rather than exact string, with a generic neural-core fallback so no
 * domain ever renders a blank/broken avatar. See
 * docs/ARC51_UI_CONTRACT.md §4.
 */
function resolveShape(domain: string): BaseShape {
  const d = domain.toLowerCase()
  if (d.includes('system') || d.includes('design')) return 'hexagon'
  if (d.includes('database') || d.includes('db') || d.includes('sql')) return 'cylinder'
  if (d.includes('security')) return 'shield'
  if (d.includes('architecture') || d.includes('infra')) return 'blueprint'
  return 'core'
}

type DomainColor = 'cyan' | 'amber' | 'violet' | 'purple'

function resolveDomainColor(domain: string): DomainColor {
  const d = domain.toLowerCase()
  if (d.includes('system') || d.includes('design')) return 'cyan'
  if (d.includes('database') || d.includes('db') || d.includes('sql')) return 'amber'
  if (d.includes('security')) return 'violet'
  return 'purple'
}

const DOMAIN_COLOR_CLASS: Record<DomainColor, string> = {
  cyan: 'text-glow-cyan',
  amber: 'text-glow-amber',
  violet: 'text-glow-violet',
  purple: 'text-glow-purple',
}

function BaseShapeSvg({ shape }: { shape: BaseShape }) {
  switch (shape) {
    case 'hexagon':
      return (
        <polygon
          points="50,6 90,28 90,72 50,94 10,72 10,28"
          fill="none"
          stroke="currentColor"
          strokeWidth="3"
          opacity="0.5"
        />
      )
    case 'cylinder':
      return (
        <g fill="none" stroke="currentColor" strokeWidth="3" opacity="0.5">
          <ellipse cx="50" cy="22" rx="32" ry="12" />
          <path d="M18,22 L18,72 A32,12 0 0 0 82,72 L82,22" />
          <ellipse cx="50" cy="50" rx="32" ry="12" opacity="0.6" />
        </g>
      )
    case 'shield':
      return (
        <path
          d="M50,6 L88,20 L88,50 C88,74 70,88 50,96 C30,88 12,74 12,50 L12,20 Z"
          fill="none"
          stroke="currentColor"
          strokeWidth="3"
          opacity="0.5"
        />
      )
    case 'blueprint':
      return (
        <g fill="none" stroke="currentColor" strokeWidth="2.5" opacity="0.5">
          <rect x="12" y="12" width="76" height="76" />
          <line x1="12" y1="38" x2="88" y2="38" />
          <line x1="12" y1="64" x2="88" y2="64" />
          <line x1="38" y1="12" x2="38" y2="88" />
          <line x1="64" y1="12" x2="64" y2="88" />
        </g>
      )
    case 'core':
    default:
      return <circle cx="50" cy="50" r="40" fill="none" stroke="currentColor" strokeWidth="3" opacity="0.5" />
  }
}

/**
 * Holographic/cybernetic silhouette overlay - admin's explicit
 * amendment to the original geometric-only proposal (§5 of the
 * contract): an abstract visor-line "face" so the avatar reads as a
 * living AI agent, not just an icon. Deliberately abstract (no
 * literal robot clip-art) - a rounded head outline + two glowing
 * visor lines, low opacity, layered in FRONT of the base shape.
 */
function HoloSilhouette() {
  return (
    <g opacity="0.85">
      <path
        d="M50,30 C62,30 70,40 70,52 C70,66 61,76 50,78 C39,76 30,66 30,52 C30,40 38,30 50,30 Z"
        fill="none"
        stroke="var(--glow-cyan)"
        strokeWidth="1.5"
        opacity="0.55"
      />
      {/* visor / eye-line */}
      <line x1="38" y1="52" x2="62" y2="52" stroke="var(--glow-cyan)" strokeWidth="3" strokeLinecap="round" />
      <line x1="42" y1="60" x2="58" y2="60" stroke="var(--glow-cyan)" strokeWidth="1.5" strokeLinecap="round" opacity="0.6" />
    </g>
  )
}

/**
 * Living-agent avatar for a domain expert. Pure presentational - no
 * data fetching. See docs/ARC51_UI_CONTRACT.md §4/§5.
 */
export function ExpertAvatar({ domain, status, mode, size = 'md', className }: ExpertAvatarProps) {
  const shape = useMemo(() => resolveShape(domain), [domain])
  const domainColorClass = useMemo(() => DOMAIN_COLOR_CLASS[resolveDomainColor(domain)], [domain])
  const px = SIZE_PX[size]

  const ringColorClass =
    status === 'responded' && mode ? MODE_RING_CLASS[mode] : 'text-glow-cyan'

  return (
    <div
      className={cn('relative inline-flex items-center justify-center', className)}
      style={{ width: px, height: px }}
      aria-hidden="true"
    >
      {/* Status ring - outside the composite, never overlapping the
          silhouette itself, so "is this agent active" stays legible
          even if the inner art gets busy (contract §5). */}
      <div
        className={cn(
          'absolute inset-0 rounded-full border-2',
          ringColorClass,
          status === 'idle' && 'border-current opacity-30',
          status === 'analyzing' &&
            'border-transparent bg-[conic-gradient(currentColor_0deg,transparent_270deg,currentColor_360deg)] p-[2px] arc-avatar-ring-analyzing',
          status === 'responded' && 'border-current arc-avatar-ring-responded'
        )}
      >
        {status === 'analyzing' && (
          <div className="h-full w-full rounded-full bg-surface-panel-hover" />
        )}
      </div>

      <svg viewBox="0 0 100 100" className={cn('relative h-[78%] w-[78%]', domainColorClass)}>
        <BaseShapeSvg shape={shape} />
        <HoloSilhouette />
      </svg>
    </div>
  )
}
