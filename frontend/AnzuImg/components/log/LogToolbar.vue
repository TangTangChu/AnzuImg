<template>
    <div class="mb-3 space-y-2">
        <div class="flex flex-wrap items-center gap-2">
            <AnzuInput
                v-model="filter.search"
                :placeholder="t('common.actions.search')"
                class="flex-1 min-w-0"
                @keyup.enter="$emit('apply')"
            />
            <AnzuInput
                v-if="showModuleFilter"
                v-model="filter.module"
                :placeholder="t('logs.filters.module')"
                class="flex-1 min-w-0"
                @keyup.enter="$emit('apply')"
            />
            <AnzuInput
                v-model="filter.ip"
                :placeholder="t('logs.filters.ip')"
                class="flex-1 min-w-0"
                @keyup.enter="$emit('apply')"
            />
        </div>

        <AnzuSelector
            v-if="showLevelFilter"
            v-model="filter.level"
            :options="levelOptions"
            @update:model-value="$emit('apply')"
        />

        <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
            <AnzuInput
                v-model="filter.start_date"
                type="date"
                @update:model-value="$emit('apply')"
            />
            <AnzuInput
                v-model="filter.end_date"
                type="date"
                @update:model-value="$emit('apply')"
            />
        </div>

        <div class="flex flex-wrap items-center justify-end gap-2">
            <AnzuButton
                v-if="enableStream"
                :variant="streaming ? 'filled' : 'text'"
                class="whitespace-nowrap"
                @click="toggleStream"
            >
                {{ streaming ? t("logs.streamStop") : t("logs.streamStart") }}
            </AnzuButton>
            <AnzuButton
                variant="text"
                class="whitespace-nowrap"
                @click="$emit('apply')"
            >
                {{ t("logs.refresh") }}
            </AnzuButton>
            <AnzuButton
                variant="text"
                class="whitespace-nowrap"
                @click="exportCsv"
            >
                {{ t("logs.exportCsv") }}
            </AnzuButton>
            <AnzuButton
                variant="text"
                class="whitespace-nowrap"
                @click="exportJson"
            >
                {{ t("logs.exportJson") }}
            </AnzuButton>
            <AnzuButton
                variant="text"
                class="whitespace-nowrap text-(--md-sys-color-error)"
                @click="onCleanup"
            >
                {{ t("logs.cleanup") }}
            </AnzuButton>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuSelector from "~/components/AnzuSelector.vue";
import AnzuButton from "~/components/AnzuButton.vue";
import type { LogFilter } from "~/types/app_log";

const props = withDefaults(
    defineProps<{
        filter: LogFilter;
        showLevelFilter?: boolean;
        showModuleFilter?: boolean;
        enableStream?: boolean;
        streaming?: boolean;
        exportCsv: () => void;
        exportJson: () => void;
        onCleanup: () => void;
    }>(),
    {
        showLevelFilter: false,
        showModuleFilter: false,
        enableStream: false,
        streaming: false,
    },
);

const emit = defineEmits<{
    (e: "apply"): void;
    (e: "update:streaming", value: boolean): void;
    (e: "stream-toggle", value: boolean): void;
}>();

const { t } = useI18n();

const levelOptions = computed(() => [
    { label: t("logs.levels.all"), value: "" },
    { label: "DEBUG", value: "debug" },
    { label: "INFO", value: "info" },
    { label: "WARN", value: "warn" },
    { label: "ERROR", value: "error" },
]);

const toggleStream = () => {
    const next = !props.streaming;
    emit("update:streaming", next);
    emit("stream-toggle", next);
};
</script>
