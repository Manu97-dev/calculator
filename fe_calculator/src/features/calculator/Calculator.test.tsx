import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'
import { Calculator } from './Calculator'

afterEach(cleanup)

it('enters and edits keypad input without evaluating it', async () => {
  const user = userEvent.setup()
  render(<Calculator />)

  const display = screen.getByLabelText('Calculator input')
  expect(display.textContent).toBe('0')

  for (const name of ['1', 'Add', '2', 'Equals']) {
    await user.click(screen.getByRole('button', { name }))
  }
  expect(display.textContent).toBe('1+2=')

  await user.click(screen.getByRole('button', { name: 'Backspace' }))
  expect(display.textContent).toBe('1+2')

  await user.click(screen.getByRole('button', { name: 'Clear all' }))
  expect(display.textContent).toBe('0')
})
