import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { Calculator } from './Calculator'

beforeEach(() => vi.stubEnv('VITE_API_BASE_URL', ''))
afterEach(() => {
  cleanup()
  vi.unstubAllEnvs()
  vi.unstubAllGlobals()
})

function renderCalculator() {
  const user = userEvent.setup()
  const queryClient = new QueryClient({ defaultOptions: { mutations: { gcTime: Infinity } } })

  render(
    <QueryClientProvider client={queryClient}>
      <Calculator />
    </QueryClientProvider>,
  )

  return async (...names: string[]) => {
    for (const name of names) {
      await user.click(screen.getByRole('button', { name }))
    }
  }
}

it('enters and edits keypad input without evaluating it until Equals is pressed', async () => {
  const fetchMock = vi.fn()
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()
  const display = screen.getByLabelText('Calculator input')

  expect(display.textContent).toBe('0')

  await press('1', 'Add', '2')
  expect(display.textContent).toBe('1+2')

  await press('Backspace')
  expect(display.textContent).toBe('1+')

  await press('Clear all')
  expect(display.textContent).toBe('0')
  expect(fetchMock).not.toHaveBeenCalled()
})

it('submits the expression once and shows the result returned by the API', async () => {
  let complete!: (response: Response) => void
  const fetchMock = vi.fn(() => new Promise<Response>((resolve) => { complete = resolve }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('1', '2', 'Decimal point', '5', 'Add', '3', 'Equals')
  await screen.findByText('Calculating…')
  await press('Equals', 'Clear all')

  expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/operations/evaluate', expect.objectContaining({
    method: 'POST',
    body: JSON.stringify({ expression: '12.5+3' }),
  }))
  expect(screen.getByLabelText('Calculator input').textContent).toBe('12.5+3')

  complete(Response.json({ result: 42 }))
  expect((await screen.findByLabelText('Calculator result')).textContent).toBe('42')
  expect(screen.getByText('12.5+3 =')).toBeDefined()

  await press('4')
  expect(screen.getByLabelText('Calculator input').textContent).toBe('4')
})

it.each([
  [10, 7],
  [0, -3],
  [-2, -5],
  [1e21, 1e21],
])('continues from the displayed result %s when an operator is pressed', async (previous, result) => {
  const fetchMock = vi.fn()
    .mockResolvedValueOnce(Response.json({ result: previous }))
    .mockResolvedValueOnce(Response.json({ result }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('5', 'Add', '5', 'Equals')
  expect((await screen.findByLabelText('Calculator result')).textContent).toBe(String(previous))

  const expression = previous < 0 ? `(${previous})−3` : `${previous}−3`
  await press('Subtract', '3')
  expect(screen.getByLabelText('Calculator input').textContent).toBe(expression)
  await press('Equals')

  expect((await screen.findByLabelText('Calculator result')).textContent).toBe(String(result))
  expect(fetchMock).toHaveBeenCalledTimes(2)
  expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/operations/evaluate', expect.objectContaining({
    method: 'POST',
    body: JSON.stringify({ expression }),
  }))
})

it('submits the subtraction expression in entry order and displays a negative result', async () => {
  const fetchMock = vi.fn().mockResolvedValue(Response.json({ result: -9.5 }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('3', 'Subtract', '1', '2', 'Decimal point', '5', 'Equals')

  expect((await screen.findByLabelText('Calculator result')).textContent).toBe('-9.5')
  expect(screen.getByText('3−12.5 =')).toBeDefined()
  expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/operations/evaluate', expect.objectContaining({
    method: 'POST',
    body: JSON.stringify({ expression: '3−12.5' }),
  }))
})

it('submits a multiplication expression and displays the decimal product from the API', async () => {
  const fetchMock = vi.fn().mockResolvedValue(Response.json({ result: 7.5 }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('2', 'Decimal point', '5', 'Multiply', '3', 'Equals')

  expect((await screen.findByLabelText('Calculator result')).textContent).toBe('7.5')
  expect(screen.getByText('2.5×3 =')).toBeDefined()
  expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/operations/evaluate', expect.objectContaining({
    method: 'POST',
    body: JSON.stringify({ expression: '2.5×3' }),
  }))
})

