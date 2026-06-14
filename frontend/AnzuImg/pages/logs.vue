<template>
    <h1 class="mb-6 text-3xl font-bold text-center">{{ t("logs.title") }}</h1>

    <div class="max-w-6xl mx-auto">
        <AnzuTabs v-model="activeTab" :tabs="tabs">
            <template #tab-content-0>
                <LogToolbar
                    :filter="appFilter"
                    :export-csv="() => exportLogs('app', 'csv')"
                    :export-json="() => exportLogs('app', 'json')"
                    :on-cleanup="() => openCleanup('app')"
                    :show-level-filter="true"
                    :show-module-filter="true"
                    :enable-stream="true"
                    v-model:streaming="streaming"
                    @apply="reloadApp(1)"
                    @stream-toggle="onStreamToggle"
                />

                <div v-if="loadingApp" class="flex justify-center py-6">
                    <AnzuProgressRing :size="40" />
                </div>
                <div
                    v-else-if="appLogs.length === 0"
                    class="rounded-lg bg-black/5 p-6 text-center dark:bg-white/5"
                >
                    <p class="text-(--md-sys-color-on-surface-variant)">
                        {{ t("logs.empty") }}
                    </p>
                </div>
                <div v-else>
                    <div class="flex flex-col gap-1">
                        <div
                            v-for="log in appLogs"
                            :key="`${log.id}-${log.created_at}`"
                            class="rounded-lg px-3 py-2.5 transition-colors hover:bg-black/5 dark:hover:bg-white/5"
                        >
                            <div class="flex flex-wrap items-center gap-2 text-xs">
                                <span :class="['inline-flex items-center px-2 py-0.5 rounded-lg text-xs font-semibold leading-4 whitespace-nowrap', levelBadge(log.level)]">
                                    {{ (log.level || '').toUpperCase() }}
                                </span>
                                <span class="inline-flex items-center px-2 py-0.5 rounded-lg text-xs font-semibold leading-4 whitespace-nowrap bg-black/5 text-(--md-sys-color-on-surface-variant) dark:bg-white/10">
                                    {{ log.module }}
                                </span>
                                <span class="ml-auto text-(--md-sys-color-on-surface-variant) whitespace-nowrap">
                                    {{ formatDate(log.created_at) }}
                                </span>
                            </div>
                            <div class="mt-1 break-words whitespace-pre-wrap text-sm text-(--md-sys-color-on-surface)">
                                {{ log.message }}
                            </div>
                            <div
                                v-if="log.request_id || log.ip_address"
                                class="mt-1 text-xs text-(--md-sys-color-on-surface-variant) break-all"
                            >
                                <span v-if="log.request_id">req_id={{ log.request_id }}</span>
                                <span v-if="log.ip_address"> · ip={{ log.ip_address }}</span>
                            </div>
                        </div>
                    </div>
                    <LogPagination
                        :page="appPage"
                        :total="appTotal"
                        :size="pageSize"
                        :loading="loadingApp"
                        @update="reloadApp"
                    />
                </div>
            </template>

            <template #tab-content-1>
                <LogToolbar
                    :filter="securityFilter"
                    :export-csv="() => exportLogs('security', 'csv')"
                    :export-json="() => exportLogs('security', 'json')"
                    :on-cleanup="() => openCleanup('security')"
                    @apply="reloadSecurity(1)"
                />
                <div v-if="loadingSecurity" class="flex justify-center py-6">
                    <AnzuProgressRing :size="40" />
                </div>
                <div
                    v-else-if="securityLogs.length === 0"
                    class="rounded-lg bg-black/5 p-6 text-center dark:bg-white/5"
                >
                    <p class="text-(--md-sys-color-on-surface-variant)">
                        {{ t("logs.empty") }}
                    </p>
                </div>
                <div v-else>
                    <div class="flex flex-col gap-1">
                        <div
                            v-for="log in securityLogs"
                            :key="log.id"
                            class="rounded-lg px-3 py-2.5 transition-colors hover:bg-black/5 dark:hover:bg-white/5"
                        >
                            <div class="flex flex-wrap items-center gap-2 text-xs">
                                <span :class="['inline-flex items-center px-2 py-0.5 rounded-lg text-xs font-semibold leading-4 whitespace-nowrap', securityLevelBadge(log.level)]">
                                    {{ (log.level || '').toUpperCase() }}
                                </span>
                                <span class="inline-flex items-center px-2 py-0.5 rounded-lg text-xs font-semibold leading-4 whitespace-nowrap bg-black/5 text-(--md-sys-color-on-surface-variant) dark:bg-white/10">
                                    {{ log.action }}
                                </span>
                                <span class="ml-auto text-(--md-sys-color-on-surface-variant) whitespace-nowrap">
                                    {{ formatDate(log.created_at) }}
                                </span>
                            </div>
                            <div class="mt-1 text-sm text-(--md-sys-color-on-surface) break-words">
                                {{ log.message }}
                            </div>
                            <div class="mt-1 text-xs text-(--md-sys-color-on-surface-variant) break-all">
                                <span v-if="log.method || log.path">{{ log.method }} {{ log.path }}</span>
                                <span v-if="log.ip_address"> · ip={{ log.ip_address }}</span>
                                <span v-if="log.username"> · user={{ log.username }}</span>
                            </div>
                        </div>
                    </div>
                    <LogPagination
                        :page="securityPage"
                        :total="securityTotal"
                        :size="pageSize"
                        :loading="loadingSecurity"
                        @update="reloadSecurity"
                    />
                </div>
            </template>

            <template #tab-content-2>
                <LogToolbar
                    :filter="tokenFilter"
                    :export-csv="() => exportLogs('token', 'csv')"
                    :export-json="() => exportLogs('token', 'json')"
                    :on-cleanup="() => openCleanup('token')"
                    @apply="reloadToken(1)"
                />
                <div v-if="loadingToken" class="flex justify-center py-6">
                    <AnzuProgressRing :size="40" />
                </div>
                <div
                    v-else-if="tokenLogs.length === 0"
                    class="rounded-lg bg-black/5 p-6 text-center dark:bg-white/5"
                >
                    <p class="text-(--md-sys-color-on-surface-variant)">
                        {{ t("logs.empty") }}
                    </p>
                </div>
                <div v-else>
                    <div class="flex flex-col gap-1">
                        <div
                            v-for="log in tokenLogs"
                            :key="log.id"
                            class="rounded-lg px-3 py-2.5 transition-colors hover:bg-black/5 dark:hover:bg-white/5"
                        >
                            <div class="flex flex-wrap items-center gap-2 text-xs">
                                <span class="inline-flex items-center px-2 py-0.5 rounded-lg text-xs font-semibold leading-4 whitespace-nowrap bg-(--md-sys-color-primary)/12 text-(--md-sys-color-primary)">
                                    {{ log.action }}
                                </span>
                                <span class="ml-auto text-(--md-sys-color-on-surface-variant) whitespace-nowrap">
                                    {{ formatDate(log.created_at) }}
                                </span>
                            </div>
                            <div class="mt-1 text-sm text-(--md-sys-color-on-surface) break-words">
                                <span>{{ log.token_name }} ({{ log.token_type }})</span>
                                <span v-if="log.ip_address" class="text-(--md-sys-color-on-surface-variant)"> · ip={{ log.ip_address }}</span>
                                <span v-if="log.image_hash" class="text-(--md-sys-color-on-surface-variant)"> · img={{ log.image_hash }}</span>
                            </div>
                            <div class="mt-1 text-xs text-(--md-sys-color-on-surface-variant) break-all">
                                {{ log.method }} {{ log.path }}
                            </div>
                        </div>
                    </div>
                    <LogPagination
                        :page="tokenPage"
                        :total="tokenTotal"
                        :size="pageSize"
                        :loading="loadingToken"
                        @update="reloadToken"
                    />
                </div>
            </template>
        </AnzuTabs>
    </div>

    <AnzuDialog
        v-model:visible="cleanupDialogVisible"
        :title="t('logs.cleanupTitle')"
        :variant="DialogVariant.DESTRUCTIVE"
        :actions="[
            { text: t('common.actions.cancel'), variant: 'text', handler: closeCleanup },
            { text: t('common.actions.confirm'), primary: true, variant: 'filled', handler: confirmCleanup, loading: cleanupRunning },
        ]"
    >
        <div class="flex flex-col gap-3">
            <p class="text-sm text-(--md-sys-color-on-surface-variant)">
                {{ t("logs.cleanupPromptDays") }}
            </p>
            <AnzuInput
                v-model="cleanupDays"
                type="number"
                :min="1"
                placeholder="30"
                @keydown.enter.prevent="confirmCleanup"
            />
        </div>
    </AnzuDialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import AnzuTabs from "~/components/AnzuTabs.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import AnzuDialog from "~/components/AnzuDialog.vue";
