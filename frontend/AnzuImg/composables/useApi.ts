import { useRequestURL, useRuntimeConfig } from '#imports'

const normalizePrefix = (prefix: string): string => {
    const trimmed = (prefix || '').trim()
    if (trimmed === '' || trimmed === '/') return ''
    const withLeading = trimmed.startsWith('/') ? trimmed : `/${trimmed}`
    return withLeading.replace(/\/+$/, '')
}

const joinPath = (prefix: string, path: string): string => {
    const normalizedPrefix = normalizePrefix(prefix)
    const normalizedPath = path.startsWith('/') ? path : `/${path}`
    return `${normalizedPrefix}${normalizedPath}`
}

const getOrigin = (): string => {
    if (import.meta.server) {
        return useRequestURL().origin
    }
    if (typeof window !== 'undefined' && window.location?.origin) {
        return window.location.origin
    }
    return ''
}

export const useApi = () => {
    const config = useRuntimeConfig()
    const raw = (config.public as any)?.apiPrefix
    const apiPrefix = normalizePrefix(raw ?? '/kotori')

    const useAbsoluteUrl = ((config.public as any)?.apiUseAbsoluteUrl ?? true) !== false

    const apiUrl = (path: string) => {
        const urlPath = joinPath(apiPrefix, path)
        if (!useAbsoluteUrl) return urlPath

        const origin = getOrigin()
        return origin ? new URL(urlPath, origin).toString() : urlPath
    }

    return {
        apiPrefix,
        apiUrl,
    }
}
