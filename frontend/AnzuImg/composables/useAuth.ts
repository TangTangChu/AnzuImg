import { startAuthentication, startRegistration } from '@simplewebauthn/browser';
import type { APIToken, CreateTokenResponse } from '~/types/api_token';
import { navigateTo, useCookie } from '#imports';
import { useAuthState } from '~/composables/useAuthState';

export const useAuth = () => {
    const token = useCookie<string | null>('auth_token');
    const authState = useAuthState();

    const login = async (password: string) => {
        try {
            const data = await $fetch<{ token: string }>('/api/v1/auth/login', {
                method: 'POST',
                body: { password }
            });
            token.value = null;
            authState.setAuthenticated(true)
            return true;
        } catch (error: any) {
            console.error('Login failed', error);
            authState.resetAuth()
            return false;
        }
    };

    const checkInit = async () => {
        try {
            const data = await $fetch<{ initialized: boolean }>('/api/v1/auth/status');
            return data.initialized;
        } catch (error: any) {
            console.error('Check init failed', error);
            return false;
        }
    };

    const setup = async (password: string, setupToken?: string) => {
        try {
            await $fetch('/api/v1/auth/setup', {
                method: 'POST',
                body: {
                    password,
                    setup_token: setupToken
                },
            });
            return true;
        } catch (error: any) {
            console.error('Setup failed', error);
            return false;
        }
    };

    const logout = async () => {
        try {
            await $fetch('/api/v1/auth/logout', { method: 'POST' })
        } catch (error: any) {
            console.warn('Logout request failed', error)
        }
        token.value = null;
        authState.resetAuth()
        navigateTo('/login');
    };

    const loginWithPasskey = async () => {
        try {
            const beginData = await $fetch<any>('/api/v1/auth/passkey/login/begin');
            const publicKey = beginData.assertion?.publicKey;

            if (!publicKey) {
                console.error('No publicKey found in beginData.assertion');
                return false;
            }

            let authResp;
            try {
                authResp = await startAuthentication({ optionsJSON: publicKey });
            } catch (authenticationError) {
                console.error('startAuthentication failed:', authenticationError);
                return false;
            }
            const finishData = await $fetch<{ token: string }>('/api/v1/auth/passkey/login/finish', {
                method: 'POST',
                body: authResp,
                headers: {
                    'X-Session-ID': beginData.session_id
                }
            });

            token.value = null;
            authState.setAuthenticated(true)
            return true;
        } catch (error: any) {
            console.error('Passkey login failed', error);
            authState.resetAuth()
            return false;
        }
    };

    const registerPasskey = async () => {
        try {
            const beginData = await $fetch<any>('/api/v1/auth/passkey/register/begin', {
            });
            const publicKey = beginData.creation?.publicKey;

            if (!publicKey) {
                console.error('No publicKey found in beginData');
                return false;
            }

            let authResp;
            try {
                authResp = await startRegistration({ optionsJSON: publicKey });
            } catch (registrationError) {
                console.error('startRegistration failed:', registrationError);
                return false;
            }
            await $fetch('/api/v1/auth/passkey/register/finish', {
                method: 'POST',
                body: authResp,
                headers: {
                    'X-Session-ID': beginData.session_id
                }
            });

            return true;
        } catch (error: any) {
            console.error('Passkey registration failed', error);
            return false;
        }
    };

    // 修改密码
    const changePassword = async (currentPassword: string, newPassword: string) => {
        try {
            await $fetch('/api/v1/auth/change-password', {
                method: 'POST',
                body: {
                    current_password: currentPassword,
                    new_password: newPassword
                }
            });
            return true;
        } catch (error: any) {
            console.error('Change password failed', error);
            return false;
        }
    };

    // 获取PassKey列表
    const listPasskeys = async () => {
        try {
            const data = await $fetch<{ credentials: any[], count: number }>('/api/v1/auth/passkeys', {
            });
            return data.credentials;
        } catch (error: any) {
            console.error('List passkeys failed', error);
            return [];
        }
    };

    // 删除PassKey
    const deletePasskey = async (credentialId: string) => {
        try {
            await $fetch(`/api/v1/auth/passkeys/${credentialId}`, {
                method: 'DELETE',
            });
            return true;
        } catch (error: any) {
            console.error('Delete passkey failed', error);
            return false;
        }
    };

    // 检查是否有PassKey
    const checkPasskeyExists = async () => {
        try {
            const data = await $fetch<{ has_passkey: boolean }>('/api/v1/auth/passkeys/check', {
            });
            return data.has_passkey;
        } catch (error: any) {
            console.error('Check passkey exists failed', error);
            return false;
        }
    };

    // API Token Management
    const createAPIToken = async (name: string, ipAllowlist: string[]) => {
        try {
            const data = await $fetch<CreateTokenResponse>('/api/v1/auth/tokens', {
                method: 'POST',
                body: { name, ip_allowlist: ipAllowlist }
            });
            return data;
        } catch (error: any) {
            console.error('Create API token failed', error);
            return null;
        }
    };

    const listAPITokens = async () => {
        try {
            const data = await $fetch<APIToken[]>('/api/v1/auth/tokens', {
            });
            return data;
        } catch (error: any) {
            console.error('List API tokens failed', error);
            return [];
        }
    };

    const deleteAPIToken = async (id: number) => {
        try {
            await $fetch(`/api/v1/auth/tokens/${id}`, {
                method: 'DELETE',
            });
            return true;
        } catch (error: any) {
            console.error('Delete API token failed', error);
            return false;
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
        deleteAPIToken
    };
}
