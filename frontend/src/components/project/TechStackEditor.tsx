import { useState, useEffect } from 'react'
import { updateProject } from '@/api/projects'
import { Button } from '@/components/ui/Button'

interface TechStackEditorProps {
  projectId: string
  techStack?: string[]
  onUpdated: () => void
}

/**
 * TechStackEditor - Feature #24
 * 
 * Displays and edits tech stack tags for a project.
 * Tags are stored as a string array in the frontend but sent as
 * an object to the backend (keys = tag names, values = true).
 * 
 * UI Pattern:
 * - Display tags as chips with × remove button
 * - Input field + "Add" button to add new tags
 * - Auto-save on add/remove (no separate Save button)
 */
export function TechStackEditor({ projectId, techStack = [], onUpdated }: TechStackEditorProps) {
  const [tags, setTags] = useState<string[]>([])
  const [newTag, setNewTag] = useState('')
  const [isAdding, setIsAdding] = useState(false)

  // Convert techStack prop to tags array on mount/update
  useEffect(() => {
    if (Array.isArray(techStack)) {
      setTags(techStack)
    } else if (techStack && typeof techStack === 'object') {
      // Backend might return object format {"React": true, "Node.js": true}
      setTags(Object.keys(techStack))
    } else {
      setTags([])
    }
  }, [techStack])

  const handleAddTag = async () => {
    const trimmed = newTag.trim()
    if (!trimmed) return
    if (tags.includes(trimmed)) {
      alert('Tag already exists')
      return
    }

    setIsAdding(true)
    const updatedTags = [...tags, trimmed]
    try {
      // Backend expects tech_stack as map[string]interface{}
      // We'll send it as an object with tags as keys (value = true)
      const techStackObj = Object.fromEntries(updatedTags.map(tag => [tag, true]))
      await updateProject(projectId, { tech_stack: techStackObj })
      setTags(updatedTags)
      setNewTag('')
      onUpdated()
    } catch (error) {
      console.error('Failed to add tag:', error)
      alert('Failed to add tag')
    } finally {
      setIsAdding(false)
    }
  }

  const handleRemoveTag = async (tagToRemove: string) => {
    const updatedTags = tags.filter(tag => tag !== tagToRemove)
    try {
      const techStackObj = Object.fromEntries(updatedTags.map(tag => [tag, true]))
      await updateProject(projectId, { tech_stack: techStackObj })
      setTags(updatedTags)
      onUpdated()
    } catch (error) {
      console.error('Failed to remove tag:', error)
      alert('Failed to remove tag')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddTag()
    }
  }

  return (
    <div>
      {/* Display existing tags */}
      <div className="flex flex-wrap gap-2 mb-3">
        {tags.length === 0 && (
          <p className="text-sm text-text-secondary italic">No tech stack tags yet</p>
        )}
        {tags.map(tag => (
          <div
            key={tag}
            className="inline-flex items-center gap-1.5 px-3 py-1 bg-bg-secondary rounded-full text-sm"
          >
            <span>{tag}</span>
            <button
              onClick={() => handleRemoveTag(tag)}
              className="text-text-secondary hover:text-text-primary transition-colors"
              aria-label={`Remove ${tag}`}
            >
              ×
            </button>
          </div>
        ))}
      </div>

      {/* Add new tag */}
      <div className="flex gap-2">
        <input
          type="text"
          value={newTag}
          onChange={e => setNewTag(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="e.g. React, Node.js, PostgreSQL"
          className="flex-1 px-3 py-1.5 text-sm bg-bg-secondary border border-border rounded focus:outline-none focus:ring-2 focus:ring-primary"
          disabled={isAdding}
        />
        <Button
          size="sm"
          onClick={handleAddTag}
          disabled={!newTag.trim() || isAdding}
        >
          {isAdding ? 'Adding...' : 'Add Tag'}
        </Button>
      </div>
    </div>
  )
}
