<template>
    <h1 class="mb-6 text-3xl font-bold text-center">{{ t("settings.title") }}</h1>

    <Dashboard />

    <!-- 修改密码 -->
    <div class="mb-12 max-w-3xl mx-auto">
        <h2 class="mb-4 text-xl font-semibold">
            {{ t("settings.changePassword.title") }}
        </h2>
        <p class="mb-6 text-(--md-sys-color-on-surface-variant)">
            {{ t("settings.changePassword.description") }}
        </p>

        <form
            @submit.prevent="handleChangePassword"
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
                v-model="passwordForm.currentPassword"
                type="password"
                :label="t('settings.changePassword.currentPassword')"
                :placeholder="t('settings.changePassword.currentPasswordPlaceholder')"
                :disabled="changingPassword"
                name="current-password"
                autocomplete="current-password"
            />

            <AnzuInput
                v-model="passwordForm.newPassword"
                type="password"
                :label="t('settings.changePassword.newPassword')"
                :placeholder="t('settings.changePassword.newPasswordPlaceholder')"
                :disabled="changingPassword"
                name="new-password"
                autocomplete="new-password"
            />

            <AnzuInput
                v-model="passwordForm.confirmPassword"
                type="password"
                :label="t('settings.changePassword.confirmPassword')"
                :placeholder="t('settings.changePassword.confirmPasswordPlaceholder')"
                :disabled="changingPassword"
                name="confirm-new-password"
                autocomplete="new-password"
            />

            <AnzuButton
                type="submit"
                :status="changingPassword ? 'loading' : 'default'"
                class="w-full sm:w-auto"
            >
                {{ t("settings.changePassword.submit") }}
            </AnzuButton>
        </form>
    </div>

    <!-- PassKey 管理 -->
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
                @click="handleRegisterPasskey"
                :status="registeringPasskey ? 'loading' : 'default'"
                variant="text"
            >
                {{ t("settings.passkeyManagement.registerNew") }}
            </AnzuButton>
        </div>

        <div v-if="loadingPasskeys" class="flex justify-center py-8">
            <AnzuProgressRing :size="48" />
        </div>

        <div
            v-else-if="passkeys.length === 0"
            class="rounded-lg border border-(--md-sys-color-outline-variant) p-8 text-center"
        >
            <p class="text-(--md-sys-color-on-surface-variant)">
                {{ t("settings.passkeyManagement.noPasskeys") }}
            </p>
        </div>

        <div v-else class="grid gap-4 sm:grid-cols-1 lg:grid-cols-2">
            <div
                v-for="passkey in passkeys"
                :key="passkey.ID"
                class="relative flex flex-col justify-between rounded-xl border border-(--md-sys-color-outline-variant) p-4 transition-colors min-w-0"
            >
                <div class="flex items-start justify-between mb-3">
                    <div class="flex items-start gap-3 overflow-hidden min-w-0">
                        <div
                            class="rounded-full border border-(--md-sys-color-outline-variant) p-2.5 text-(--md-sys-color-on-surface-variant) shrink-0"
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
                                class="font-semibold text-(--md-sys-color-on-surface) truncate"
                                :title="passkey.DeviceName || `Passkey #${passkey.ID}`"
                            >
                                {{ passkey.DeviceName || `Passkey #${passkey.ID}` }}
                            </h3>
                            <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-1.5 text-xs text-(--md-sys-color-on-surface-variant)">
                                <span
                                    v-if="passkey.IPAddress"
                                    class="font-mono"
                                >
                                    {{ passkey.IPAddress }}
                                </span>
                                <span
                                    :title="formatDate(passkey.UpdatedAt || passkey.CreatedAt)"
                                >
                                    {{ formatRelativeTime(passkey.UpdatedAt || passkey.CreatedAt, locale) }}
                                </span>
                            </div>
                        </div>
                    </div>

                    <AnzuButton
                        @click="() => handleDeletePasskey(passkey.CredentialID)"
                        variant="text"
                        class="min-w-0! p-2! h-9! w-9! shrink-0 -mr-2 -mt-2 text-(--md-sys-color-error)"
                        :status="deletingPasskeyId === passkey.CredentialID ? 'loading' : 'default'"
                    >
                        <TrashIcon
                            v-if="deletingPasskeyId !== passkey.CredentialID"
                            class="h-5 w-5"
                        />
                    </AnzuButton>
                </div>
            </div>
        </div>
    </div>

    <!-- API Token 管理 -->
    <div class="mb-12 max-w-3xl mx-auto">
        <div class="mb-6 flex items-center justify-between">
            <div>
                <h2 class="text-xl font-semibold">API Token</h2>
                <p class="mt-1 text-(--md-sys-color-on-surface-variant)">
                    {{ t("settings.apiTokens.description") }}
                </p>
            </div>
            <AnzuButton @click="showCreateTokenDialog = true" variant="text">
                {{ t("settings.apiTokens.createNew") }}
            </AnzuButton>
        </div>

        <div v-if="loadingTokens" class="flex justify-center py-8">
            <AnzuProgressRing :size="48" />
        </div>
        <div
            v-else-if="apiTokens.length === 0"
            class="rounded-lg border border-(--md-sys-color-outline-variant) p-8 text-center"
        >
            <p class="text-(--md-sys-color-on-surface-variant)">
                {{ t("settings.apiTokens.noTokens") }}
            </p>
        </div>
        <div v-else class="grid gap-4 sm:grid-cols-1 lg:grid-cols-2">
            <div
                v-for="token in apiTokens"
                :key="token.id"
                class="flex flex-col justify-between rounded-xl border border-(--md-sys-color-outline-variant) p-4 transition-colors"
            >
                <div class="flex items-start justify-between">
                    <div class="min-w-0">
                        <h3 class="font-semibold text-(--md-sys-color-on-surface) truncate">{{ token.name }}</h3>
                        <div class="mt-1.5 flex flex-wrap items-center gap-1.5">
                            <span class="inline-flex items-center text-xs px-1.5 py-0.5 rounded border border-(--md-sys-color-outline-variant) text-(--md-sys-color-on-surface-variant)">
                                {{ getTokenTypeLabel(token.token_type) }}
                            </span>
                            <span
                                v-if="!token.ip_allowlist || token.ip_allowlist.length === 0"
                                class="inline-flex items-center text-xs px-1.5 py-0.5 rounded border border-(--md-sys-color-error) text-(--md-sys-color-error)"
                            >
                                {{ t("settings.apiTokens.anyIP") }}
                            </span>
                            <span
                                v-else
                                v-for="ip in token.ip_allowlist"
                                :key="ip"
                                class="inline-flex items-center text-xs px-1.5 py-0.5 rounded border border-(--md-sys-color-outline-variant) text-(--md-sys-color-on-surface-variant) font-mono"
                            >
                                {{ ip }}
                            </span>
                        </div>
                    </div>
                    <AnzuButton
                        @click="() => handleDeleteToken(token.id)"
                        variant="text"
                        class="min-w-0! p-2! h-9! w-9! shrink-0 text-(--md-sys-color-error)"
                    >
                        <TrashIcon class="h-5 w-5" />
                    </AnzuButton>
                </div>
                <div class="mt-3 pt-2 border-t border-(--md-sys-color-outline-variant) flex justify-between text-xs text-(--md-sys-color-on-surface-variant)">
                    <span>{{ t("settings.apiTokens.created") }}: {{ formatDate(token.created_at) }}</span>
                    <span>
                        {{ t("settings.apiTokens.lastUsed") }}:
                        {{
                            token.last_used_at
                                ? formatRelativeTime(token.last_used_at, locale)
                                : t("settings.apiTokens.neverUsed")
                        }}
                    </span>
                </div>
            </div>
        </div>
    </div>

    <!-- 系统配置 -->
    <div class="mb-12 max-w-3xl mx-auto">
        <div class="mb-4">
            <h2 class="text-xl font-semibold">{{ t("settings.systemConfig.title") }}</h2>
            <p class="mt-1 text-(--md-sys-color-on-surface-variant)">
                {{ t("settings.systemConfig.description") }}
            </p>
        </div>

        <div v-if="settingsLoading" class="flex justify-center py-12">
            <AnzuProgressRing :size="48" />
        </div>
        <SystemConfigSection
            v-else-if="settingsResp"
            :schema="settingsResp.schema"
            :values="settingsResp.values"
            :allow-web-modify="settingsResp.allow_web_modify"
            :saving="savingSettings"
            @save="onSaveSettings"
            @reset="onResetSettings"
        />
    </div>

    <AnzuDialog
        v-model:visible="showCreateTokenDialog"
        :title="t('settings.apiTokens.createNew')"
        :actions="[
            { text: t('common.actions.cancel'), variant: 'text', handler: () => { showCreateTokenDialog = false; } },
            { text: t('settings.apiTokens.createNew'), primary: true, variant: 'filled', handler: handleCreateToken, loading: creatingToken },
        ]"
    >
        <div class="flex flex-col gap-4">
            <div class="flex flex-col gap-2">
                <span class="text-sm font-medium text-(--md-sys-color-on-surface)">{{ t("settings.apiTokens.tokenType") }}</span>
                <AnzuSelector
                    v-model="tokenForm.tokenType"
                    :options="tokenTypeOptions"
                />
            </div>
            <AnzuInput
                v-model="tokenForm.name"
                :label="t('settings.apiTokens.name')"
                :placeholder="t('settings.apiTokens.namePlaceholder')"
                name="token-name"
                autocomplete="off"
            />
            <AnzuTags
                v-model="tokenForm.ipAllowlist"
                :label="t('settings.apiTokens.ipAllowlist')"
                :max-tags="10"
                :hint="t('settings.apiTokens.ipAllowlistTip')"
            />
        </div>
    </AnzuDialog>

    <AnzuDialog
        v-model:visible="showTokenResultDialog"
        :title="t('settings.apiTokens.tokenCreatedTitle')"
        :actions="[
            { text: t('common.actions.close'), variant: 'filled', handler: () => { showTokenResultDialog = false; } },
        ]"
    >
        <div class="flex flex-col gap-4">
            <p class="text-(--md-sys-color-on-surface-variant)">{{ t("settings.apiTokens.tokenCreatedMessage") }}</p>
            <div class="relative">
                <AnzuInput :model-value="createdTokenRaw" readonly name="api-token" autocomplete="off" />
                <AnzuButton @click="copyToken" variant="tonal" class="absolute right-1 top-1 bottom-1">
                    {{ t("settings.apiTokens.copy") }}
                </AnzuButton>
            </div>
        </div>
    </AnzuDialog>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import Dashboard from "~/components/Dashboard.vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuSelector from "~/components/AnzuSelector.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuTags from "~/components/AnzuTags.vue";
