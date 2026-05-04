import { ref, onScopeDispose } from 'vue'
import { useApi } from '~/composables/useApi'
import type { AppLog } from '~/types/app_log'

export interface LogStreamOptions {
    level?: string
    module?: string
    onLog: (log: AppLog) => void
}

export const useLogStream = () => {
    const { apiUrl } = useApi()
    const connected = ref(false)
    let es: EventSource | null = null

    const stop = () => {
        if (es) {
            es.close()
            es = null
        }
        connected.value = false
    }

    const start = (opts: LogStreamOptions) => {
        stop()
        const params = new URLSearchParams({
            source: 'app',
            level: opts.level ?? 'info',
        })
        if (opts.module) params.set('module', opts.module)
        const url = apiUrl(`/api/v1/logs/stream?${params.toString()}`)
        es = new EventSource(url, { withCredentials: true })
        es.addEventListener('log', (ev: MessageEvent) => {
            try {
                opts.onLog(JSON.parse(ev.data) as AppLog)
            } catch {
                /* ignore */
            }
        })
        es.onopen = () => {
            connected.value = true
        }
        es.onerror = () => {
            connected.value = false
        }
    }

    onScopeDispose(stop)
    return { start, stop, connected }
}
