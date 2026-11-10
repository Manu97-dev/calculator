import { HttpError, request } from '../../shared/http/request'

export async function calculate(expression: string): Promise<number> {
  const response = await request('/operations/evaluate', {
    method: 'POST',
    json: { expression },
  })

  if (
    typeof response === 'object' && response !== null && 'result' in response &&
    typeof response.result === 'number' && Number.isFinite(response.result)
  ) {
    return response.result
  }

  throw new Error('Invalid calculation response')
}

export function calculationErrorMessage(error: Error): string {
  if (error instanceof HttpError) {
    const body = error.body
    if (typeof body === 'object' && body !== null && 'error' in body) {
      const detail = body.error
      if (
        typeof detail === 'object' && detail !== null &&
        'code' in detail && typeof detail.code === 'string' && detail.code.trim() &&
        'message' in detail && typeof detail.message === 'string' && detail.message.trim()
      ) {
        return detail.message
      }
    }
  }

  return 'Couldn’t calculate the result. Try again.'
}
