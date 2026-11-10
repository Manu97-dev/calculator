import styles from './Page.module.scss'

export function HomePage() {
  return (
    <main className={styles.page}>
      <section className={styles.panel} aria-labelledby="page-title">
        <h1 id="page-title">Calculator frontend</h1>
        <p>The frontend starter is ready for the first feature.</p>
      </section>
    </main>
  )
}
