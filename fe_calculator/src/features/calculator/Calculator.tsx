import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { calculate, calculationErrorMessage } from './operations'
import styles from './Calculator.module.scss'

// Match the UTF-8 byte limit in the Go expression API.
const maxExpressionBytes = 1024
const encoder = new TextEncoder()
const editKeys = ['AC', '⌫', '=']

const keys = [
  { value: '^', label: 'Power', tone: 'extra' },
  { value: '√', label: 'Square root', tone: 'extra', unavailable: true },
  { value: '%', label: 'Percent', tone: 'extra', unavailable: true },
  { value: '⌫', label: 'Backspace', tone: 'extra' },
  { value: 'AC', label: 'Clear all', tone: 'clear' },
  { value: '(', label: 'Open parenthesis', tone: 'extra' },
  { value: ')', label: 'Close parenthesis', tone: 'extra' },
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
  const [input, setInput] = useState('')
  const calculation = useMutation({ mutationFn: calculate })
  const errorMessage = calculation.error && calculationErrorMessage(calculation.error)
  const inputBytes = encoder.encode(input).length
  const remainingBytes = maxExpressionBytes - inputBytes

  function pressKey(value: string) {
    if (value === '=') {
      if (input && !calculation.isSuccess) calculation.mutate(input)
      return
    }

    const previousResult = calculation.data
    calculation.reset()

    if (value === 'AC') {
      setInput('')
    } else if (value === '⌫') {
      setInput((current) => current.slice(0, -1))
    } else if (previousResult !== undefined) {
      const leftOperand = previousResult < 0 ? `(${previousResult})` : String(previousResult)
      setInput('+−×÷^'.includes(value) ? `${leftOperand}${value}` : value)
    } else {
      setInput((current) => current + value)
    }
  }

  return (
    <section className={styles.calculator} aria-label="Calculator keypad">
      <div className={styles.display}>
        <div
          className={styles.displayContent}
          aria-label="Calculation display"
          role="region"
          tabIndex={0}
          ref={(element) => {
            if (element) element.scrollTop = element.scrollHeight
          }}
        >
          <div className={styles.expression}>{calculation.isSuccess ? `${input} =` : 'Input'}</div>
          <output
            aria-label={calculation.isSuccess ? 'Calculator result' : 'Calculator input'}
            className={styles.result}
          >
            {calculation.isSuccess ? calculation.data : input || '0'}
          </output>
        </div>
        {calculation.isPending && <span role="status">Calculating…</span>}
      </div>
      <div className={styles.inputFeedback}>
        <span role="status">
          {!calculation.isSuccess && remainingBytes === 0 && 'Input limit reached. Delete or clear to keep typing.'}
        </span>
        <span>{inputBytes} / {maxExpressionBytes} bytes</span>
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
            disabled={unavailable || calculation.isPending ||
              (!calculation.isSuccess && !editKeys.includes(value) && encoder.encode(value).length > remainingBytes)}
            onClick={() => pressKey(value)}
          >
            {value}
          </button>
        ))}
      </div>
    </section>
  )
}
