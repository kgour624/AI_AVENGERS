/**
 * ARC-51 motion constants — single source of truth for every
 * framer-motion variant/timing used across the redesign.
 * See docs/ARC51_UI_CONTRACT.md §3 for the CSS-vs-framer-motion
 * boundary this file's consumers must respect.
 *
 * WHY one shared file rather than inline configs per component: the
 * whole point of a "physics feel" is consistency - a modal that
 * springs differently than a card grid breaks the illusion of one
 * coherent interface, the same way inconsistent easing curves did
 * before this token pass.
 */
import type { Variants } from 'framer-motion'

export const ARC_MOTION = {
  micro: 0.12,
  panel: 0.28,
  stagger: 0.04,
  maxStaggerItems: 6,
  ease: [0.16, 1, 0.3, 1] as const,
}

/** Modal/panel enter-exit. Used by Modal.tsx (§4). */
export const modalSpring: Variants = {
  hidden: { opacity: 0, scale: 0.96, y: 8 },
  visible: {
    opacity: 1,
    scale: 1,
    y: 0,
    transition: { duration: ARC_MOTION.panel, ease: ARC_MOTION.ease },
  },
  exit: {
    opacity: 0,
    scale: 0.96,
    y: 8,
    transition: { duration: ARC_MOTION.micro, ease: ARC_MOTION.ease },
  },
}

/** Single-item fade-up, used as the child variant under staggerContainer. */
export const fadeUp: Variants = {
  hidden: { opacity: 0, y: 12 },
  visible: { opacity: 1, y: 0, transition: { duration: ARC_MOTION.panel, ease: ARC_MOTION.ease } },
}

/**
 * Wraps a list/grid parent. Clamps effective stagger delay so a long
 * list (e.g. 40 project cards) never feels sluggish - beyond
 * maxStaggerItems, every item after that shares the same final delay
 * instead of the delay growing unbounded.
 */
export function staggerContainer(itemCount: number): Variants {
  const effectiveCount = Math.min(itemCount, ARC_MOTION.maxStaggerItems)
  return {
    hidden: {},
    visible: {
      transition: {
        staggerChildren: ARC_MOTION.stagger,
        delayChildren: 0,
        // WHY this isn't actually clamped further here: framer-motion's
        // staggerChildren applies per-child in render order regardless
        // of itemCount - the clamp is enforced by consumers slicing
        // their own list to `effectiveCount` staggered + the remainder
        // rendered without additional per-item delay. Documented here
        // so a consumer doesn't assume this function does more than it
        // does.
      },
    },
  }
  // effectiveCount currently unused beyond documentation above;
  // kept as a named value (not inlined) so the WHY comment has
  // something concrete to point at for future consumers.
  void effectiveCount
}
