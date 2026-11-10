import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, useLocation } from 'react-router-dom'
import { afterEach, expect, it } from 'vitest'
import { AppRoutes } from './AppRoutes'

afterEach(cleanup)

function CurrentPath() {
  return <output aria-label="Current route">{useLocation().pathname}</output>
}

it('recovers from an unknown route by navigating back to the home page', async () => {
  const user = userEvent.setup()

  render(
    <MemoryRouter initialEntries={['/unknown']}>
      <CurrentPath />
      <AppRoutes />
    </MemoryRouter>,
  )

  const backLink = screen.getByRole('link', { name: 'Back to home' })
  expect(backLink.getAttribute('href')).toBe('/')
  await user.click(backLink)

  expect(screen.getByLabelText('Current route').textContent).toBe('/')
  expect(screen.queryByRole('link', { name: 'Back to home' })).toBeNull()
  expect(screen.getByRole('heading', { name: 'Calculator' })).toBeDefined()
})
