import styles from './StatusBadge.module.css'

type Status = 'pending' | 'processing' | 'completed' | 'failed' | 'unreviewed' | 'approved' | 'rejected'

interface StatusBadgeProps {
  status: Status
  size?: 'small' | 'medium'
}

export function StatusBadge({ status, size = 'medium' }: StatusBadgeProps) {
  return (
    <span className={`${styles.badge} ${styles[status]} ${styles[size]}`}>
      {status}
    </span>
  )
}
