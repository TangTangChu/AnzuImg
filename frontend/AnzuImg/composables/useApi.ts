import { useRuntimeConfig } from '#imports'

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

export const useApi = () => {
    const config = useRuntimeConfig()
    const raw = (config.public as any)?.apiPrefix
    const apiPrefix = normalizePrefix(raw ?? '/korori')

    const apiUrl = (path: string) => joinPath(apiPrefix, path)

    return {
        apiPrefix,
        apiUrl,
    }
}
