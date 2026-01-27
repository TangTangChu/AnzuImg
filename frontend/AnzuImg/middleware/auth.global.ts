// 全局鉴权中间件
export default defineNuxtRouteMiddleware(async (to) => {
    const authState = useAuthState()
    const publicPaths = new Set<string>(['/login', '/setup'])
    const headers = process.server ? useRequestHeaders(['cookie']) : undefined
    let initialized = authState.initialized.value
    if (initialized !== true) {
        try {
            const data = await $fetch<{ initialized: boolean }>('/api/v1/auth/status', { headers })
            initialized = !!data.initialized
            authState.setInitialized(initialized)
        } catch {
            return
        }
    }

    if (!initialized) {
        authState.resetAuth()
        if (to.path !== '/setup') {
            return navigateTo('/setup')
        }
        return
    }

    if (to.path === '/setup') {
        return navigateTo('/login')
    }

    if (to.path === '/login') {
        try {
            await $fetch('/api/v1/auth/validate', { headers })
            authState.setAuthenticated(true)
            return navigateTo('/gallery')
        } catch {
            authState.resetAuth()
            return
        }
    }

    if (!publicPaths.has(to.path)) {
        if (authState.isAuthValid()) {
            return
        }

        try {
            await $fetch('/api/v1/auth/validate', { headers })
            authState.setAuthenticated(true)
        } catch {
            authState.resetAuth()
            return navigateTo('/login')
        }
    }
})


