const API_BASE = import.meta.env.PROD
  ? 'https://diffit-api.markhamilton.dev'
  : 'http://localhost:4007'

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(error.error || `HTTP ${response.status}`)
  }

  return response.json()
}

// Projects API
export async function getProjects(page = 1, perPage = 20): Promise<Project[]> {
  return fetchApi<Project[]>(`/api/projects?page=${page}&per_page=${perPage}`)
}

export async function getProject(id: string): Promise<Project> {
  return fetchApi<Project>(`/api/projects/${id}`)
}

export async function getProjectBySlug(slug: string): Promise<Project> {
  return fetchApi<Project>(`/api/projects/slug/${slug}`)
}

export async function createProject(data: Partial<Project>): Promise<Project> {
  return fetchApi<Project>('/api/projects', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updateProject(id: string, data: Partial<Project>): Promise<Project> {
  return fetchApi<Project>(`/api/projects/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deleteProject(id: string): Promise<void> {
  await fetchApi(`/api/projects/${id}`, { method: 'DELETE' })
}

// Builds API
export async function getBuilds(projectId: string, page = 1, perPage = 20): Promise<Build[]> {
  return fetchApi<Build[]>(`/api/projects/${projectId}/builds?page=${page}&per_page=${perPage}`)
}

export async function getBuild(id: string): Promise<Build> {
  return fetchApi<Build>(`/api/builds/${id}`)
}

export async function getLatestBuild(projectId: string): Promise<Build | null> {
  try {
    return await fetchApi<Build>(`/api/projects/${projectId}/builds/latest`)
  } catch {
    return null
  }
}

export async function createBuild(data: Partial<Build>): Promise<Build> {
  return fetchApi<Build>('/api/builds', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function deleteBuild(id: string): Promise<void> {
  await fetchApi(`/api/builds/${id}`, { method: 'DELETE' })
}

export async function finalizeBuild(id: string): Promise<Build> {
  return fetchApi<Build>(`/api/builds/${id}/finalize`, { method: 'POST' })
}

// Snapshots API
export async function getSnapshots(buildId: string, page = 1, perPage = 50): Promise<Snapshot[]> {
  return fetchApi<Snapshot[]>(`/api/builds/${buildId}/snapshots?page=${page}&per_page=${perPage}`)
}

export async function getChangedSnapshots(buildId: string): Promise<Snapshot[]> {
  return fetchApi<Snapshot[]>(`/api/builds/${buildId}/snapshots/changed`)
}

export async function getSnapshot(id: string): Promise<Snapshot> {
  return fetchApi<Snapshot>(`/api/snapshots/${id}`)
}

export async function reviewSnapshot(id: string, action: 'approve' | 'reject'): Promise<Snapshot> {
  return fetchApi<Snapshot>(`/api/snapshots/${id}/review`, {
    method: 'POST',
    body: JSON.stringify({ action }),
  })
}

export async function batchReviewSnapshots(snapshotIds: string[], action: 'approve' | 'reject'): Promise<{ updated: number }> {
  return fetchApi<{ updated: number }>('/api/snapshots/batch-review', {
    method: 'POST',
    body: JSON.stringify({ snapshot_ids: snapshotIds, action }),
  })
}

export function getSnapshotImageUrl(snapshotId: string, imageType: 'base' | 'comparison' | 'diff'): string {
  return `${API_BASE}/api/snapshots/${snapshotId}/image/${imageType}`
}

// Baselines API
export async function getBaselines(projectId: string, page = 1, perPage = 50): Promise<Baseline[]> {
  return fetchApi<Baseline[]>(`/api/projects/${projectId}/baselines?page=${page}&per_page=${perPage}`)
}

export async function getBaseline(id: string): Promise<Baseline> {
  return fetchApi<Baseline>(`/api/baselines/${id}`)
}

export async function createBaselineFromSnapshot(snapshotId: string): Promise<Baseline> {
  return fetchApi<Baseline>('/api/baselines/from-snapshot', {
    method: 'POST',
    body: JSON.stringify({ snapshot_id: snapshotId }),
  })
}

export async function deleteBaseline(id: string): Promise<void> {
  await fetchApi(`/api/baselines/${id}`, { method: 'DELETE' })
}

export function getBaselineImageUrl(baselineId: string): string {
  return `${API_BASE}/api/baselines/${baselineId}/image`
}
