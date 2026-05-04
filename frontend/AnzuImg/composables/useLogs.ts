import { useApi } from '~/composables/useApi'
import { useStepUp } from '~/composables/useStepUp'
import { parseApiError } from '~/utils/api-error'
import type { AppLogListResponse, LogFilter, LogSource } from '~/types/app_log'
import type { SecurityLogListResponse } from '~/types/security_log'
import type { APITokenLogListResponse } from '~/types/api_token'

const buildQuery = (page: number, size: number, filter: LogFilter) => ({
    page,
    page_size: size,
    search: filter.search ?? '',
    level: filter.level ?? '',
    module: filter.module ?? '',
    ip: filter.ip ?? '',
    action: filter.action ?? '',
    start_date: filter.start_date ?? '',
    end_date: filter.end_date ?? '',
    failed_only: filter.failed_only ?? false,
})

export const useLogs = () => {
    const { apiUrl } = useApi()
    const stepUp = useStepUp()

    const listApp = async (page = 1, size = 50, filter: LogFilter = {}) => {
        try {
            return await $fetch<AppLogListResponse>(apiUrl('/api/v1/logs/app'), {
                query: buildQuery(page, size, filter),
            })
        } catch {
            return { data: [], total: 0, page, size } as AppLogListResponse
        }
    }

    const listSecurity = async (page = 1, size = 50, filter: LogFilter = {}) => {
        try {
            return await $fetch<SecurityLogListResponse>(apiUrl('/api/v1/logs/security'), {
                query: buildQuery(page, size, filter),
            })
        } catch {
            return { data: [], total: 0, page, size } as SecurityLogListResponse
        }
    }

    const listToken = async (page = 1, size = 50, filter: LogFilter = {}) => {
        try {
            return await $fetch<APITokenLogListResponse>(apiUrl('/api/v1/logs/token'), {
                query: buildQuery(page, size, filter),
            })
        } catch {
            return { data: [], total: 0, page, size } as APITokenLogListResponse
        }
    }

    const cleanup = async (source: LogSource, days: number): Promise<{ deleted: number; source: LogSource } | null> => {
        const run = async () =>
            await $fetch<{ deleted: number; source: LogSource }>(
                apiUrl(`/api/v1/logs/${source}`),
                { method: 'DELETE', query: { days } },
            )
        try {
            return await run()
        } catch (error: any) {
            const parsed = parseApiError(error, 'cleanup failed')
            if (parsed.code === 'step_up_required') {
                const ok = await stepUp.request(error?.data?.available_methods, error?.data?.max_age_seconds)
                if (!ok) return null
                try {
                    return await run()
                } catch {
                    try {
                        return await $fetch<{ deleted: number; source: LogSource }>(
                            apiUrl(`/api/v1/logs/${source}/cleanup`),
                            { method: 'POST', query: { days } },
                        )
                    } catch {
                        return null
                    }
                }
            }
            return null
        }
    }

    const exportUrl = (source: LogSource, format: 'csv' | 'json', filter: LogFilter, limit = 10000) => {
        const qs = new URLSearchParams({
            source,
            format,
            limit: String(limit),
            ...Object.fromEntries(
                Object.entries(buildQuery(1, limit, filter)).map(([k, v]) => [k, String(v)]),
            ),
        })
        return apiUrl(`/api/v1/logs/export?${qs.toString()}`)
    }

    return { listApp, listSecurity, listToken, cleanup, exportUrl }
}
