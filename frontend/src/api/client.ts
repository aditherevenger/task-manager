export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function readErrorMessage(res: Response): Promise<string> {
  const contentType = res.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    try {
      const body = (await res.json()) as unknown
      if (body && typeof body === 'object' && 'error' in body && typeof (body as any).error === 'string') {
        return (body as any).error
      }
      return JSON.stringify(body)
    } catch {
      return res.statusText
    }
  }

  try {
    return await res.text()
  } catch {
    return res.statusText
  }
}

export async function api<T>(
  path: string,
  init?: RequestInit & { json?: unknown },
): Promise<T> {
  const headers = new Headers(init?.headers)
  headers.set('accept', 'application/json')

  let body: BodyInit | undefined
  if (init && 'json' in init) {
    headers.set('content-type', 'application/json')
    body = JSON.stringify(init.json)
  } else {
    const b = init?.body
    body = b === null ? undefined : b
  }

  const res = await fetch(path, {
    ...init,
    headers,
    body,
  })

  if (!res.ok) {
    const msg = await readErrorMessage(res)
    throw new ApiError(res.status, msg)
  }

  if (res.status === 204) {
    return undefined as T
  }

  return (await res.json()) as T
}
