import { Calculator } from '../features/calculator/Calculator'
import styles from './CalculatorPage.module.scss'

export function HomePage() {
  return (
    <main className={styles.page}>
      <header className={styles.intro}>
        <p className={styles.eyebrow}>Everyday math</p>
        <h1>Calculator</h1>
        <p className={styles.description}>
          Enter two numbers with + or − between them, then press =. Other operations are coming later.
        </p>
      </header>
      <Calculator />
    </main>
  )
}
