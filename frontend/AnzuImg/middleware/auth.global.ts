// 全局鉴权中间件
export default defineNuxtRouteMiddleware(async (to) => {
    const authState = useAuthState()
    const { checkInit } = useAuth()
    const { apiUrl } = useApi()
    const publicPaths = new Set<string>(['/login', '/setup'])
    const headers = process.server ? useRequestHeaders(['cookie']) : undefined
    let initialized = authState.initialized.value
    if (initialized !== true) {
        try {
            initialized = await checkInit({ headers, throwOnError: true })
        } catch {
            authState.resetAuth()
            if (!publicPaths.has(to.path)) {
                return navigateTo('/login')
            }
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
            await $fetch(apiUrl('/api/v1/auth/validate'), { headers })
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
            await $fetch(apiUrl('/api/v1/auth/validate'), { headers })
            authState.setAuthenticated(true)
        } catch {
            authState.resetAuth()
            return navigateTo('/login')
        }
    }
})


