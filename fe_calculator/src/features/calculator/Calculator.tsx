import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { addOperands, additionErrorMessage } from './addition'
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
  { value: '−', label: 'Subtract', tone: 'operator', unavailable: true },
  { value: '1' }, { value: '2' }, { value: '3' },
  { value: '+', label: 'Add', tone: 'operator' },
  { value: '0', tone: 'zero' }, { value: '.', label: 'Decimal point' },
  { value: '=', label: 'Equals', tone: 'equals' },
]

const additionExpression = /^(\d+\.?\d*|\.\d+)\+(\d+\.?\d*|\.\d+)$/

export function Calculator() {
  const [input, setInput] = useState('')
  const [inputError, setInputError] = useState('')
  const addition = useMutation({ mutationFn: addOperands })
  const errorMessage = inputError || (addition.error && additionErrorMessage(addition.error))

  function pressKey(value: string) {
    if (addition.isPending) return

    if (value === '=') {
      if (addition.isSuccess) return

      const operands = additionExpression.exec(input)
      if (!operands) {
        setInputError('Enter two numbers with one plus sign, such as 12.5+3.')
        return
      }

      const left = Number(operands[1])
      const right = Number(operands[2])
      if (!Number.isFinite(left) || !Number.isFinite(right)) {
        setInputError('Enter smaller numbers.')
        return
      }

      setInputError('')
      addition.mutate({ left, right })
      return
    }

    const startFresh = addition.isSuccess
    setInputError('')
    addition.reset()

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
        <span className={styles.expression}>{addition.isSuccess ? `${input} =` : 'Input'}</span>
        <output
          aria-label={addition.isSuccess ? 'Calculator result' : 'Calculator input'}
          className={styles.result}
        >
          {addition.isSuccess ? addition.data : input || '0'}
        </output>
        {addition.isPending && <span role="status">Calculating…</span>}
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
            disabled={unavailable || addition.isPending}
            onClick={() => pressKey(value)}
          >
            {value}
          </button>
        ))}
      </div>
    </section>
  )
}
