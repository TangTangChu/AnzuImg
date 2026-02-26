import { startAuthentication, startRegistration } from '@simplewebauthn/browser';
import type { APIToken, APITokenLogListResponse, CreateTokenResponse } from '~/types/api_token';
import type { PasskeyCredential } from '~/types/passkey';
import type { SecurityLogListResponse } from '~/types/security_log';
import { navigateTo, ref, useCookie } from '#imports';
import { useAuthState } from '~/composables/useAuthState';
import { useApi } from '~/composables/useApi';
import { parseApiError, type ParsedApiError } from '~/utils/api-error';
interface PasskeyBeginResponse {
    session_id?: string
    assertion?: {
        publicKey?: unknown
    }
    creation?: {
        publicKey?: unknown
    }
}

export const useAuth = () => {
    const token = useCookie<string | null>('auth_token');
    const authState = useAuthState();
    const { apiUrl } = useApi();
    const lastApiError = ref<ParsedApiError | null>(null)

    const captureApiError = (context: string, error: any) => {
        const parsed = parseApiError(error, context)
        lastApiError.value = parsed
        console.error(`${context}: ${parsed.displayMessage}`, {
            code: parsed.code,
            requestId: parsed.requestId,
        })
        return parsed
    }

    const clearLastApiError = () => {
        lastApiError.value = null
    }

    const getLastApiErrorDisplay = (fallbackMessage: string) => {
        return lastApiError.value?.displayMessage || fallbackMessage
    }

    const login = async (password: string) => {
        try {
            await $fetch<{ token: string }>(apiUrl('/api/v1/auth/login'), {
                method: 'POST',
                body: { password }
            });
            clearLastApiError()
            token.value = null;
            authState.setAuthenticated(true)
            return true;
        } catch (error: any) {
            captureApiError('Login failed', error)
            authState.resetAuth()
            return false;
        }
    };

    const checkInit = async (options?: { headers?: HeadersInit; throwOnError?: boolean }) => {
        try {
            const data = await $fetch<{ initialized: boolean }>(apiUrl('/api/v1/auth/status'), {
                headers: options?.headers,
            });
            clearLastApiError()
            const initialized = !!data.initialized;
            authState.setInitialized(initialized)
            return initialized;
        } catch (error: any) {
            captureApiError('Check init failed', error)
            if (options?.throwOnError) {
                throw error
            }
            return false;
        }
    };

    const setup = async (password: string, setupToken?: string) => {
        try {
            await $fetch(apiUrl('/api/v1/auth/setup'), {
                method: 'POST',
                body: {
                    password,
                    setup_token: setupToken
                },
            });
            clearLastApiError()
            return true;
        } catch (error: any) {
            captureApiError('Setup failed', error)
            return false;
        }
    };

    const logout = async () => {
        try {
            await $fetch(apiUrl('/api/v1/auth/logout'), { method: 'POST' })
            clearLastApiError()
        } catch (error: any) {
            console.warn('Logout request failed', error)
        }
        token.value = null;
        authState.resetAuth()
        navigateTo('/login');
    };

    const loginWithPasskey = async () => {
        try {
            const beginData = await $fetch<PasskeyBeginResponse>(apiUrl('/api/v1/auth/passkey/login/begin'));
            const publicKey = beginData.assertion?.publicKey;
            const sessionId = beginData.session_id;

            if (!publicKey) {
                console.error('No publicKey found in beginData.assertion');
                return false;
            }

            if (!sessionId) {
                console.error('No session_id found in beginData.assertion');
                return false;
            }

            let authResp;
            try {
                authResp = await startAuthentication({ optionsJSON: publicKey as any });
            } catch (authenticationError) {
                console.error('startAuthentication failed:', authenticationError);
                return false;
            }
            await $fetch<{ token: string }>(apiUrl('/api/v1/auth/passkey/login/finish'), {
                method: 'POST',
                body: authResp,
                headers: {
                    'X-Session-ID': sessionId
                }
            });

            clearLastApiError()
            token.value = null;
            authState.setAuthenticated(true)
            return true;
        } catch (error: any) {
            captureApiError('Passkey login failed', error)
            authState.resetAuth()
            return false;
        }
    };

    const registerPasskey = async () => {
        try {
            const beginData = await $fetch<PasskeyBeginResponse>(apiUrl('/api/v1/auth/passkey/register/begin'));
            const publicKey = beginData.creation?.publicKey;
            const sessionId = beginData.session_id;

            if (!publicKey) {
                console.error('No publicKey found in beginData');
                return false;
            }

            if (!sessionId) {
                console.error('No session_id found in beginData');
                return false;
            }

            let authResp;
            try {
                authResp = await startRegistration({ optionsJSON: publicKey as any });
            } catch (registrationError) {
                console.error('startRegistration failed:', registrationError);
                return false;
            }
            await $fetch(apiUrl('/api/v1/auth/passkey/register/finish'), {
                method: 'POST',
                body: authResp,
                headers: {
                    'X-Session-ID': sessionId
                }
            });

            clearLastApiError()
            return true;
        } catch (error: any) {
            captureApiError('Passkey registration failed', error)
            return false;
        }
    };

    // 修改密码
    const changePassword = async (currentPassword: string, newPassword: string) => {
        try {
            await $fetch(apiUrl('/api/v1/auth/change-password'), {
                method: 'POST',
                body: {
                    current_password: currentPassword,
                    new_password: newPassword
                }
            });
            clearLastApiError()
            return true;
        } catch (error: any) {
            captureApiError('Change password failed', error)
            return false;
        }
    };

    // 获取PassKey列表
    const listPasskeys = async () => {
        try {
            const data = await $fetch<{ credentials: PasskeyCredential[], count: number }>(apiUrl('/api/v1/auth/passkeys'));
            clearLastApiError()
            return data.credentials;
        } catch (error: any) {
            captureApiError('List passkeys failed', error)
            return [];
        }
    };

    // 删除PassKey
    const deletePasskey = async (credentialId: string) => {
        try {
            await $fetch(apiUrl(`/api/v1/auth/passkeys/${credentialId}`), {
                method: 'DELETE',
            });
            clearLastApiError()
            return true;
        } catch (error: any) {
            try {
                await $fetch(apiUrl(`/api/v1/auth/passkeys/${credentialId}/delete`), {
                    method: 'POST',
                });
                clearLastApiError()
                return true;
            } catch (fallbackError: any) {
                captureApiError('Delete passkey failed', fallbackError)
                return false;
            }
        }
    };

    // 检查是否有PassKey
    const checkPasskeyExists = async () => {
        try {
            const data = await $fetch<{ has_passkey: boolean }>(apiUrl('/api/v1/auth/passkeys/check'));
            clearLastApiError()
            return data.has_passkey;
        } catch (error: any) {
            captureApiError('Check passkey exists failed', error)
            return false;
        }
    };

    // API Token Management
    const createAPIToken = async (name: string, ipAllowlist: string[], tokenType: string) => {
        try {
            const data = await $fetch<CreateTokenResponse>(apiUrl('/api/v1/auth/tokens'), {
                method: 'POST',
                body: { name, ip_allowlist: ipAllowlist, token_type: tokenType }
            });
            clearLastApiError()
            return data;
        } catch (error: any) {
            captureApiError('Create API token failed', error)
            return null;
        }
    };

    const listAPITokens = async () => {
        try {
            const data = await $fetch<APIToken[]>(apiUrl('/api/v1/auth/tokens'));
            clearLastApiError()
            return data;
        } catch (error: any) {
            captureApiError('List API tokens failed', error)
            return [];
        }
    };

    const deleteAPIToken = async (id: number) => {
        try {
            await $fetch(apiUrl(`/api/v1/auth/tokens/${id}`), {
                method: 'DELETE',
            });
            clearLastApiError()
            return true;
        } catch (error: any) {
            try {
                await $fetch(apiUrl(`/api/v1/auth/tokens/${id}/delete`), {
                    method: 'POST',
                });
                clearLastApiError()
                return true;
            } catch (fallbackError: any) {
                captureApiError('Delete API token failed', fallbackError)
                return false;
            }
        }
    };

    const listAPITokenLogs = async (page = 1, pageSize = 20, search = "", startDate = "", endDate = "", type = "") => {
        try {
            const data = await $fetch<APITokenLogListResponse>(apiUrl('/api/v1/auth/tokens/logs'), {
                query: { page, page_size: pageSize, search, start_date: startDate, end_date: endDate, type }
            });
            clearLastApiError()
            return data;
        } catch (error: any) {
            captureApiError('List API token logs failed', error)
            return { data: [], total: 0, page, size: pageSize } as APITokenLogListResponse;
        }
    };

    const cleanupAPITokenLogs = async (days: number) => {
        try {
            const data = await $fetch<{ deleted: number; cutoff: string }>(apiUrl('/api/v1/auth/tokens/logs'), {
                method: 'DELETE',
                query: { days }
            });
            clearLastApiError()
            return data;
        } catch (error: any) {
            try {
                const data = await $fetch<{ deleted: number; cutoff: string }>(apiUrl('/api/v1/auth/tokens/logs/cleanup'), {
                    method: 'POST',
                    body: { days }
                });
                clearLastApiError()
                return data;
            } catch (fallbackError: any) {
                captureApiError('Cleanup API token logs failed', fallbackError)
                return null;
            }
        }
    };

    const listSecurityLogs = async (page = 1, pageSize = 20, failedOnly = false, search = "", startDate = "", endDate = "", type = "") => {
        try {
            const data = await $fetch<SecurityLogListResponse>(apiUrl('/api/v1/auth/security/logs'), {
                query: {
                    page,
                    page_size: pageSize,
                    failed_only: failedOnly,
                    search,
                    start_date: startDate,
                    end_date: endDate,
                    type
                }
            });
            clearLastApiError()
            return data;
        } catch (error: any) {
            captureApiError('List security logs failed', error)
            return { data: [], total: 0, page, size: pageSize } as SecurityLogListResponse;
        }
    };

    return {
        token,
        login,
        loginWithPasskey,
        registerPasskey,
        logout,
        checkInit,
        setup,
        changePassword,
        listPasskeys,
        deletePasskey,
        checkPasskeyExists,
        createAPIToken,
        listAPITokens,
        deleteAPIToken,
        listAPITokenLogs,
        cleanupAPITokenLogs,
        listSecurityLogs,
        getLastApiErrorDisplay,
        lastApiError,
    };
}