import AnzuDialog from "~/components/AnzuDialog.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import SystemConfigSection from "~/components/SystemConfigSection.vue";
import { useAuth } from "~/composables/useAuth";
import { useSettings } from "~/composables/useSettings";
import { useNotification } from "~/composables/useNotification";
import { useDialog, isDialogDismissedError } from "~/composables/useDialog";
import { parseApiError } from "~/utils/api-error";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import { formatDate, formatRelativeTime } from "~/utils/format";
import { validatePassword } from "~/utils/password";
import { TrashIcon } from "@heroicons/vue/24/outline";
import type { APIToken } from "~/types/api_token";
import type { PasskeyCredential } from "~/types/passkey";
import type { SettingsResponse } from "~/types/settings";

const { t, locale } = useI18n();
const {
    changePassword,
    logout,
    listPasskeys,
    deletePasskey,
    registerPasskey,
    createAPIToken,
    listAPITokens,
    deleteAPIToken,
    getLastApiErrorDisplay,
} = useAuth();
const settingsApi = useSettings();
const { notify } = useNotification();
const { confirm } = useDialog();

const passwordForm = ref({ currentPassword: "", newPassword: "", confirmPassword: "" });
const changingPassword = ref(false);

const passkeys = ref<PasskeyCredential[]>([]);
const loadingPasskeys = ref(false);
const registeringPasskey = ref(false);
const deletingPasskeyId = ref<string | null>(null);