it('submits the division expression in entry order and displays the quotient from the API', async () => {
  const fetchMock = vi.fn().mockResolvedValue(Response.json({ result: 2.5 }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('5', 'Divide', '2', 'Equals')

  expect((await screen.findByLabelText('Calculator result')).textContent).toBe('2.5')
  expect(screen.getByText('5÷2 =')).toBeDefined()
  expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/operations/evaluate', expect.objectContaining({
    method: 'POST',
    body: JSON.stringify({ expression: '5÷2' }),
  }))
})

it('shows the API division-by-zero error and keeps the expression editable', async () => {
  const fetchMock = vi.fn().mockResolvedValue(Response.json({
    error: { code: 'DIVISION_BY_ZERO', message: 'Cannot divide by zero.' },
  }, { status: 422 }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('5', 'Divide', '0', 'Equals')

  expect((await screen.findByRole('alert')).textContent).toBe('Cannot divide by zero.')
  expect(screen.getByLabelText('Calculator input').textContent).toBe('5÷0')
  expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/operations/evaluate', expect.objectContaining({
    method: 'POST',
    body: JSON.stringify({ expression: '5÷0' }),
  }))

  await press('Backspace')
  expect(screen.queryByRole('alert')).toBeNull()
  expect(screen.getByLabelText('Calculator input').textContent).toBe('5÷')
})

it('submits a mixed expression as one request and displays the returned result', async () => {
  const fetchMock = vi.fn().mockResolvedValue(Response.json({ result: 14 }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('2', 'Add', '3', 'Multiply', '4')
  expect(fetchMock).not.toHaveBeenCalled()
  await press('Equals')

  expect((await screen.findByLabelText('Calculator result')).textContent).toBe('14')
  expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/operations/evaluate', expect.objectContaining({
    body: JSON.stringify({ expression: '2+3×4' }),
  }))
})

it('shows syntax errors from the backend and lets the user correct the expression', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(Response.json({
    error: { code: 'INVALID_EXPRESSION', message: 'Enter a valid expression.' },
  }, { status: 400 })))
  const press = renderCalculator()

  await press('2', 'Add', 'Equals')

  expect((await screen.findByRole('alert')).textContent).toBe('Enter a valid expression.')
  expect(screen.getByLabelText('Calculator input').textContent).toBe('2+')
  await press('Backspace')
  expect(screen.queryByRole('alert')).toBeNull()
  expect(screen.getByLabelText('Calculator input').textContent).toBe('2')
})

