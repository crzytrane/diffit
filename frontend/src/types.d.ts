// Existing types
type UploadableImageProps = {
  $name: string
  imgSrc: string
  setImageSrc: React.Dispatch<React.SetStateAction<string>>
}

// API Entity Types
interface Project {
  id: string
  name: string
  slug: string
  description: string
  repository_url: string
  default_branch: string
  created_at: string
  updated_at: string
}

interface Build {
  id: string
  project_id: string
  branch: string
  commit_sha: string
  commit_message: string
  pr_number: number | null
  status: 'pending' | 'processing' | 'completed' | 'failed'
  total_snapshots: number
  changed_snapshots: number
  approved_snapshots: number
  created_at: string
  updated_at: string
}

interface Snapshot {
  id: string
  build_id: string
  baseline_id: string | null
  name: string
  browser: string
  viewport: string
  width: number
  height: number
  base_image_path: string
  comparison_image_path: string
  diff_image_path: string
  diff_percentage: number
  status: 'pending' | 'processing' | 'completed' | 'failed'
  review_status: 'unreviewed' | 'approved' | 'rejected'
  created_at: string
  updated_at: string
}

interface Baseline {
  id: string
  project_id: string
  snapshot_id: string | null
  name: string
  branch: string
  browser: string
  viewport: string
  image_path: string
  created_at: string
  updated_at: string
}

interface PaginatedResponse<T> {
  data: T[]
  page: number
  per_page: number
  total: number
}

interface ApiError {
  error: string
}
