import { useEffect, useState, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getProjectBySlug, getBuilds, deleteBuild } from '../api'
import { Loading, EmptyState, ErrorState } from '../components/Loading'
import { StatusBadge } from '../components/StatusBadge'
import { Pagination } from '../components/Pagination'
import styles from './ProjectPage.module.css'

export function ProjectPage() {
  const { slug } = useParams<{ slug: string }>()
  const [project, setProject] = useState<Project | null>(null)
  const [builds, setBuilds] = useState<Build[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const perPage = 20

  const fetchData = useCallback(async () => {
    if (!slug) return
    setLoading(true)
    setError(null)
    try {
      const projectData = await getProjectBySlug(slug)
      setProject(projectData)
      const buildsData = await getBuilds(projectData.id, page, perPage)
      setBuilds(buildsData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load project')
    } finally {
      setLoading(false)
    }
  }, [slug, page])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleDeleteBuild = async (id: string) => {
    if (!confirm('Delete this build? This cannot be undone.')) return
    try {
      await deleteBuild(id)
      fetchData()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete build')
    }
  }

  if (loading) return <Loading message="Loading project..." />
  if (error) return <ErrorState message={error} onRetry={fetchData} />
  if (!project) return <ErrorState message="Project not found" />

  return (
    <div className={styles.container}>
      <nav className={styles.breadcrumb}>
        <Link to="/">Projects</Link>
        <span className={styles.separator}>/</span>
        <span>{project.name}</span>
      </nav>

      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>{project.name}</h1>
          {project.description && <p className={styles.description}>{project.description}</p>}
          <div className={styles.meta}>
            {project.repository_url && (
              <a
                href={project.repository_url}
                target="_blank"
                rel="noopener noreferrer"
                className={styles.repoLink}
              >
                Repository
              </a>
            )}
            {project.default_branch && (
              <span className={styles.branch}>{project.default_branch}</span>
            )}
          </div>
        </div>
      </div>

      <h2 className={styles.sectionTitle}>Builds</h2>

      {builds.length === 0 ? (
        <EmptyState
          title="No builds yet"
          description="Builds will appear here when you run visual tests"
        />
      ) : (
        <>
          <div className={styles.buildsList}>
            {builds.map((build) => (
              <div key={build.id} className={styles.buildCard}>
                <Link to={`/builds/${build.id}`} className={styles.buildLink}>
                  <div className={styles.buildHeader}>
                    <div className={styles.buildInfo}>
                      <span className={styles.buildBranch}>{build.branch}</span>
                      <StatusBadge status={build.status} size="small" />
                    </div>
                    <span className={styles.buildDate}>
                      {new Date(build.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  {build.commit_message && (
                    <p className={styles.commitMessage}>{build.commit_message}</p>
                  )}
                  {build.commit_sha && (
                    <code className={styles.commitSha}>{build.commit_sha.slice(0, 8)}</code>
                  )}
                  <div className={styles.buildStats}>
                    <span className={styles.stat}>
                      <strong>{build.total_snapshots}</strong> snapshots
                    </span>
                    {build.changed_snapshots > 0 && (
                      <span className={`${styles.stat} ${styles.changed}`}>
                        <strong>{build.changed_snapshots}</strong> changed
                      </span>
                    )}
                    {build.approved_snapshots > 0 && (
                      <span className={`${styles.stat} ${styles.approved}`}>
                        <strong>{build.approved_snapshots}</strong> approved
                      </span>
                    )}
                  </div>
                </Link>
                <button
                  className={styles.deleteButton}
                  onClick={(e) => {
                    e.preventDefault()
                    handleDeleteBuild(build.id)
                  }}
                  title="Delete build"
                >
                  Delete
                </button>
              </div>
            ))}
          </div>
          <Pagination page={page} hasMore={builds.length === perPage} onPageChange={setPage} />
        </>
      )}
    </div>
  )
}
