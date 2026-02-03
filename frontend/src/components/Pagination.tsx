import styles from './Pagination.module.css'

interface PaginationProps {
  page: number
  hasMore: boolean
  onPageChange: (page: number) => void
}

export function Pagination({ page, hasMore, onPageChange }: PaginationProps) {
  return (
    <div className={styles.pagination}>
      <button
        className={styles.button}
        onClick={() => onPageChange(page - 1)}
        disabled={page <= 1}
      >
        Previous
      </button>
      <span className={styles.pageInfo}>Page {page}</span>
      <button
        className={styles.button}
        onClick={() => onPageChange(page + 1)}
        disabled={!hasMore}
      >
        Next
      </button>
    </div>
  )
}
