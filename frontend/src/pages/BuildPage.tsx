import { useEffect, useState, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getBuild, getProject, getSnapshots, reviewSnapshot, batchReviewSnapshots, getSnapshotImageUrl } from '../api'
import { Loading, EmptyState, ErrorState } from '../components/Loading'
import { StatusBadge } from '../components/StatusBadge'
import styles from './BuildPage.module.css'

type FilterType = 'all' | 'changed' | 'unreviewed' | 'approved' | 'rejected'

export function BuildPage() {
  const { buildId } = useParams<{ buildId: string }>()
  const [build, setBuild] = useState<Build | null>(null)
  const [project, setProject] = useState<Project | null>(null)
  const [snapshots, setSnapshots] = useState<Snapshot[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<FilterType>('all')
  const [selectedSnapshots, setSelectedSnapshots] = useState<Set<string>>(new Set())
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(null)

  const fetchData = useCallback(async () => {
    if (!buildId) return
    setLoading(true)
    setError(null)
    try {
      const buildData = await getBuild(buildId)
      setBuild(buildData)
      const projectData = await getProject(buildData.project_id)
      setProject(projectData)
      const snapshotsData = await getSnapshots(buildId, 1, 200)
      setSnapshots(snapshotsData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load build')
    } finally {
      setLoading(false)
    }
  }, [buildId])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleReview = async (snapshotId: string, action: 'approve' | 'reject') => {
    try {
      const updated = await reviewSnapshot(snapshotId, action)
      setSnapshots((prev) => prev.map((s) => (s.id === updated.id ? updated : s)))
      if (selectedSnapshot?.id === updated.id) {
        setSelectedSnapshot(updated)
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to review snapshot')
    }
  }

  const handleBatchReview = async (action: 'approve' | 'reject') => {
    if (selectedSnapshots.size === 0) return
    try {
      await batchReviewSnapshots(Array.from(selectedSnapshots), action)
      fetchData()
      setSelectedSnapshots(new Set())
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to review snapshots')
    }
  }

  const toggleSnapshotSelection = (id: string) => {
    setSelectedSnapshots((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const selectAllVisible = () => {
    const visibleIds = filteredSnapshots.map((s) => s.id)
    setSelectedSnapshots(new Set(visibleIds))
  }

  const clearSelection = () => {
    setSelectedSnapshots(new Set())
  }

  const filteredSnapshots = snapshots.filter((s) => {
    if (filter === 'all') return true
    if (filter === 'changed') return s.diff_percentage > 0
    return s.review_status === filter
  })

  if (loading) return <Loading message="Loading build..." />
  if (error) return <ErrorState message={error} onRetry={fetchData} />
  if (!build || !project) return <ErrorState message="Build not found" />

  return (
    <div className={styles.container}>
      <nav className={styles.breadcrumb}>
        <Link to="/">Projects</Link>
        <span className={styles.separator}>/</span>
        <Link to={`/projects/${project.slug}`}>{project.name}</Link>
        <span className={styles.separator}>/</span>
        <span>Build</span>
      </nav>

      <div className={styles.header}>
        <div>
          <div className={styles.titleRow}>
            <h1 className={styles.title}>{build.branch}</h1>
            <StatusBadge status={build.status} />
          </div>
          {build.commit_message && <p className={styles.commitMessage}>{build.commit_message}</p>}
          <div className={styles.meta}>
            {build.commit_sha && <code className={styles.sha}>{build.commit_sha.slice(0, 8)}</code>}
            <span className={styles.date}>{new Date(build.created_at).toLocaleString()}</span>
          </div>
        </div>
        <div className={styles.stats}>
          <div className={styles.statItem}>
            <span className={styles.statValue}>{build.total_snapshots}</span>
            <span className={styles.statLabel}>Total</span>
          </div>
          <div className={`${styles.statItem} ${styles.changed}`}>
            <span className={styles.statValue}>{build.changed_snapshots}</span>
            <span className={styles.statLabel}>Changed</span>
          </div>
          <div className={`${styles.statItem} ${styles.approved}`}>
            <span className={styles.statValue}>{build.approved_snapshots}</span>
            <span className={styles.statLabel}>Approved</span>
          </div>
        </div>
      </div>

      <div className={styles.toolbar}>
        <div className={styles.filters}>
          {(['all', 'changed', 'unreviewed', 'approved', 'rejected'] as FilterType[]).map((f) => (
            <button
              key={f}
              className={`${styles.filterButton} ${filter === f ? styles.active : ''}`}
              onClick={() => setFilter(f)}
            >
              {f.charAt(0).toUpperCase() + f.slice(1)}
            </button>
          ))}
        </div>
        {selectedSnapshots.size > 0 && (
          <div className={styles.batchActions}>
            <span className={styles.selectedCount}>{selectedSnapshots.size} selected</span>
            <button className={styles.batchApprove} onClick={() => handleBatchReview('approve')}>
              Approve All
            </button>
            <button className={styles.batchReject} onClick={() => handleBatchReview('reject')}>
              Reject All
            </button>
            <button className={styles.clearSelection} onClick={clearSelection}>
              Clear
            </button>
          </div>
        )}
        {selectedSnapshots.size === 0 && filteredSnapshots.length > 0 && (
          <button className={styles.selectAll} onClick={selectAllVisible}>
            Select All
          </button>
        )}
      </div>

      {filteredSnapshots.length === 0 ? (
        <EmptyState
          title="No snapshots found"
          description={filter === 'all' ? 'No snapshots in this build' : `No ${filter} snapshots`}
        />
      ) : (
        <div className={styles.snapshotGrid}>
          {filteredSnapshots.map((snapshot) => (
            <div
              key={snapshot.id}
              className={`${styles.snapshotCard} ${selectedSnapshots.has(snapshot.id) ? styles.selected : ''}`}
            >
              <div className={styles.snapshotCheckbox}>
                <input
                  type="checkbox"
                  checked={selectedSnapshots.has(snapshot.id)}
                  onChange={() => toggleSnapshotSelection(snapshot.id)}
                />
              </div>
              <div className={styles.snapshotPreview} onClick={() => setSelectedSnapshot(snapshot)}>
                {snapshot.diff_image_path ? (
                  <img
                    src={getSnapshotImageUrl(snapshot.id, 'diff')}
                    alt={snapshot.name}
                    className={styles.previewImage}
                  />
                ) : snapshot.comparison_image_path ? (
                  <img
                    src={getSnapshotImageUrl(snapshot.id, 'comparison')}
                    alt={snapshot.name}
                    className={styles.previewImage}
                  />
                ) : (
                  <div className={styles.noImage}>No image</div>
                )}
                {snapshot.diff_percentage > 0 && (
                  <span className={styles.diffBadge}>{snapshot.diff_percentage.toFixed(1)}%</span>
                )}
              </div>
              <div className={styles.snapshotInfo}>
                <span className={styles.snapshotName}>{snapshot.name}</span>
                <div className={styles.snapshotMeta}>
                  <span>{snapshot.browser}</span>
                  <span>{snapshot.viewport}</span>
                </div>
                <div className={styles.snapshotStatus}>
                  <StatusBadge status={snapshot.review_status} size="small" />
                </div>
              </div>
              <div className={styles.snapshotActions}>
                <button
                  className={`${styles.actionButton} ${styles.approveButton}`}
                  onClick={() => handleReview(snapshot.id, 'approve')}
                  disabled={snapshot.review_status === 'approved'}
                >
                  Approve
                </button>
                <button
                  className={`${styles.actionButton} ${styles.rejectButton}`}
                  onClick={() => handleReview(snapshot.id, 'reject')}
                  disabled={snapshot.review_status === 'rejected'}
                >
                  Reject
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {selectedSnapshot && (
        <SnapshotModal
          snapshot={selectedSnapshot}
          onClose={() => setSelectedSnapshot(null)}
          onReview={(action) => handleReview(selectedSnapshot.id, action)}
        />
      )}
    </div>
  )
}

interface SnapshotModalProps {
  snapshot: Snapshot
  onClose: () => void
  onReview: (action: 'approve' | 'reject') => void
}

function SnapshotModal({ snapshot, onClose, onReview }: SnapshotModalProps) {
  const [viewMode, setViewMode] = useState<'diff' | 'base' | 'comparison' | 'overlay'>('diff')
  const [overlayOpacity, setOverlayOpacity] = useState(0.5)

  return (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.modalHeader}>
          <div>
            <h2 className={styles.modalTitle}>{snapshot.name}</h2>
            <div className={styles.modalMeta}>
              <span>{snapshot.browser}</span>
              <span>{snapshot.viewport}</span>
              <span>{snapshot.width}x{snapshot.height}</span>
              <StatusBadge status={snapshot.review_status} size="small" />
              {snapshot.diff_percentage > 0 && (
                <span className={styles.diffPercent}>{snapshot.diff_percentage.toFixed(2)}% diff</span>
              )}
            </div>
          </div>
          <button className={styles.closeButton} onClick={onClose}>
            Close
          </button>
        </div>

        <div className={styles.viewModes}>
          {(['diff', 'base', 'comparison', 'overlay'] as const).map((mode) => (
            <button
              key={mode}
              className={`${styles.viewModeButton} ${viewMode === mode ? styles.active : ''}`}
              onClick={() => setViewMode(mode)}
            >
              {mode.charAt(0).toUpperCase() + mode.slice(1)}
            </button>
          ))}
        </div>

        {viewMode === 'overlay' && (
          <div className={styles.opacityControl}>
            <label>Opacity: {Math.round(overlayOpacity * 100)}%</label>
            <input
              type="range"
              min="0"
              max="1"
              step="0.01"
              value={overlayOpacity}
              onChange={(e) => setOverlayOpacity(parseFloat(e.target.value))}
            />
          </div>
        )}

        <div className={styles.imageContainer}>
          {viewMode === 'diff' && snapshot.diff_image_path && (
            <img src={getSnapshotImageUrl(snapshot.id, 'diff')} alt="Diff" className={styles.modalImage} />
          )}
          {viewMode === 'base' && snapshot.base_image_path && (
            <img src={getSnapshotImageUrl(snapshot.id, 'base')} alt="Base" className={styles.modalImage} />
          )}
          {viewMode === 'comparison' && snapshot.comparison_image_path && (
            <img src={getSnapshotImageUrl(snapshot.id, 'comparison')} alt="Comparison" className={styles.modalImage} />
          )}
          {viewMode === 'overlay' && (
            <div className={styles.overlayContainer}>
              {snapshot.base_image_path && (
                <img src={getSnapshotImageUrl(snapshot.id, 'base')} alt="Base" className={styles.overlayBase} />
              )}
              {snapshot.comparison_image_path && (
                <img
                  src={getSnapshotImageUrl(snapshot.id, 'comparison')}
                  alt="Comparison"
                  className={styles.overlayComparison}
                  style={{ opacity: overlayOpacity }}
                />
              )}
            </div>
          )}
        </div>

        <div className={styles.modalActions}>
          <button
            className={`${styles.reviewButton} ${styles.approveButton}`}
            onClick={() => onReview('approve')}
            disabled={snapshot.review_status === 'approved'}
          >
            Approve
          </button>
          <button
            className={`${styles.reviewButton} ${styles.rejectButton}`}
            onClick={() => onReview('reject')}
            disabled={snapshot.review_status === 'rejected'}
          >
            Reject
          </button>
        </div>
      </div>
    </div>
  )
}