const apiTokens = ref<APIToken[]>([]);
const loadingTokens = ref(false);
const showCreateTokenDialog = ref(false);
const creatingToken = ref(false);
const tokenForm = ref({ name: "", ipAllowlist: [] as string[], tokenType: "full" });
const showTokenResultDialog = ref(false);
const createdTokenRaw = ref("");

const settingsResp = ref<SettingsResponse | null>(null);
const settingsLoading = ref(false);
const savingSettings = ref(false);

const tokenTypeOptions = computed(() => [
    { value: "full", label: t("settings.apiTokens.tokenTypes.full") },
    { value: "upload", label: t("settings.apiTokens.tokenTypes.upload") },
    { value: "list", label: t("settings.apiTokens.tokenTypes.list") },
]);

const loadTokens = async () => {
    loadingTokens.value = true;
    apiTokens.value = await listAPITokens();
    loadingTokens.value = false;
};

const handleCreateToken = async () => {
    if (!tokenForm.value.name) return;
    creatingToken.value = true;
    const res = await createAPIToken(
        tokenForm.value.name,
        tokenForm.value.ipAllowlist,
        tokenForm.value.tokenType,
    );
    creatingToken.value = false;
    if (res) {
        createdTokenRaw.value = res.raw_token;
        showCreateTokenDialog.value = false;
        showTokenResultDialog.value = true;
        loadTokens();
        tokenForm.value = { name: "", ipAllowlist: [], tokenType: "full" };
        notify({ message: t("settings.apiTokens.createSuccess"), type: NotificationType.SUCCESS });
    } else {
        notify({
            message: getLastApiErrorDisplay(t("settings.apiTokens.createFailed")),
            type: NotificationType.ERROR,
        });
    }
};

