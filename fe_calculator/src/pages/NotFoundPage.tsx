import { Link } from 'react-router-dom'
import styles from './Page.module.scss'

export function NotFoundPage() {
  return (
    <main className={styles.page}>
      <section className={styles.panel} aria-labelledby="page-title">
        <h1 id="page-title">Page not found</h1>
        <p>The address you opened doesn’t match a page.</p>
        <Link to="/">Back to home</Link>
      </section>
    </main>
  )
}
