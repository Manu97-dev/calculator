import { screen } from '@testing-library/react'
import { expect, it } from 'vitest'

it('mounts the calculator with its production providers', async () => {
  document.body.innerHTML = '<div id="root"></div>'
  await import('./main')

  expect(await screen.findByRole('heading', { name: 'Calculator' })).toBeDefined()
  expect(screen.getByLabelText('Calculator input').textContent).toBe('0')
})
