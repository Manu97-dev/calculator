import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { HttpError, request } from './request'

beforeEach(() => vi.stubEnv('VITE_API_BASE_URL', ''))
afterEach(() => {
  vi.unstubAllEnvs()
  vi.unstubAllGlobals()
})

describe('request', () => {
  it('forwards request options and serializes the JSON body', async () => {
    const fetchMock = vi.fn().mockResolvedValue(Response.json({ created: true }))
    vi.stubGlobal('fetch', fetchMock)
    const { signal } = new AbortController()

    await expect(
      request('/items', {
        method: 'POST',
        signal,
        credentials: 'same-origin',
        headers: { 'X-Request-ID': 'example' },
        json: { value: 5 },
      }),
    ).resolves.toEqual({ created: true })

    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/items')
    expect(init).toMatchObject({
      method: 'POST',
      credentials: 'same-origin',
      body: '{"value":5}',
    })
    expect(init.signal).toBe(signal)
    expect(init.headers.get('X-Request-ID')).toBe('example')
    expect(init.headers.get('content-type')).toBe('application/json')
  })

  it('exposes the status and parsed body of a failed HTTP response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(
      Response.json({ error: 'Invalid input' }, { status: 400 }),
    ))

    const result = request('/items')
    await expect(result).rejects.toBeInstanceOf(HttpError)
    await expect(result).rejects.toMatchObject({
      message: 'HTTP 400',
      status: 400,
      body: { error: 'Invalid input' },
    })
  })

  it('uses the configured API base without duplicating path separators', async () => {
    vi.stubEnv('VITE_API_BASE_URL', 'https://api.example.com/api/')
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(request('/items')).resolves.toBeUndefined()
    expect(fetchMock.mock.calls[0][0]).toBe('https://api.example.com/api/items')
  })

  it('keeps the HTTP status when an error response contains malformed JSON', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('Bad gateway', {
      status: 502,
      headers: { 'content-type': 'application/json' },
    })))

    await expect(request('/items')).rejects.toMatchObject({
      status: 502,
      body: 'Bad gateway',
    })
  })
})
