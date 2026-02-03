import { useEffect, useState, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { getProjects, createProject, deleteProject } from '../api'
import { Loading, EmptyState, ErrorState } from '../components/Loading'
import { Pagination } from '../components/Pagination'
import styles from './ProjectsPage.module.css'

export function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const perPage = 12

  const fetchProjects = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await getProjects(page, perPage)
      setProjects(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects')
    } finally {
      setLoading(false)
    }
  }, [page])

  useEffect(() => {
    fetchProjects()
  }, [fetchProjects])

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Delete project "${name}"? This cannot be undone.`)) return
    try {
      await deleteProject(id)
      fetchProjects()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete project')
    }
  }

  if (loading) return <Loading message="Loading projects..." />
  if (error) return <ErrorState message={error} onRetry={fetchProjects} />

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>Projects</h1>
        <button className={styles.createButton} onClick={() => setShowCreateModal(true)}>
          New Project
        </button>
      </div>

      {projects.length === 0 ? (
        <EmptyState
          title="No projects yet"
          description="Create your first project to start visual testing"
          action={
            <button className={styles.createButton} onClick={() => setShowCreateModal(true)}>
              Create Project
            </button>
          }
        />
      ) : (
        <>
          <div className={styles.grid}>
            {projects.map((project) => (
              <div key={project.id} className={styles.card}>
                <Link to={`/projects/${project.slug}`} className={styles.cardLink}>
                  <h3 className={styles.cardTitle}>{project.name}</h3>
                  {project.description && (
                    <p className={styles.cardDescription}>{project.description}</p>
                  )}
                  <div className={styles.cardMeta}>
                    <span className={styles.slug}>/{project.slug}</span>
                    {project.default_branch && (
                      <span className={styles.branch}>{project.default_branch}</span>
                    )}
                  </div>
                </Link>
                <button
                  className={styles.deleteButton}
                  onClick={(e) => {
                    e.preventDefault()
                    handleDelete(project.id, project.name)
                  }}
                  title="Delete project"
                >
                  Delete
                </button>
              </div>
            ))}
          </div>
          <Pagination
            page={page}
            hasMore={projects.length === perPage}
            onPageChange={setPage}
          />
        </>
      )}

      {showCreateModal && (
        <CreateProjectModal
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false)
            fetchProjects()
          }}
        />
      )}
    </div>
  )
}

interface CreateProjectModalProps {
  onClose: () => void
  onCreated: () => void
}

function CreateProjectModal({ onClose, onCreated }: CreateProjectModalProps) {
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [description, setDescription] = useState('')
  const [repositoryUrl, setRepositoryUrl] = useState('')
  const [defaultBranch, setDefaultBranch] = useState('main')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await createProject({
        name,
        slug: slug || name.toLowerCase().replace(/[^a-z0-9]+/g, '-'),
        description,
        repository_url: repositoryUrl,
        default_branch: defaultBranch,
      })
      onCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create project')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <h2 className={styles.modalTitle}>Create Project</h2>
        <form onSubmit={handleSubmit}>
          <div className={styles.field}>
            <label className={styles.label}>Name *</label>
            <input
              type="text"
              className={styles.input}
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              autoFocus
            />
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Slug</label>
            <input
              type="text"
              className={styles.input}
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder={name.toLowerCase().replace(/[^a-z0-9]+/g, '-') || 'auto-generated'}
            />
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Description</label>
            <textarea
              className={styles.textarea}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Repository URL</label>
            <input
              type="url"
              className={styles.input}
              value={repositoryUrl}
              onChange={(e) => setRepositoryUrl(e.target.value)}
              placeholder="https://github.com/org/repo"
            />
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Default Branch</label>
            <input
              type="text"
              className={styles.input}
              value={defaultBranch}
              onChange={(e) => setDefaultBranch(e.target.value)}
            />
          </div>
          {error && <p className={styles.error}>{error}</p>}
          <div className={styles.modalActions}>
            <button type="button" className={styles.cancelButton} onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className={styles.submitButton} disabled={submitting || !name}>
              {submitting ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
