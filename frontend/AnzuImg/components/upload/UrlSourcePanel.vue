<template>
  <div :class="compact ? 'flex flex-wrap items-center gap-2' : 'flex flex-col gap-3'">
    <div :class="compact ? 'flex-1 min-w-50' : 'w-full'">
      <AnzuInput
        v-model="urlInput"
        type="url"
        :label="compact ? '' : t('upload.tabs.url')"
        :placeholder="compact ? t('upload.url.compactPlaceholder') : t('upload.url.placeholder')"
        :disabled="loading"
        @keydown.enter.prevent="submit"
      />
    </div>

    <div v-if="!compact" class="flex flex-col gap-2">
      <label class="flex items-start gap-2 cursor-pointer text-sm">
        <input
          type="radio"
          name="url-source-mode"
          value="browser"
          v-model="mode"
          class="mt-1 accent-(--md-sys-color-primary)"
          :disabled="loading"
        />
        <span class="flex-1">
          <span class="block text-(--md-sys-color-on-surface)">{{ t("upload.url.modeBrowser") }}</span>
          <span class="block text-xs text-(--md-sys-color-on-surface-variant)">{{ t("upload.url.modeBrowserHint") }}</span>
        </span>
      </label>
      <label class="flex items-start gap-2 cursor-pointer text-sm">
        <input
          type="radio"
          name="url-source-mode"
          value="server"
          v-model="mode"
          class="mt-1 accent-(--md-sys-color-primary)"
          :disabled="loading"
          @change="handleServerModeSelect"
        />
        <span class="flex-1">
          <span class="block text-(--md-sys-color-on-surface)">{{ t("upload.url.modeServer") }}</span>
          <span class="block text-xs text-(--md-sys-color-on-surface-variant)">{{ t("upload.url.modeServerHint") }}</span>
        </span>
      </label>
    </div>

    <div v-else class="inline-flex rounded-lg border border-(--md-sys-color-outline-variant) overflow-hidden">
      <button
        type="button"
        class="px-3 h-9 text-xs transition-colors"
        :class="mode === 'browser'
          ? 'bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)'
          : 'text-(--md-sys-color-on-surface-variant) hover:bg-(--md-sys-color-on-surface)/5'"
        :disabled="loading"
        @click="selectMode('browser')"
      >
        {{ t("upload.url.modeBrowser") }}
      </button>
      <button
        type="button"
        class="px-3 h-9 text-xs flex items-center gap-1 border-l border-(--md-sys-color-outline-variant) transition-colors"
        :class="mode === 'server'
          ? 'bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)'
          : 'text-(--md-sys-color-on-surface-variant) hover:bg-(--md-sys-color-on-surface)/5'"
        :disabled="loading"
        @click="selectMode('server')"
      >
        <ExclamationTriangleIcon v-if="mode === 'server' && serverModeAcknowledged" class="w-3.5 h-3.5 text-(--md-sys-color-error)" />
        {{ t("upload.url.modeServer") }}
      </button>
    </div>

    <div
      v-if="!compact && mode === 'server' && serverModeAcknowledged"
      class="flex items-start gap-2 text-xs text-(--md-sys-color-on-surface-variant)"
    >
      <ExclamationTriangleIcon class="w-4 h-4 mt-0.5 shrink-0 text-(--md-sys-color-error)" />
      <span>{{ t("upload.url.modeServerHint") }}</span>
    </div>

    <AnzuButton
      :variant="compact ? 'tonal' : 'filled'"
      :disabled="!canAdd"
      :status="loading ? 'loading' : 'default'"
      :class="compact ? '' : 'w-full'"
      @click="submit"
    >
      {{ loading ? t("upload.url.loading") : compact ? t("upload.actions.addUrl") : t("upload.url.add") }}
    </AnzuButton>

    <AnzuDialog
      v-model:visible="riskDialogVisible"
      :title="t('upload.url.riskTitle')"
      :actions="[
        { text: t('upload.url.riskCancel'), variant: 'text', handler: cancelRisk },
        { text: t('upload.url.riskAck'), primary: true, variant: 'filled', handler: confirmRisk },
      ]"
    >
      <p class="text-(--md-sys-color-on-surface-variant) text-sm leading-relaxed">
        {{ t("upload.url.riskBody") }}
      </p>
    </AnzuDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { ExclamationTriangleIcon } from "@heroicons/vue/24/outline";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuDialog from "~/components/AnzuDialog.vue";
import { useNotification } from "~/composables/useNotification";
import { NotificationType } from "~/types/notification";

const props = withDefaults(
  defineProps<{
    serverModeAcknowledged: boolean;
    loading: boolean;
    compact?: boolean;
  }>(),
  { compact: false },
);

const emit = defineEmits<{
  (e: "add", url: string, mode: "browser" | "server"): void;
  (e: "update:serverModeAcknowledged", value: boolean): void;
}>();

const { t } = useI18n();
const { notify } = useNotification();

const urlInput = ref("");
const mode = ref<"browser" | "server">("browser");
const riskDialogVisible = ref(false);
const pendingServerSwitch = ref(false);

const canAdd = computed(() => {
  if (props.loading) return false;
  const value = urlInput.value.trim();
  if (!value) return false;
  return isValidHttpUrl(value);
});

function isValidHttpUrl(value: string): boolean {
  try {
    const parsed = new URL(value);
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch {
    return false;
  }
}

function selectMode(target: "browser" | "server") {
  if (props.loading) return;
  if (mode.value === target) return;
  mode.value = target;
  if (target === "server") {
    handleServerModeSelect();
  }
}

function handleServerModeSelect() {
  if (mode.value !== "server") return;
  if (props.serverModeAcknowledged) return;
  pendingServerSwitch.value = true;
  riskDialogVisible.value = true;
}

function cancelRisk() {
  riskDialogVisible.value = false;
  if (pendingServerSwitch.value) {
    mode.value = "browser";
    pendingServerSwitch.value = false;
  }
}

function confirmRisk() {
  emit("update:serverModeAcknowledged", true);
  riskDialogVisible.value = false;
  pendingServerSwitch.value = false;
}

watch(riskDialogVisible, (visible) => {
  if (!visible && pendingServerSwitch.value && !props.serverModeAcknowledged) {
    mode.value = "browser";
    pendingServerSwitch.value = false;
  }
});

function submit() {
  const value = urlInput.value.trim();
  if (!value) return;
  if (!isValidHttpUrl(value)) {
    notify({
      message: t("upload.url.invalid"),
      type: NotificationType.ERROR,
    });
    return;
  }
  emit("add", value, mode.value);
  urlInput.value = "";
}
</script>
