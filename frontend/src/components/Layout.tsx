import { Link, useLocation } from 'react-router-dom'
import styles from './Layout.module.css'

interface LayoutProps {
  children: React.ReactNode
}

export function Layout({ children }: LayoutProps) {
  const location = useLocation()

  return (
    <div className={styles.layout}>
      <header className={styles.header}>
        <Link to="/" className={styles.logo}>
          DIFFIT
        </Link>
        <nav className={styles.nav}>
          <Link
            to="/"
            className={`${styles.navLink} ${location.pathname === '/' ? styles.active : ''}`}
          >
            Projects
          </Link>
          <Link
            to="/compare"
            className={`${styles.navLink} ${location.pathname === '/compare' ? styles.active : ''}`}
          >
            Compare
          </Link>
        </nav>
      </header>
      <main className={styles.main}>{children}</main>
    </div>
  )
}