const handleDeleteToken = async (id: number) => {
    const result = await confirm(t("common.actions.deleteConfirm"), {
        title: t("common.actions.delete"),
        variant: DialogVariant.DESTRUCTIVE,
        actions: [
            { text: t("common.actions.cancel"), variant: "text" },
            { text: t("common.actions.delete"), primary: true, variant: "filled" },
        ],
    });
    if (!result) return;
    if (await deleteAPIToken(id)) {
        loadTokens();
        notify({ message: t("common.actions.deleteSuccess"), type: NotificationType.SUCCESS });
    } else {
        notify({
            message: getLastApiErrorDisplay(t("common.actions.deleteFailed")),
            type: NotificationType.ERROR,
        });
    }
};

const copyToken = () => {
    navigator.clipboard.writeText(createdTokenRaw.value);
    notify({ message: t("settings.apiTokens.copySuccess"), type: NotificationType.SUCCESS });
};

const getTokenTypeLabel = (type: string) => {
    switch (type) {
        case "upload":
            return t("settings.apiTokens.tokenTypes.upload");
        case "list":
            return t("settings.apiTokens.tokenTypes.list");
        default:
            return t("settings.apiTokens.tokenTypes.full");
    }
};

const handleChangePassword = async () => {
    if (
        !passwordForm.value.currentPassword ||
        !passwordForm.value.newPassword ||
        !passwordForm.value.confirmPassword
    ) {
        notify({ message: t("settings.changePassword.fillAllFields"), type: NotificationType.WARNING });
        return;
    }
    const validation = validatePassword(passwordForm.value.newPassword, t);
    if (!validation.valid) {
        notify({ message: validation.error!, type: NotificationType.WARNING });
        return;
    }
    if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
        notify({ message: t("settings.changePassword.passwordMatchError"), type: NotificationType.WARNING });
        return;
    }
    changingPassword.value = true;
    const success = await changePassword(passwordForm.value.currentPassword, passwordForm.value.newPassword);
    if (success) {
        notify({ message: t("settings.changePassword.success"), type: NotificationType.SUCCESS });
        setTimeout(() => logout(), 1500);
    } else {
        notify({
            message: getLastApiErrorDisplay(t("settings.changePassword.failed")),
            type: NotificationType.ERROR,
        });
    }
    changingPassword.value = false;
};

