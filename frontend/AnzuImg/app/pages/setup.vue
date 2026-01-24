<template>
    <div class="flex min-h-[calc(100vh-8rem)] items-center justify-center">
        <div class="w-full max-w-md rounded-xl">
            <h1 class="mb-2 text-center text-2xl font-bold">{{ t('setup.title') }}</h1>
            <p class="mb-6 text-center text-(--md-sys-color-on-surface-variant)">{{ t('setup.description') }}</p>

            <form @submit.prevent="handleSetup" class="flex flex-col gap-4">
                <AnzuInput v-model="password" type="password" :label="t('common.labels.password')"
                    :placeholder="t('setup.passwordPlaceholder')" />

                <AnzuInput v-model="confirmPassword" type="password" :label="t('setup.confirmPasswordLabel')"
                    :placeholder="t('setup.confirmPasswordPlaceholder')" />

                <AnzuButton type="submit" :status="loading ? 'loading' : 'default'" class="w-full">
                    {{ t('setup.submit') }}
                </AnzuButton>
            </form>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import AnzuButton from '~/components/AnzuButton.vue';
import AnzuInput from '~/components/AnzuInput.vue';
import { useAuth } from '~/composables/useAuth';
import { useNotification } from '~/composables/useNotification';
import { NotificationType } from '~/types/notification';
import { validatePassword } from '~/utils/password';

const { t } = useI18n();
const password = ref('');
const confirmPassword = ref('');
const loading = ref(false);
const { setup } = useAuth();
const router = useRouter();
const { notify } = useNotification();

const handleSetup = async () => {
    if (!password.value || !confirmPassword.value) return;

    const validation = validatePassword(password.value, t);
    if (!validation.valid) {
        notify({
            message: validation.error!,
            type: NotificationType.WARNING,
        });
        return;
    }

    if (password.value !== confirmPassword.value) {
        notify({
            message: t('setup.passwordMatchError'),
            type: NotificationType.WARNING,
        });
        return;
    }

    loading.value = true;

    const success = await setup(password.value);
    if (success) {
        notify({
            message: t('setup.success'),
            type: NotificationType.SUCCESS,
        });
        router.push('/login');
    } else {
        notify({
            message: t('setup.setupFailed'),
            type: NotificationType.ERROR,
        });
    }
    loading.value = false;
};
</script>