import LogToolbar from "~/components/log/LogToolbar.vue";
import LogPagination from "~/components/log/LogPagination.vue";
import { useLogs } from "~/composables/useLogs";
import { useLogStream } from "~/composables/useLogStream";
import { useNotification } from "~/composables/useNotification";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import { formatDate } from "~/utils/format";
import type { AppLog, LogFilter, LogSource } from "~/types/app_log";
import type { SecurityLog } from "~/types/security_log";
import type { APITokenLog } from "~/types/api_token";

const { t } = useI18n();
const { notify } = useNotification();
const logs = useLogs();
const stream = useLogStream();

const tabs = computed(() => [
    { label: t("logs.tabs.app"), value: "app" },
    { label: t("logs.tabs.security"), value: "security" },
    { label: t("logs.tabs.token"), value: "token" },
]);

const activeTab = ref<string | number>("app");
const pageSize = 50;

const emptyFilter = (): LogFilter => ({
    search: "",
    level: "",
    module: "",
    ip: "",
    action: "",
    start_date: "",
    end_date: "",
});

const appFilter = ref<LogFilter>(emptyFilter());
const securityFilter = ref<LogFilter>(emptyFilter());
const tokenFilter = ref<LogFilter>(emptyFilter());

const appLogs = ref<AppLog[]>([]);
const appPage = ref(1);
const appTotal = ref(0);
const loadingApp = ref(false);

