<template>
    <div class="mb-12 max-w-3xl mx-auto">
        <h2 class="mb-4 text-xl font-semibold">
            {{ t("settings.changePassword.title") }}
        </h2>
        <p class="mb-6 text-(--md-sys-color-on-surface-variant)">
            {{ t("settings.changePassword.description") }}
        </p>

        <form
            @submit.prevent="handleSubmit"
            class="flex flex-col gap-4"
            autocomplete="on"
        >
            <input
                type="text"
                name="username"
                autocomplete="username"
                value="anzuimg"
                style="display: none"
            />

            <AnzuInput
                v-model="form.currentPassword"
                type="password"
                :label="t('settings.changePassword.currentPassword')"
                :placeholder="t('settings.changePassword.currentPasswordPlaceholder')"
                :disabled="loading"
                name="current-password"
                autocomplete="current-password"
            />

            <AnzuInput
                v-model="form.newPassword"
                type="password"
                :label="t('settings.changePassword.newPassword')"
                :placeholder="t('settings.changePassword.newPasswordPlaceholder')"
                :disabled="loading"
                name="new-password"
                autocomplete="new-password"
            />

            <AnzuInput
                v-model="form.confirmPassword"
                type="password"
                :label="t('settings.changePassword.confirmPassword')"
                :placeholder="t('settings.changePassword.confirmPasswordPlaceholder')"
                :disabled="loading"
                name="confirm-new-password"
                autocomplete="new-password"
            />

            <AnzuButton
                type="submit"
                :status="loading ? 'loading' : 'default'"
                class="w-full sm:w-auto"
            >
                {{ t("settings.changePassword.submit") }}
            </AnzuButton>
        </form>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import { useAuth } from "~/composables/useAuth";
import { useStepUp } from "~/composables/useStepUp";
import { useNotification } from "~/composables/useNotification";
import { parseApiError } from "~/utils/api-error";
import { NotificationType } from "~/types/notification";
import { validatePassword } from "~/utils/password";

const { t } = useI18n();
const { changePassword, logout } = useAuth();
const stepUp = useStepUp();
const { notify } = useNotification();

const form = ref({ currentPassword: "", newPassword: "", confirmPassword: "" });
const loading = ref(false);

const handleSubmit = async () => {
    if (!form.value.currentPassword || !form.value.newPassword || !form.value.confirmPassword) {
        notify({ message: t("settings.changePassword.fillAllFields"), type: NotificationType.WARNING });
        return;
    }
    const validation = validatePassword(form.value.newPassword, t);
    if (!validation.valid) {
        notify({ message: validation.error!, type: NotificationType.WARNING });
        return;
    }
    if (form.value.newPassword !== form.value.confirmPassword) {
        notify({ message: t("settings.changePassword.passwordMatchError"), type: NotificationType.WARNING });
        return;
    }

    const ok = await stepUp.request();
    if (!ok) return;

    loading.value = true;
    try {
        await changePassword(form.value.currentPassword, form.value.newPassword);
        notify({ message: t("settings.changePassword.success"), type: NotificationType.SUCCESS });
        setTimeout(() => logout(), 1500);
    } catch (error: any) {
        const parsed = parseApiError(error, t("settings.changePassword.failed"));
        notify({ message: parsed.displayMessage, type: NotificationType.ERROR });
    } finally {
        loading.value = false;
    }
};
</script>