const handleRegisterPasskey = async () => {
    registeringPasskey.value = true;
    const success = await registerPasskey();
    if (success) {
        notify({ message: t("settings.passkeyManagement.registerSuccess"), type: NotificationType.SUCCESS });
        loadPasskeys();
    } else {
        notify({
            message: getLastApiErrorDisplay(t("settings.passkeyManagement.registerFailed")),
            type: NotificationType.ERROR,
        });
    }
    registeringPasskey.value = false;
};

const handleDeletePasskey = async (credentialId: string) => {
    try {
        const result = await confirm(t("common.actions.deleteConfirm"), {
            title: t("common.actions.delete"),
            variant: DialogVariant.DESTRUCTIVE,
            actions: [
                { text: t("common.actions.cancel"), variant: "text" },
                { text: t("common.actions.delete"), primary: true, variant: "filled", loading: false },
            ],
        });
        if (!result) return;
        deletingPasskeyId.value = credentialId;
        const success = await deletePasskey(credentialId);
        if (success) {
            notify({ message: t("common.actions.deleteSuccess"), type: NotificationType.SUCCESS });
            loadPasskeys();
        } else {
            notify({
                message: getLastApiErrorDisplay(t("common.actions.deleteFailed")),
                type: NotificationType.ERROR,
            });
        }
    } catch (e: any) {
        if (isDialogDismissedError(e)) return;
        const parsed = parseApiError(e, t("common.actions.deleteFailed"));
        notify({ message: parsed.displayMessage, type: NotificationType.ERROR });
    } finally {
        deletingPasskeyId.value = null;
    }
};

const loadPasskeys = async () => {
    loadingPasskeys.value = true;
    passkeys.value = await listPasskeys();
    loadingPasskeys.value = false;
};

const getDeviceIcon = (ua: string = "") => {
    ua = ua.toLowerCase();
    if (ua.includes("mobile") || ua.includes("android") || ua.includes("iphone") || ua.includes("ipad")) {
        return "M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3";
    }
    return "M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25";
};

const loadSettings = async () => {
    settingsLoading.value = true;
    settingsResp.value = await settingsApi.get();
    settingsLoading.value = false;
};

const onSaveSettings = async (values: Record<string, string>) => {
    savingSettings.value = true;
    const ok = await settingsApi.patch(values);
    savingSettings.value = false;
    if (ok) {
        notify({ message: t("settings.systemConfig.saveSuccess"), type: NotificationType.SUCCESS });
        await loadSettings();
    } else {
        notify({ message: t("settings.systemConfig.saveFailed"), type: NotificationType.ERROR });
    }
};

const onResetSettings = async (keys: string[]) => {
    if (keys.length === 0) return;
    const ok = await confirm(t("settings.systemConfig.resetConfirm", { count: keys.length }), {
        title: t("settings.systemConfig.resetOverridden"),
        variant: DialogVariant.WARNING,
        actions: [
            { text: t("common.actions.cancel"), variant: "text" },
            { text: t("common.actions.confirm"), primary: true, variant: "filled" },
        ],
    });
    if (!ok) return;
    savingSettings.value = true;
    const r = await settingsApi.reset(keys);
    savingSettings.value = false;
    if (r) {
        notify({ message: t("settings.systemConfig.resetSuccess"), type: NotificationType.SUCCESS });
        await loadSettings();
    } else {
        notify({ message: t("settings.systemConfig.resetFailed"), type: NotificationType.ERROR });
    }
};

onMounted(() => {
    loadPasskeys();
    loadTokens();
    loadSettings();
});
</script>
