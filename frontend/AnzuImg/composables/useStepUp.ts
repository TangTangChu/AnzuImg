import { startAuthentication } from '@simplewebauthn/browser'
import { useApi } from '~/composables/useApi'
import { useDialog, isDialogDismissedError } from '~/composables/useDialog'
import StepUpModal from '~/components/StepUpModal.vue'

interface PasskeyBeginResponse {
    session_id?: string
    assertion?: { publicKey?: unknown }
}

export interface StepUpHandle {
    available: string[]
    maxAgeSeconds: number
    /** 调用方在收到 step_up_required 错误时通过这个解析出 available_methods */
    fromError: (error: any) => boolean
    /** 弹出 step-up 模态,返回是否通过 */
    request: (available?: string[], maxAge?: number) => Promise<boolean>
    stepUpWithPassword: (password: string) => Promise<boolean>
    stepUpWithPasskey: () => Promise<boolean>
}

export const useStepUp = (): StepUpHandle => {
    const { apiUrl } = useApi()
    const { custom } = useDialog()
    const { t } = useI18n()

    const stepUpWithPassword = async (password: string): Promise<boolean> => {
        try {
            await $fetch(apiUrl('/api/v1/auth/step-up/password'), {
                method: 'POST',
                body: { password },
            })
            return true
        } catch {
            return false
        }
    }

    const stepUpWithPasskey = async (): Promise<boolean> => {
        try {
            const begin = await $fetch<PasskeyBeginResponse>(
                apiUrl('/api/v1/auth/step-up/passkey/begin'),
            )
            const publicKey = begin.assertion?.publicKey
            const sessionId = begin.session_id
            if (!publicKey || !sessionId) return false
            const resp = await startAuthentication({ optionsJSON: publicKey as any })
            await $fetch(apiUrl('/api/v1/auth/step-up/passkey/finish'), {
                method: 'POST',
                body: resp,
                headers: { 'X-Session-ID': sessionId },
            })
            return true
        } catch {
            return false
        }
    }

    const fromError = (error: any): boolean => {
        const code = error?.data?.code
        return code === 'step_up_required'
    }

    const request = async (
        available?: string[],
        maxAge?: number,
    ): Promise<boolean> => {
        return new Promise<boolean>((resolve) => {
            let resolved = false
            const onResult = (ok: boolean) => {
                if (resolved) return
                resolved = true
                resolve(ok)
            }
            custom(
                StepUpModal,
                {
                    available: available && available.length > 0 ? available : ['password', 'passkey'],
                    maxAgeSeconds: maxAge ?? 120,
                    runPassword: stepUpWithPassword,
                    runPasskey: stepUpWithPasskey,
                    onResult,
                },
                {
                    title: t('auth.stepUp.title'),
                    closeOnClickOutside: false,
                    closeOnEsc: true,
                    persistent: true,
                },
            )
                .catch((e) => {
                    if (isDialogDismissedError(e)) {
                        if (!resolved) {
                            resolved = true
                            resolve(false)
                        }
                    }
                })
                .then(() => {
                    if (!resolved) {
                        resolved = true
                        resolve(false)
                    }
                })
        })
    }

    return {
        available: ['password', 'passkey'],
        maxAgeSeconds: 120,
        fromError,
        request,
        stepUpWithPassword,
        stepUpWithPasskey,
    }
}
