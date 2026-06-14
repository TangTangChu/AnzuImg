<template>
    <div class="mb-12 max-w-3xl mx-auto">
        <div class="mb-6 flex items-center justify-between">
            <div>
                <h2 class="text-xl font-semibold">API Token</h2>
                <p class="mt-1 text-(--md-sys-color-on-surface-variant)">
                    {{ t("settings.apiTokens.description") }}
                </p>
            </div>
            <AnzuButton @click="handleOpenCreate" variant="text">
                {{ t("settings.apiTokens.createNew") }}
            </AnzuButton>
        </div>

        <div v-if="loading" class="flex justify-center py-8">
            <AnzuProgressRing :size="48" />
        </div>
        <div
            v-else-if="tokens.length === 0"
            class="rounded-lg bg-black/5 p-8 text-center dark:bg-white/5"
        >
            <p class="text-(--md-sys-color-on-surface-variant)">
                {{ t("settings.apiTokens.noTokens") }}
            </p>
        </div>
        <div v-else class="flex flex-col gap-1">
            <div
                v-for="token in tokens"
                :key="token.id"
                class="flex flex-col rounded-lg px-3 py-2.5 transition-colors hover:bg-black/5 dark:hover:bg-white/5"
            >
                <div class="flex items-center justify-between">
                    <div class="min-w-0 flex flex-wrap items-center gap-2">
                        <h3 class="font-semibold text-(--md-sys-color-on-surface) break-words">{{ token.name }}</h3>
                        <span class="inline-flex items-center text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)">
                            {{ getTokenTypeLabel(token.token_type) }}
                        </span>
                        <span
                            v-if="!token.ip_allowlist || token.ip_allowlist.length === 0"
                            class="inline-flex items-center text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-error)/12 text-(--md-sys-color-error)"
                        >
                            {{ t("settings.apiTokens.anyIP") }}
                        </span>
                        <span
                            v-else
                            v-for="ip in token.ip_allowlist"
                            :key="ip"
                            class="inline-flex items-center text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-primary-container) text-(--md-sys-color-on-primary-container) font-mono"
                        >
                            {{ ip }}
                        </span>
                    </div>
                    <AnzuButton
                        @click="() => handleDelete(token.id)"
                        variant="text"
                        class="shrink-0 text-(--md-sys-color-error)"
                    >
                        <template #icon>
                            <TrashIcon class="h-5 w-5" />
                        </template>
                    </AnzuButton>
                </div>
                <div class="mt-3 flex justify-between text-xs text-(--md-sys-color-on-surface-variant)">
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

        <AnzuDialog
            v-model:visible="showCreateDialog"
            :title="t('settings.apiTokens.createNew')"
            :actions="[
                { text: t('common.actions.cancel'), variant: 'text', handler: () => { showCreateDialog = false; } },
                { text: t('settings.apiTokens.createNew'), primary: true, variant: 'filled', handler: handleCreate, loading: creating },
            ]"
        >
            <div class="flex flex-col gap-4">
                <div class="flex flex-col gap-2">
                    <span class="text-sm font-medium text-(--md-sys-color-on-surface)">{{ t("settings.apiTokens.tokenType") }}</span>
                    <AnzuSelector
                        v-model="form.tokenType"
                        :options="tokenTypeOptions"
                    />
                </div>
                <AnzuInput
                    v-model="form.name"
                    :label="t('settings.apiTokens.name')"
                    :placeholder="t('settings.apiTokens.namePlaceholder')"
                    name="token-name"
                    autocomplete="off"
                />
                <AnzuTags
                    v-model="form.ipAllowlist"
                    :label="t('settings.apiTokens.ipAllowlist')"
                    :max-tags="10"
                    :hint="t('settings.apiTokens.ipAllowlistTip')"
                />
            </div>
        </AnzuDialog>

        <AnzuDialog
            v-model:visible="showResultDialog"
            :title="t('settings.apiTokens.tokenCreatedTitle')"
            :actions="[
                { text: t('common.actions.close'), variant: 'filled', handler: () => { showResultDialog = false; } },
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
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuSelector from "~/components/AnzuSelector.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuTags from "~/components/AnzuTags.vue";
import AnzuDialog from "~/components/AnzuDialog.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import { useAuth } from "~/composables/useAuth";
import { useStepUp } from "~/composables/useStepUp";
import { useApi } from "~/composables/useApi";
import { useNotification } from "~/composables/useNotification";
import { useDialog } from "~/composables/useDialog";
import { parseApiError } from "~/utils/api-error";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import { formatDate, formatRelativeTime } from "~/utils/format";
import { TrashIcon } from "@heroicons/vue/24/outline";
import type { APIToken, CreateTokenResponse } from "~/types/api_token";

const { t, locale } = useI18n();
const { listAPITokens, deleteAPIToken } = useAuth();
const { apiUrl } = useApi();
const stepUp = useStepUp();
const { notify } = useNotification();
const { confirm } = useDialog();

const tokens = ref<APIToken[]>([]);
const loading = ref(false);
const showCreateDialog = ref(false);
const creating = ref(false);
const form = ref({ name: "", ipAllowlist: [] as string[], tokenType: "full" });
const showResultDialog = ref(false);
const createdTokenRaw = ref("");

const tokenTypeOptions = computed(() => [
    { value: "full", label: t("settings.apiTokens.tokenTypes.full") },
    { value: "upload", label: t("settings.apiTokens.tokenTypes.upload") },
    { value: "list", label: t("settings.apiTokens.tokenTypes.list") },
]);

const loadTokens = async () => {
    loading.value = true;
    tokens.value = await listAPITokens();
    loading.value = false;
};

const handleOpenCreate = async () => {
    const ok = await stepUp.request();
    if (!ok) return;
    showCreateDialog.value = true;
};

const handleCreate = async () => {
    if (!form.value.name) return;
    creating.value = true;
    try {
        const res = await $fetch<CreateTokenResponse>(apiUrl('/api/v1/auth/tokens'), {
            method: 'POST',
            body: {
                name: form.value.name,
                ip_allowlist: form.value.ipAllowlist,
                token_type: form.value.tokenType,
            },
        });
        createdTokenRaw.value = res.raw_token;
        showCreateDialog.value = false;
        showResultDialog.value = true;
        await loadTokens();
        form.value = { name: "", ipAllowlist: [], tokenType: "full" };
        notify({ message: t("settings.apiTokens.createSuccess"), type: NotificationType.SUCCESS });
    } catch (error: any) {
        const parsed = parseApiError(error, t("settings.apiTokens.createFailed"));
        notify({ message: parsed.displayMessage, type: NotificationType.ERROR });
    } finally {
        creating.value = false;
    }
};

const handleDelete = async (id: number) => {
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

    try {
        await deleteAPIToken(id);
        await loadTokens();
        notify({ message: t("common.actions.deleteSuccess"), type: NotificationType.SUCCESS });
    } catch (error: any) {
        const parsed = parseApiError(error, t("common.actions.deleteFailed"));
        notify({ message: parsed.displayMessage, type: NotificationType.ERROR });
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

onMounted(loadTokens);
</script>
