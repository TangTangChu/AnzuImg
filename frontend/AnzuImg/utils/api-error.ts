export interface ParsedApiError {
    code?: string
    message: string
    requestId?: string
    displayMessage: string
}

const getHeaderValue = (headers: unknown, key: string): string | undefined => {
    if (!headers) return undefined

    if (typeof (headers as { get?: unknown }).get === 'function') {
        const value = (headers as { get: (name: string) => string | null }).get(key)
        return typeof value === 'string' && value ? value : undefined
    }

    if (typeof headers === 'object') {
        const record = headers as Record<string, unknown>
        const direct = record[key]
        if (typeof direct === 'string' && direct) return direct

        const lower = record[key.toLowerCase()]
        if (typeof lower === 'string' && lower) return lower
    }

    return undefined
}

export const parseApiError = (error: any, fallbackMessage: string): ParsedApiError => {
    const data = error?.data ?? {}
    const responseHeaders = error?.response?.headers

    const requestIdFromHeader =
        getHeaderValue(responseHeaders, 'x-request-id') ||
        getHeaderValue(responseHeaders, 'X-Request-ID')

    const message =
        (typeof data?.message === 'string' && data.message) ||
        (typeof error?.message === 'string' && error.message) ||
        fallbackMessage

    const requestId =
        (typeof data?.request_id === 'string' && data.request_id) ||
        requestIdFromHeader ||
        undefined

    const displayMessage = requestId ? `${message} (request_id: ${requestId})` : message

    return {
        code: typeof data?.code === 'string' ? data.code : undefined,
        message,
        requestId,
        displayMessage,
    }
}
