<template>
    <div class="mb-12 max-w-3xl mx-auto">
        <div class="mb-6 flex items-center justify-between">
            <div>
                <h2 class="text-xl font-semibold">
                    {{ t("settings.passkeyManagement.title") }}
                </h2>
                <p class="mt-1 text-(--md-sys-color-on-surface-variant)">
                    {{ t("settings.passkeyManagement.description") }}
                </p>
            </div>
            <AnzuButton
                @click="handleRegister"
                :status="registering ? 'loading' : 'default'"
                variant="text"
            >
                {{ t("settings.passkeyManagement.registerNew") }}
            </AnzuButton>
        </div>

        <div v-if="loading" class="flex justify-center py-8">
            <AnzuProgressRing :size="48" />
        </div>

        <div
            v-else-if="passkeys.length === 0"
            class="rounded-lg bg-black/5 p-8 text-center dark:bg-white/5"
        >
            <p class="text-(--md-sys-color-on-surface-variant)">
                {{ t("settings.passkeyManagement.noPasskeys") }}
            </p>
        </div>

        <div v-else class="flex flex-col gap-1">
            <div
                v-for="passkey in passkeys"
                :key="passkey.ID"
                class="flex items-start justify-between gap-2 rounded-lg px-3 py-2.5 transition-colors hover:bg-black/5 dark:hover:bg-white/5 min-w-0"
            >
                <div class="flex items-start gap-3 overflow-hidden min-w-0">
                    <div
                        class="rounded-full bg-black/5 p-2.5 text-(--md-sys-color-on-surface-variant) shrink-0 dark:bg-white/10"
                    >
                        <svg
                            class="h-6 w-6"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="1.5"
                                :d="getDeviceIcon(passkey.UserAgent || '')"
                            />
                        </svg>
                    </div>
                    <div class="min-w-0">
                        <h3
                            class="font-semibold text-(--md-sys-color-on-surface) break-words"
                            :title="passkey.DeviceName || `Passkey #${passkey.ID}`"
                        >
                            {{ passkey.DeviceName || `Passkey #${passkey.ID}` }}
                        </h3>
                        <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-1.5 text-xs text-(--md-sys-color-on-surface-variant)">
                            <span v-if="passkey.IPAddress" class="font-mono">
                                {{ passkey.IPAddress }}
                            </span>
                            <span :title="formatDate(passkey.UpdatedAt || passkey.CreatedAt)">
                                {{ formatRelativeTime(passkey.UpdatedAt || passkey.CreatedAt, locale) }}
                            </span>
                        </div>
                    </div>
                </div>

                <AnzuButton
                    @click="() => handleDelete(passkey.CredentialID)"
                    variant="text"
                    class="shrink-0 text-(--md-sys-color-error)"
                    :status="deletingId === passkey.CredentialID ? 'loading' : 'default'"
                >
                    <template #icon>
                        <TrashIcon
                            v-if="deletingId !== passkey.CredentialID"
                            class="h-5 w-5"
                        />
                    </template>
                </AnzuButton>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import { useAuth } from "~/composables/useAuth";
import { useStepUp } from "~/composables/useStepUp";
import { useNotification } from "~/composables/useNotification";
import { useDialog, isDialogDismissedError } from "~/composables/useDialog";
import { parseApiError } from "~/utils/api-error";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import { formatDate, formatRelativeTime } from "~/utils/format";
import { TrashIcon } from "@heroicons/vue/24/outline";
import type { PasskeyCredential } from "~/types/passkey";

const { t, locale } = useI18n();
const { listPasskeys, deletePasskey, registerPasskey } = useAuth();
const stepUp = useStepUp();
const { notify } = useNotification();
const { confirm } = useDialog();

const passkeys = ref<PasskeyCredential[]>([]);
const loading = ref(false);
const registering = ref(false);
const deletingId = ref<string | null>(null);

const loadPasskeys = async () => {
    loading.value = true;
    passkeys.value = await listPasskeys();
    loading.value = false;
};

const handleRegister = async () => {
    registering.value = true;
    try {
        await registerPasskey();
        notify({ message: t("settings.passkeyManagement.registerSuccess"), type: NotificationType.SUCCESS });
        await loadPasskeys();
    } catch (error: any) {
        const parsed = parseApiError(error, t("settings.passkeyManagement.registerFailed"));
        notify({ message: parsed.displayMessage, type: NotificationType.ERROR });
    } finally {
        registering.value = false;
    }
};

const handleDelete = async (credentialId: string) => {
    const result = await confirm(t("common.actions.deleteConfirm"), {
        title: t("common.actions.delete"),
        variant: DialogVariant.DESTRUCTIVE,
        actions: [
            { text: t("common.actions.cancel"), variant: "text" },
            { text: t("common.actions.delete"), primary: true, variant: "filled" },
        ],
    });
    if (!result) return;

    const ok = await stepUp.request();
    if (!ok) return;

    deletingId.value = credentialId;
    try {
        await deletePasskey(credentialId);
        notify({ message: t("common.actions.deleteSuccess"), type: NotificationType.SUCCESS });
        await loadPasskeys();
    } catch (error: any) {
        if (!isDialogDismissedError(error)) {
            const parsed = parseApiError(error, t("common.actions.deleteFailed"));
            notify({ message: parsed.displayMessage, type: NotificationType.ERROR });
        }
    } finally {
        deletingId.value = null;
    }
};

const getDeviceIcon = (ua: string = "") => {
    ua = ua.toLowerCase();
    if (ua.includes("mobile") || ua.includes("android") || ua.includes("iphone") || ua.includes("ipad")) {
        return "M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3";
    }
    return "M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25";
};

onMounted(loadPasskeys);
</script>
