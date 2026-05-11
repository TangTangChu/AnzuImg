<template>
    <h1 class="mb-6 text-3xl font-bold text-center">{{ t("logs.title") }}</h1>

    <div class="max-w-5xl mx-auto">
        <AnzuTabs v-model="activeTab" :tabs="tabs">
            <template #tab-content-0>
                <LogToolbar
                    :filter="appFilter"
                    @apply="reloadApp(1)"
                    :export-csv="() => exportLogs('app', 'csv')"
                    :export-json="() => exportLogs('app', 'json')"
                    :on-cleanup="() => onCleanup('app')"
                    :show-level-filter="true"
                    :show-module-filter="true"
                    :enable-stream="true"
                    v-model:streaming="streaming"
                    @stream-toggle="onStreamToggle"
                />

                <div v-if="loadingApp" class="flex justify-center py-6">
                    <AnzuProgressRing :size="40" />
                </div>
                <div
                    v-else-if="appLogs.length === 0"
                    class="rounded-lg border border-(--md-sys-color-outline-variant) p-6 text-center"
                >
                    <p class="text-(--md-sys-color-on-surface-variant)">
                        {{ t("logs.empty") }}
                    </p>
                </div>
                <div v-else class="space-y-2">
                    <div
                        v-for="log in appLogs"
                        :key="`${log.id}-${log.created_at}`"
                        class="rounded-lg border border-(--md-sys-color-outline-variant) p-3 text-sm"
                    >
                        <div class="flex flex-wrap items-center gap-2">
                            <span
                                class="inline-flex items-center text-xs px-1.5 py-1 rounded font-semibold whitespace-nowrap"
                                :class="levelClass(log.level)"
                            >
                                {{ (log.level || '').toUpperCase() }}
                            </span>
                            <span class="inline-flex items-center text-xs px-1.5 py-1 rounded bg-(--md-sys-color-surface-variant) text-(--md-sys-color-on-surface-variant) whitespace-nowrap">
                                {{ log.module }}
                            </span>
                            <span class="ml-auto inline-flex items-center text-xs text-(--md-sys-color-on-surface-variant) whitespace-nowrap">
                                {{ formatDate(log.created_at) }}
                            </span>
                        </div>
                        <div class="mt-1 break-all whitespace-pre-wrap text-sm">
                            {{ log.message }}
                        </div>
                        <div v-if="log.request_id || log.ip_address" class="mt-1 text-xs opacity-70 break-all">
                            <span v-if="log.request_id">req_id: {{ log.request_id }} </span>
                            <span v-if="log.ip_address"> · ip: {{ log.ip_address }}</span>
                        </div>
                    </div>
                    <Pagination
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
                    @apply="reloadSecurity(1)"
                    :export-csv="() => exportLogs('security', 'csv')"
                    :export-json="() => exportLogs('security', 'json')"
                    :on-cleanup="() => onCleanup('security')"
                />
                <div v-if="loadingSecurity" class="flex justify-center py-6">
                    <AnzuProgressRing :size="40" />
                </div>
                <div
                    v-else-if="securityLogs.length === 0"
                    class="rounded-lg border border-(--md-sys-color-outline-variant) p-6 text-center"
                >
                    <p class="text-(--md-sys-color-on-surface-variant)">
                        {{ t("logs.empty") }}
                    </p>
                </div>
                <div v-else class="space-y-2">
                    <div
                        v-for="log in securityLogs"
                        :key="log.id"
                        class="rounded-lg border border-(--md-sys-color-outline-variant) p-3 text-sm"
                    >
                        <div class="flex flex-wrap items-center gap-2">
                            <span
                                class="inline-flex items-center text-xs px-1.5 py-1 rounded font-semibold whitespace-nowrap"
                                :class="securityLevelClass(log.level)"
                            >
                                {{ (log.level || '').toUpperCase() }}
                            </span>
                            <span class="inline-flex items-center text-xs px-1.5 py-1 rounded bg-(--md-sys-color-surface-variant) text-(--md-sys-color-on-surface-variant) font-semibold whitespace-nowrap">{{ log.action }}</span>
                            <span class="ml-auto inline-flex items-center text-xs text-(--md-sys-color-on-surface-variant) whitespace-nowrap">
                                {{ formatDate(log.created_at) }}
                            </span>
                        </div>
                        <div class="mt-1 text-xs text-(--md-sys-color-on-surface-variant) break-all">
                            {{ log.message }}
                        </div>
                        <div class="mt-1 text-xs opacity-70 break-all">
                            <span v-if="log.method || log.path">{{ log.method }} {{ log.path }} · </span>
                            <span v-if="log.ip_address">ip: {{ log.ip_address }}</span>
                            <span v-if="log.username"> · user: {{ log.username }}</span>
                        </div>
                    </div>
                    <Pagination
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
                    @apply="reloadToken(1)"
                    :export-csv="() => exportLogs('token', 'csv')"
                    :export-json="() => exportLogs('token', 'json')"
                    :on-cleanup="() => onCleanup('token')"
                />
                <div v-if="loadingToken" class="flex justify-center py-6">
                    <AnzuProgressRing :size="40" />
                </div>
                <div
                    v-else-if="tokenLogs.length === 0"
                    class="rounded-lg border border-(--md-sys-color-outline-variant) p-6 text-center"
                >
                    <p class="text-(--md-sys-color-on-surface-variant)">
                        {{ t("logs.empty") }}
                    </p>
                </div>
                <div v-else class="space-y-2">
                    <div
                        v-for="log in tokenLogs"
                        :key="log.id"
                        class="rounded-lg border border-(--md-sys-color-outline-variant) p-3 text-sm"
                    >
                        <div class="flex flex-wrap items-center gap-2">
                            <span class="inline-flex items-center text-xs px-1.5 py-1 rounded bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container) font-semibold whitespace-nowrap">{{ log.action }}</span>
                            <span class="ml-auto inline-flex items-center text-xs text-(--md-sys-color-on-surface-variant) whitespace-nowrap">
                                {{ formatDate(log.created_at) }}
                            </span>
                        </div>
                        <div class="mt-1 text-xs text-(--md-sys-color-on-surface-variant) break-all">
                            <span>{{ log.token_name }} ({{ log.token_type }})</span>
                            <span v-if="log.ip_address"> · ip: {{ log.ip_address }}</span>
                            <span v-if="log.image_hash"> · img: {{ log.image_hash }}</span>
                        </div>
                        <div class="mt-1 text-xs opacity-70 break-all">{{ log.method }} {{ log.path }}</div>
                    </div>
                    <Pagination
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
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h, defineComponent } from "vue";
import AnzuTabs from "~/components/AnzuTabs.vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuSelector from "~/components/AnzuSelector.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import { useLogs } from "~/composables/useLogs";
import { useLogStream } from "~/composables/useLogStream";
import { useNotification } from "~/composables/useNotification";
import { useDialog } from "~/composables/useDialog";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import { formatDate } from "~/utils/format";
import type { AppLog, LogFilter, LogSource } from "~/types/app_log";
import type { SecurityLog } from "~/types/security_log";
import type { APITokenLog } from "~/types/api_token";

