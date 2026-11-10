import { Calculator } from '../features/calculator/Calculator'
import styles from './CalculatorPage.module.scss'

export function HomePage() {
  return (
    <main className={styles.page}>
      <header className={styles.intro}>
        <p className={styles.eyebrow}>Everyday math</p>
        <h1>Calculator</h1>
        <p className={styles.description}>
          Enter an expression, then press =. Use ^ for powers and parentheses to group terms.
        </p>
      </header>
      <Calculator />
    </main>
  )
}