it('shows the API error and allows retrying the same expression', async () => {
  const fetchMock = vi.fn()
    .mockResolvedValueOnce(Response.json({
      error: { code: 'RESULT_OUT_OF_RANGE', message: 'The sum is out of range.' },
    }, { status: 422 }))
    .mockResolvedValueOnce(Response.json({ result: 3 }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('1', 'Add', '2', 'Equals')
  expect((await screen.findByRole('alert')).textContent).toBe('The sum is out of range.')
  expect(screen.getByLabelText('Calculator input').textContent).toBe('1+2')

  await press('Equals')
  expect((await screen.findByLabelText('Calculator result')).textContent).toBe('3')
  expect(screen.queryByRole('alert')).toBeNull()
  expect(fetchMock).toHaveBeenCalledTimes(2)
})

it.each([
  ['a network failure', () => Promise.reject(new TypeError('Network error'))],
  ['an invalid result', () => Promise.resolve(Response.json({ result: '3' }))],
  ['an invalid error envelope', () => Promise.resolve(Response.json({
    error: { code: false, message: 'Unvalidated message' },
  }, { status: 400 }))],
])('shows a generic message for %s', async (_scenario, respond) => {
  vi.stubGlobal('fetch', vi.fn(respond))
  const press = renderCalculator()

  await press('1', 'Add', '2', 'Equals')

  expect((await screen.findByRole('alert')).textContent).toBe('Couldn’t calculate the result. Try again.')
})

it('ignores Equals on an empty display and after a successful calculation', async () => {
  const fetchMock = vi.fn().mockResolvedValue(Response.json({ result: 3 }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('Equals')
  expect(fetchMock).not.toHaveBeenCalled()
  await press('1', 'Add', '2', 'Equals')
  await screen.findByLabelText('Calculator result')
  await press('Equals')
  expect(fetchMock).toHaveBeenCalledTimes(1)
})

it('shows a generic message for a plain-text HTTP error', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('Bad gateway', { status: 502 })))
  const press = renderCalculator()

  await press('1', 'Add', '2', 'Equals')
  expect((await screen.findByRole('alert')).textContent).toBe('Couldn’t calculate the result. Try again.')
})

it('counts multibyte operators and keeps editing and evaluation available at the byte limit', async () => {
  const fetchMock = vi.fn().mockResolvedValue(Response.json({ result: 0 }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()
  const zero = screen.getByRole('button', { name: '0' })
  const subtract = screen.getByRole('button', { name: 'Subtract' })

  fireEvent.click(zero)
  for (let i = 0; i < 255; i++) {
    fireEvent.click(subtract)
    fireEvent.click(zero)
  }
  await press('Add', '0')
  expect(screen.getByText('1023 / 1024 bytes')).toBeDefined()
  expect(subtract).toHaveProperty('disabled', true)

  await press('0')
  expect(screen.getByText('Input limit reached. Delete or clear to keep typing.')).toBeDefined()
  expect(zero).toHaveProperty('disabled', true)

  await press('Backspace')
  expect(zero).toHaveProperty('disabled', false)
  await press('0', 'Equals')
  expect((await screen.findByLabelText('Calculator result')).textContent).toBe('0')
  expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/operations/evaluate', expect.objectContaining({
    body: JSON.stringify({ expression: `0${'−0'.repeat(255)}+00` }),
  }))

  await press('Subtract', '1')
  expect(screen.getByLabelText('Calculator input').textContent).toBe('0−1')
  expect(screen.getByText('5 / 1024 bytes')).toBeDefined()
  await press('Clear all')
  expect(screen.getByText('0 / 1024 bytes')).toBeDefined()
})

it('submits a chain of powers and continues from the returned result', async () => {
  const fetchMock = vi.fn()
    .mockResolvedValueOnce(Response.json({ result: 512 }))
    .mockResolvedValueOnce(Response.json({ result: 262144 }))
  vi.stubGlobal('fetch', fetchMock)
  const press = renderCalculator()

  await press('2', 'Power', '3', 'Power', '2', 'Equals')
  expect((await screen.findByLabelText('Calculator result')).textContent).toBe('512')
  expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/operations/evaluate', expect.objectContaining({
    body: JSON.stringify({ expression: '2^3^2' }),
  }))
  await press('Power', '2', 'Equals')
  expect((await screen.findByLabelText('Calculator result')).textContent).toBe('262144')
  expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/operations/evaluate', expect.objectContaining({
    body: JSON.stringify({ expression: '512^2' }),
  }))
})

it('supports grouped negative bases and displays power domain errors', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(Response.json({
    error: { code: 'INVALID_DOMAIN', message: 'This power has no real-number result.' },
  }, { status: 422 })))
  const press = renderCalculator()

  await press('Open parenthesis', 'Subtract', '2', 'Close parenthesis', 'Power', '0', 'Decimal point', '5', 'Equals')
  expect((await screen.findByRole('alert')).textContent).toBe('This power has no real-number result.')
  expect(screen.getByLabelText('Calculator input').textContent).toBe('(−2)^0.5')
})
