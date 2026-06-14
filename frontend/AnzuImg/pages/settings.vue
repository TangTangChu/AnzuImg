<template>
    <h1 class="mb-6 text-3xl font-bold text-center">{{ t("settings.title") }}</h1>

    <Dashboard />

    <PasswordSection />
    <PasskeySection />
    <TokenSection />

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
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import Dashboard from "~/components/Dashboard.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import PasswordSection from "~/components/settings/PasswordSection.vue";
import PasskeySection from "~/components/settings/PasskeySection.vue";
import TokenSection from "~/components/settings/TokenSection.vue";
import SystemConfigSection from "~/components/SystemConfigSection.vue";
import { useSettings } from "~/composables/useSettings";
import { useNotification } from "~/composables/useNotification";
import { useDialog } from "~/composables/useDialog";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import type { SettingsResponse } from "~/types/settings";

const { t } = useI18n();
const settingsApi = useSettings();
const { notify } = useNotification();
const { confirm } = useDialog();

const settingsResp = ref<SettingsResponse | null>(null);
const settingsLoading = ref(false);
const savingSettings = ref(false);

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
    const result = await confirm(t("settings.systemConfig.resetConfirm", { count: keys.length }), {
        title: t("settings.systemConfig.resetOverridden"),
        variant: DialogVariant.WARNING,
        actions: [
            { text: t("common.actions.cancel"), variant: "text" },
            { text: t("common.actions.confirm"), primary: true, variant: "filled" },
        ],
    });
    if (!result) return;

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

onMounted(loadSettings);
</script>
