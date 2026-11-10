import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { calculate, calculationErrorMessage } from './operations'
import styles from './Calculator.module.scss'

const keys = [
  { value: '^', label: 'Power', tone: 'extra', unavailable: true },
  { value: '√', label: 'Square root', tone: 'extra', unavailable: true },
  { value: '%', label: 'Percent', tone: 'extra', unavailable: true },
  { value: '⌫', label: 'Backspace', tone: 'extra' },
  { value: 'AC', label: 'Clear all', tone: 'clear' },
  { value: '÷', label: 'Divide', tone: 'operator', unavailable: true },
  { value: '7' }, { value: '8' }, { value: '9' },
  { value: '×', label: 'Multiply', tone: 'operator', unavailable: true },
  { value: '4' }, { value: '5' }, { value: '6' },
  { value: '−', label: 'Subtract', tone: 'operator' },
  { value: '1' }, { value: '2' }, { value: '3' },
  { value: '+', label: 'Add', tone: 'operator' },
  { value: '0', tone: 'zero' }, { value: '.', label: 'Decimal point' },
  { value: '=', label: 'Equals', tone: 'equals' },
]

const expressionPattern = /^(\d+\.?\d*|\.\d+)([+−])(\d+\.?\d*|\.\d+)$/

export function Calculator() {
  const [input, setInput] = useState('')
  const [inputError, setInputError] = useState('')
  const calculation = useMutation({ mutationFn: calculate })
  const errorMessage = inputError || (calculation.error && calculationErrorMessage(calculation.error))

  function pressKey(value: string) {
    if (calculation.isPending) return

    if (value === '=') {
      if (calculation.isSuccess) return

      const expression = expressionPattern.exec(input)
      if (!expression) {
        setInputError('Enter two numbers with one + or − between them.')
        return
      }

      const [, leftInput, operator, rightInput] = expression
      const left = Number(leftInput)
      const right = Number(rightInput)
      if (!Number.isFinite(left) || !Number.isFinite(right)) {
        setInputError('Enter smaller numbers.')
        return
      }

      setInputError('')
      calculation.mutate({ operation: operator === '+' ? 'addition' : 'subtraction', left, right })
      return
    }

    const startFresh = calculation.isSuccess
    setInputError('')
    calculation.reset()

    if (value === 'AC') {
      setInput('')
    } else if (value === '⌫') {
      setInput((current) => current.slice(0, -1))
    } else {
      setInput((current) => (startFresh ? '' : current) + value)
    }
  }

  return (
    <section className={styles.calculator} aria-label="Calculator keypad">
      <div className={styles.display}>
        <span className={styles.expression}>{calculation.isSuccess ? `${input} =` : 'Input'}</span>
        <output
          aria-label={calculation.isSuccess ? 'Calculator result' : 'Calculator input'}
          className={styles.result}
        >
          {calculation.isSuccess ? calculation.data : input || '0'}
        </output>
        {calculation.isPending && <span role="status">Calculating…</span>}
      </div>
      {errorMessage && <p className={styles.error} role="alert">{errorMessage}</p>}
      <div className={styles.keypad}>
        {keys.map(({ value, label, tone, unavailable }) => (
          <button
            key={value}
            type="button"
            aria-label={label}
            data-tone={tone}
            className={styles.key}
            disabled={unavailable || calculation.isPending}
            onClick={() => pressKey(value)}
          >
            {value}
          </button>
        ))}
      </div>
    </section>
  )
}
