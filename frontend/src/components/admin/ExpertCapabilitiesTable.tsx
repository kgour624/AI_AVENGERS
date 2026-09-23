import { useState } from 'react'
import { Card } from '@/components/ui/Card'
import type { Expert } from '@/types/expert'

/**
 * ExpertCapabilitiesTable - Feature #26
 * 
 * Displays expert capabilities in a structured table format.
 * Shows topics, depth levels, and chunk counts.
 * 
 * NOTE: This is a simplified version that displays aggregate stats
 * from the expert object. A full implementation would fetch individual
 * topics from GET /admin/experts/:id/capabilities endpoint (not yet implemented).
 * 
 * Current display:
 * - Total Topics: expert.totalTopics
 * - Avg Depth Level: expert.avgDepthLevel (displayed as 1-5 stars)
 * - Total Chunks: expert.totalChunks
 * - Avg Rating: expert.avgRating
 */

interface ExpertCapabilitiesTableProps {
  expert: Expert
}

function renderStars(level: number): string {
  const fullStars = Math.floor(level)
  const hasHalfStar = level % 1 >= 0.5
  const emptyStars = 5 - fullStars - (hasHalfStar ? 1 : 0)
  
  return (
    '★'.repeat(fullStars) +
    (hasHalfStar ? '½' : '') +
    '☆'.repeat(emptyStars)
  )
}

function getDepthLabel(level: number): string {
  if (level >= 4.5) return 'Expert'
  if (level >= 3.5) return 'Advanced'
  if (level >= 2.5) return 'Intermediate'
  if (level >= 1.5) return 'Basic'
  return 'Beginner'
}

function getDepthColor(level: number): string {
  if (level >= 4.5) return 'text-purple-400'
  if (level >= 3.5) return 'text-blue-400'
  if (level >= 2.5) return 'text-green-400'
  if (level >= 1.5) return 'text-yellow-400'
  return 'text-gray-400'
}

export function ExpertCapabilitiesTable({ expert }: ExpertCapabilitiesTableProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  return (
    <div className="mt-3">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex items-center gap-2 text-sm font-medium text-text-secondary hover:text-text-primary transition-colors"
      >
        <span>{isExpanded ? '▼' : '▶'}</span>
        <span>Capabilities Overview</span>
      </button>

      {isExpanded && (
        <Card className="mt-2 p-4" glow="cyan">
          <div className="space-y-3">
            {/* Summary Stats */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-text-secondary">Total Topics</p>
                <p className="text-lg font-semibold text-text-primary">{expert.totalTopics}</p>
              </div>
              <div>
                <p className="text-xs text-text-secondary">Total Chunks</p>
                <p className="text-lg font-semibold text-text-primary">{expert.totalChunks}</p>
              </div>
            </div>

            {/* Depth Level */}
            <div>
              <p className="text-xs text-text-secondary mb-1">Average Depth Level</p>
              <div className="flex items-center gap-3">
                <span className="text-2xl text-amber-400">
                  {renderStars(expert.avgDepthLevel)}
                </span>
                <span className={`text-sm font-medium ${getDepthColor(expert.avgDepthLevel)}`}>
                  {expert.avgDepthLevel.toFixed(1)} - {getDepthLabel(expert.avgDepthLevel)}
                </span>
              </div>
            </div>

            {/* Rating */}
            <div>
              <p className="text-xs text-text-secondary mb-1">Average Rating</p>
              <div className="flex items-center gap-2">
                <span className="text-lg font-semibold text-text-primary">
                  {expert.avgRating.toFixed(2)}
                </span>
                <span className="text-xs text-text-secondary">
                  ({expert.totalRatings ?? 0} ratings)
                </span>
              </div>
            </div>

            {/* Note about full implementation */}
            <div className="mt-4 p-3 bg-bg-tertiary rounded border border-border">
              <p className="text-xs text-text-secondary italic">
                💡 <strong>Note:</strong> This is an aggregate view. A full capabilities table
                showing individual topics with their depth levels and chunk counts would require
                a new backend endpoint: <code className="text-xs bg-bg-secondary px-1 py-0.5 rounded">GET /admin/experts/:id/capabilities</code>
              </p>
            </div>
          </div>
        </Card>
      )}
    </div>
  )
}