const { t } = useI18n();
const { notify } = useNotification();
const { confirm } = useDialog();
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

const onCleanup = async (source: LogSource) => {
    const days = await promptDays();
    if (!days) return;
    const ok = await confirm(t("logs.cleanupConfirm", { days, source }), {
        title: t("logs.cleanupTitle"),
        variant: DialogVariant.DESTRUCTIVE,
        actions: [
            { text: t("common.actions.cancel"), variant: "text" },
            { text: t("common.actions.confirm"), primary: true, variant: "filled" },
        ],
    });
    if (!ok) return;
    const res = await logs.cleanup(source, days);
    if (res) {
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

const promptDays = async (): Promise<number | null> => {
    const input = window.prompt(t("logs.cleanupPromptDays"), "30");
    if (input === null) return null;
    const n = Number(input);
    if (!n || n <= 0 || !Number.isInteger(n)) {
        notify({ message: t("logs.cleanupInvalidDays"), type: NotificationType.WARNING });
        return null;
    }
    return n;
};

const levelClass = (level: string) => {
    switch ((level || "").toUpperCase()) {
        case "DEBUG":
            return "bg-(--md-sys-color-surface-variant) text-(--md-sys-color-on-surface-variant)";
        case "WARN":
            return "bg-yellow-100 text-yellow-900";
        case "ERROR":
        case "FATAL":
            return "bg-(--md-sys-color-error-container) text-(--md-sys-color-on-error-container)";
        default:
            return "bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)";
    }
};

const securityLevelClass = (level: string) => {
    if (level === "warning" || level === "error") {
        return "bg-(--md-sys-color-error-container) text-(--md-sys-color-on-error-container)";
    }
    return "bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)";
};

onMounted(() => {
    reloadApp();
    reloadSecurity();
    reloadToken();
});

const LogToolbar = defineComponent({
    components: { AnzuInput, AnzuSelector, AnzuButton },
    props: {
        filter: { type: Object as () => LogFilter, required: true },
        showLevelFilter: { type: Boolean, default: false },
        showModuleFilter: { type: Boolean, default: false },
        enableStream: { type: Boolean, default: false },
        streaming: { type: Boolean, default: false },
        exportCsv: { type: Function as unknown as () => () => void, required: true },
        exportJson: { type: Function as unknown as () => () => void, required: true },
        onCleanup: { type: Function as unknown as () => () => void, required: true },
    },
    emits: ["apply", "update:streaming", "stream-toggle"],
    setup(p, { emit }) {
        const levelOptions = [
            { label: t("logs.levels.all"), value: "" },
            { label: "DEBUG", value: "debug" },
            { label: "INFO", value: "info" },
            { label: "WARN", value: "warn" },
            { label: "ERROR", value: "error" },
        ];
        const toggleStream = () => {
            const next = !p.streaming;
            emit("update:streaming", next);
            emit("stream-toggle", next);
        };
        return () =>
            h("div", { class: "mb-3 space-y-2" }, [
                h(
                    "div",
                    { class: "flex flex-wrap items-center gap-2" },
                    [
                        h(AnzuInput, {
                            modelValue: p.filter.search,
                            placeholder: t("common.actions.search"),
                            class: "flex-1 min-w-0",
                            "onUpdate:modelValue": (v: string) => (p.filter.search = v),
                            onKeyup: (e: KeyboardEvent) => {
                                if (e.key === "Enter") emit("apply");
                            },
                        }),
                        p.showModuleFilter
                            ? h(AnzuInput, {
                                  modelValue: p.filter.module,
                                  placeholder: t("logs.filters.module"),
                                  class: "flex-1 min-w-0",
                                  "onUpdate:modelValue": (v: string) => (p.filter.module = v),
                                  onKeyup: (e: KeyboardEvent) => {
                                      if (e.key === "Enter") emit("apply");
                                  },
                              })
                            : null,
                        h(AnzuInput, {
                            modelValue: p.filter.ip,
                            placeholder: t("logs.filters.ip"),
                            class: "flex-1 min-w-0",
                            "onUpdate:modelValue": (v: string) => (p.filter.ip = v),
                            onKeyup: (e: KeyboardEvent) => {
                                if (e.key === "Enter") emit("apply");
                            },
                        }),
                    ],
                ),
                p.showLevelFilter
                    ? h(AnzuSelector, {
                          modelValue: p.filter.level,
                          options: levelOptions,
                          "onUpdate:modelValue": (v: string) => {
                              p.filter.level = v;
                              emit("apply");
                          },
                      })
                    : null,
                h(
                    "div",
                    { class: "grid grid-cols-1 gap-2 sm:grid-cols-2" },
                    [
                        h(AnzuInput, {
                            modelValue: p.filter.start_date,
                            type: "date",
                            "onUpdate:modelValue": (v: string) => {
                                p.filter.start_date = v;
                                emit("apply");
                            },
                        }),
                        h(AnzuInput, {
                            modelValue: p.filter.end_date,
                            type: "date",
                            "onUpdate:modelValue": (v: string) => {
                                p.filter.end_date = v;
                                emit("apply");
                            },
                        }),
                    ],
                ),
                h(
                    "div",
                    { class: "flex flex-wrap items-center justify-end gap-2" },
                    [
                        p.enableStream
                            ? h(
                                  AnzuButton,
                                  {
                                      variant: p.streaming ? "filled" : "text",
                                      class: "whitespace-nowrap",
                                      onClick: toggleStream,
                                  },
                                  () =>
                                      p.streaming
                                          ? t("logs.streamStop")
                                          : t("logs.streamStart"),
                              )
                            : null,
                        h(
                            AnzuButton,
                            {
                                variant: "text",
                                class: "whitespace-nowrap",
                                onClick: () => emit("apply"),
                            },
                            () => t("logs.refresh"),
                        ),
                        h(
                            AnzuButton,
                            {
                                variant: "text",
                                class: "whitespace-nowrap",
                                onClick: p.exportCsv,
                            },
                            () => t("logs.exportCsv"),
                        ),
                        h(
                            AnzuButton,
                            {
                                variant: "text",
                                class: "whitespace-nowrap",
                                onClick: p.exportJson,
                            },
                            () => t("logs.exportJson"),
                        ),
                        h(
                            AnzuButton,
                            {
                                variant: "text",
                                class: "whitespace-nowrap text-(--md-sys-color-error)",
                                onClick: p.onCleanup,
                            },
                            () => t("logs.cleanup"),
                        ),
                    ],
                ),
            ]);
    },
});

const Pagination = defineComponent({
    components: { AnzuButton },
    props: {
        page: { type: Number, required: true },
        total: { type: Number, required: true },
        size: { type: Number, required: true },
        loading: { type: Boolean, default: false },
    },
    emits: ["update"],
    setup(p, { emit }) {
        const totalPages = computed(() => Math.max(1, Math.ceil(p.total / p.size)));
        return () =>
            totalPages.value > 1
                ? h(
                      "div",
                      {
                          class:
                              "mt-2 flex items-center justify-end gap-2 text-xs text-(--md-sys-color-on-surface-variant)",
                      },
                      [
                          h(
                              AnzuButton,
                              {
                                  variant: "text",
                                  disabled: p.page <= 1 || p.loading,
                                  onClick: () => emit("update", p.page - 1),
                              },
                              () => t("common.actions.paginationPrevious"),
                          ),
                          h("span", null, `${p.page} / ${totalPages.value}`),
                          h(
                              AnzuButton,
                              {
                                  variant: "text",
                                  disabled: p.page >= totalPages.value || p.loading,
                                  onClick: () => emit("update", p.page + 1),
                              },
                              () => t("common.actions.paginationNext"),
                          ),
                      ],
                  )
                : null;
    },
});
</script>
