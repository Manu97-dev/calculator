export type RequestOptions = RequestInit & { json?: unknown }

export class HttpError extends Error {
  readonly status: number
  readonly body: unknown

  constructor(response: Response, body: unknown) {
    super(`HTTP ${response.status}${response.statusText ? ` ${response.statusText}` : ''}`)
    this.name = 'HttpError'
    this.status = response.status
    this.body = body
  }
}

export async function request(path: string, options: RequestOptions = {}): Promise<unknown> {
  const { json, headers: requestHeaders, ...init } = options
  const headers = new Headers(requestHeaders)

  if (json !== undefined) {
    if (init.body != null) {
      throw new TypeError('Provide either json or body when making a request.')
    }

    if (!headers.has('content-type')) {
      headers.set('content-type', 'application/json')
    }

    init.body = JSON.stringify(json)
  }

  const apiBase = import.meta.env.VITE_API_BASE_URL?.trim() || '/api'
  const url = `${apiBase.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`
  const response = await fetch(url, { ...init, headers })
  const contentType = response.headers.get('content-type')?.split(';', 1)[0].trim().toLowerCase()
  const text = await response.text()
  let body: unknown = text || undefined

  if (text && (contentType === 'application/json' || contentType?.endsWith('+json'))) {
    try {
      body = JSON.parse(text) as unknown
    } catch (error) {
      if (response.ok) throw error
    }
  }

  if (!response.ok) throw new HttpError(response, body)
  return body
}
