import styles from './Calculator.module.scss'

const keys = [
  { value: '^', label: 'Power', tone: 'extra' },
  { value: '√', label: 'Square root', tone: 'extra' },
  { value: '%', label: 'Percent', tone: 'extra' },
  { value: '⌫', label: 'Backspace', tone: 'extra' },
  { value: 'AC', label: 'Clear all', tone: 'clear' },
  { value: '÷', label: 'Divide', tone: 'operator' },
  { value: '7' }, { value: '8' }, { value: '9' },
  { value: '×', label: 'Multiply', tone: 'operator' },
  { value: '4' }, { value: '5' }, { value: '6' },
  { value: '−', label: 'Subtract', tone: 'operator' },
  { value: '1' }, { value: '2' }, { value: '3' },
  { value: '+', label: 'Add', tone: 'operator' },
  { value: '0', tone: 'zero' }, { value: '.', label: 'Decimal point' },
  { value: '=', label: 'Equals', tone: 'equals' },
]

export function Calculator() {
  return (
    <section className={styles.calculator} aria-label="Calculator keypad">
      <div className={styles.display}>
        <span className={styles.expression}>Design preview</span>
        <output aria-label="Calculator result" className={styles.result}>0</output>
      </div>
      <div className={styles.keypad}>
        {keys.map(({ value, label, tone }) => (
          <button
            key={value}
            type="button"
            aria-label={label}
            data-tone={tone}
            className={styles.key}
            disabled
          >
            {value}
          </button>
        ))}
      </div>
    </section>
  )
}
