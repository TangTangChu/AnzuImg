import { useApi } from '~/composables/useApi'
import { useStepUp } from '~/composables/useStepUp'
import { parseApiError } from '~/utils/api-error'
import type { SettingsResponse } from '~/types/settings'

export const useSettings = () => {
    const { apiUrl } = useApi()
    const stepUp = useStepUp()

    const callWithStepUp = async <T>(fn: () => Promise<T>): Promise<T | null> => {
        try {
            return await fn()
        } catch (error: any) {
            const parsed = parseApiError(error, 'request failed')
            if (parsed.code === 'step_up_required') {
                const ok = await stepUp.request(error?.data?.available_methods, error?.data?.max_age_seconds)
                if (!ok) return null
                try {
                    return await fn()
                } catch {
                    return null
                }
            }
            throw error
        }
    }

    const get = async (): Promise<SettingsResponse | null> => {
        try {
            return await $fetch<SettingsResponse>(apiUrl('/api/v1/settings'))
        } catch {
            return null
        }
    }

    const patch = async (values: Record<string, string>): Promise<boolean> => {
        const res = await callWithStepUp(async () =>
            await $fetch<{ message: string }>(apiUrl('/api/v1/settings'), {
                method: 'PATCH',
                body: { values },
            }),
        )
        return res !== null
    }

    const reset = async (keys: string[]): Promise<boolean> => {
        const res = await callWithStepUp(async () =>
            await $fetch<{ message: string }>(apiUrl('/api/v1/settings/reset'), {
                method: 'POST',
                body: { keys },
            }),
        )
        return res !== null
    }

    return { get, patch, reset }
}