const securityLogs = ref<SecurityLog[]>([]);
const securityPage = ref(1);
const securityTotal = ref(0);
const loadingSecurity = ref(false);

const tokenLogs = ref<APITokenLog[]>([]);
const tokenPage = ref(1);
const tokenTotal = ref(0);
const loadingToken = ref(false);

const streaming = ref(false);

const cleanupDialogVisible = ref(false);
const cleanupSource = ref<LogSource | null>(null);
const cleanupDays = ref<string>("30");
const cleanupRunning = ref(false);

const reloadApp = async (page = appPage.value) => {
    if (streaming.value) return;
    loadingApp.value = true;
    const res = await logs.listApp(page, pageSize, appFilter.value);
    appLogs.value = res.data;
    appPage.value = res.page;
    appTotal.value = res.total;
    loadingApp.value = false;
};

const reloadSecurity = async (page = securityPage.value) => {
    loadingSecurity.value = true;
    const res = await logs.listSecurity(page, pageSize, securityFilter.value);
    securityLogs.value = res.data;
    securityPage.value = res.page;
    securityTotal.value = res.total;
    loadingSecurity.value = false;
};

const reloadToken = async (page = tokenPage.value) => {
    loadingToken.value = true;
    const res = await logs.listToken(page, pageSize, tokenFilter.value);
    tokenLogs.value = res.data;
    tokenPage.value = res.page;
    tokenTotal.value = res.total;
    loadingToken.value = false;
};

const onStreamToggle = (on: boolean) => {
    if (on) {
        appLogs.value = [];
        appTotal.value = 0;
        stream.start({
            level: appFilter.value.level || "info",
            module: appFilter.value.module,
            onLog: (log) => {
                appLogs.value = [log, ...appLogs.value].slice(0, 200);
            },
        });
    } else {
        stream.stop();
        reloadApp(1);
    }
};

const exportLogs = (source: LogSource, format: "csv" | "json") => {
    const filter =
        source === "app"
            ? appFilter.value
            : source === "security"
              ? securityFilter.value
              : tokenFilter.value;
    const url = logs.exportUrl(source, format, filter);
    window.open(url, "_blank");
};

const openCleanup = (source: LogSource) => {
    cleanupSource.value = source;
    cleanupDays.value = "30";
    cleanupDialogVisible.value = true;
};

const closeCleanup = () => {
    if (cleanupRunning.value) return;
    cleanupDialogVisible.value = false;
};

const confirmCleanup = async () => {
    const n = Number(cleanupDays.value);
    if (!n || n <= 0 || !Number.isInteger(n)) {
        notify({ message: t("logs.cleanupInvalidDays"), type: NotificationType.WARNING });
        return;
    }
    const source = cleanupSource.value;
    if (!source) return;

    cleanupRunning.value = true;
    const res = await logs.cleanup(source, n);
    cleanupRunning.value = false;

    if (res) {
        cleanupDialogVisible.value = false;
        notify({
            message: t("logs.cleanupSuccess", { count: res.deleted }),
            type: NotificationType.SUCCESS,
        });
        if (source === "app") reloadApp(1);
        else if (source === "security") reloadSecurity(1);
        else reloadToken(1);
    } else {
        notify({ message: t("logs.cleanupFailed"), type: NotificationType.ERROR });
    }
};

const levelBadge = (level: string) => {
    switch ((level || "").toUpperCase()) {
        case "DEBUG":
            return "bg-slate-500/15 text-slate-600 dark:text-slate-300";
        case "INFO":
            return "bg-sky-500/15 text-sky-700 dark:text-sky-300";
        case "WARN":
            return "bg-amber-500/15 text-amber-700 dark:text-amber-300";
        case "ERROR":
        case "FATAL":
            return "bg-rose-500/15 text-rose-700 dark:text-rose-300";
        default:
            return "bg-(--md-sys-color-primary)/12 text-(--md-sys-color-primary)";
    }
};

const securityLevelBadge = (level: string) => {
    const l = (level || "").toLowerCase();
    if (l === "warn" || l === "warning") {
        return "bg-amber-500/15 text-amber-700 dark:text-amber-300";
    }
    if (l === "error" || l === "fatal") {
        return "bg-rose-500/15 text-rose-700 dark:text-rose-300";
    }
    return "bg-sky-500/15 text-sky-700 dark:text-sky-300";
};

onMounted(() => {
    reloadApp();
    reloadSecurity();
    reloadToken();
});
</script>
