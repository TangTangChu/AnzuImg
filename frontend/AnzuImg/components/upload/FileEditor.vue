<template>
  <div class="flex-1 flex flex-col min-h-0 h-full">
    <div
      v-if="allDone && totalFiles > 0"
      class="px-4 py-3 flex flex-wrap items-center gap-x-4 gap-y-2"
    >
      <span class="text-xs uppercase tracking-wide text-(--md-sys-color-on-surface-variant)">
        {{ t("upload.summary.label") }}
      </span>
      <span class="flex items-center gap-1 text-sm text-(--md-sys-color-on-surface)">
        <CheckCircleIcon class="w-4 h-4 text-green-600 dark:text-green-400" />
        {{ t("upload.summary.success", { count: successCount }) }}
      </span>
      <span
        v-if="failedCount > 0"
        class="flex items-center gap-1 text-sm text-(--md-sys-color-on-surface)"
      >
        <XCircleIcon class="w-4 h-4 text-(--md-sys-color-error)" />
        {{ t("upload.summary.failed", { count: failedCount }) }}
      </span>
      <div class="ml-auto flex items-center gap-2">
        <AnzuButton
          v-if="successCount > 0"
          variant="tonal"
          @click="$emit('copy-all')"
        >
          <template #icon>
            <ClipboardIcon class="w-4 h-4" />
          </template>
          {{ t("upload.summary.copyAll") }}
        </AnzuButton>
        <AnzuButton variant="text" @click="$emit('clear')">
          {{ t("upload.result.startOver") }}
        </AnzuButton>
      </div>
    </div>

    <div
      v-if="selectedFile"
      class="p-4 flex gap-4 items-center"
    >
      <div
        class="h-16 w-16 rounded-lg overflow-hidden shrink-0 flex items-center justify-center bg-(--md-sys-color-surface-container-lowest)"
      >
        <template v-if="hasLocalMedia(selectedFile)">
          <img
            v-if="!isVideoFile(selectedFile.file!)"
            :src="selectedFile.previewUrl"
            class="w-full h-full object-contain"
          />
          <video
            v-else
            :src="selectedFile.previewUrl"
            class="w-full h-full object-contain"
            controls
            playsinline
            preload="metadata"
          ></video>
        </template>
        <LinkIcon
          v-else
          class="w-7 h-7 text-(--md-sys-color-on-surface-variant)"
        />
      </div>
      <div class="flex-1 min-w-0">
        <h3 class="font-bold truncate text-(--md-sys-color-on-surface)">
          {{ selectedFile.displayName }}
        </h3>
        <p
          v-if="hasLocalMedia(selectedFile)"
          class="text-xs text-(--md-sys-color-on-surface-variant)"
        >
          {{ formatFileSize(selectedFile.displaySize) }} ·
          {{ selectedFile.displayMime }}<template v-if="selectedFile.displayWidth && selectedFile.displayHeight">
            · {{ selectedFile.displayWidth }}x{{ selectedFile.displayHeight }}
          </template>
        </p>
        <p
          v-else
          class="text-xs text-(--md-sys-color-on-surface-variant) truncate"
          :title="selectedFile.sourceUrl"
        >
          {{ t("upload.url.serverItemHint") }} · {{ selectedFile.sourceUrl }}
        </p>
      </div>
    </div>

    <template v-if="selectedFile">
      <div
        v-if="selectedFile.status === 'success' && selectedFile.resultUrl"
        class="px-6 pb-4 space-y-2"
      >
        <div class="flex items-center gap-2">
          <CheckCircleIcon class="w-4 h-4 shrink-0 text-green-600 dark:text-green-400" />
          <span class="text-sm font-bold text-(--md-sys-color-on-surface)">{{ t("upload.success") }}</span>
          <span class="text-xs text-(--md-sys-color-on-surface-variant)">{{ t("upload.result.linkLabel") }}</span>
        </div>
        <div class="flex items-center gap-2">
          <AnzuInput
            :model-value="selectedFile.resultUrl"
            readonly
            class="flex-1 min-w-0 font-mono"
            @focus="selectAllText"
          />
          <AnzuSplitButton
            variant="tonal"
            class="shrink-0"
            @click="copyCurrentLink('url')"
          >
            <template #icon>
              <ClipboardIcon class="w-4 h-4" />
            </template>
            <template #menu="{ close }">
              <AnzuButton variant="text" size="sm" class="w-full justify-start" @click="copyCurrentLink('url'); close()">
                {{ t("common.actions.copyUrl") }}
              </AnzuButton>
              <AnzuButton variant="text" size="sm" class="w-full justify-start" @click="copyCurrentLink('markdown'); close()">
                {{ t("common.actions.copyMarkdown") }}
              </AnzuButton>
            </template>
          </AnzuSplitButton>
          <AnzuButton
            variant="text"
            class="shrink-0"
            :title="t('upload.result.openLink')"
            :href="selectedFile.resultUrl"
            target="_blank"
          >
            <template #icon>
              <ArrowTopRightOnSquareIcon class="w-4 h-4" />
            </template>
          </AnzuButton>
        </div>
      </div>

      <div
        v-else-if="selectedFile.status === 'error'"
        class="px-6 pb-4"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="font-bold text-sm flex items-center gap-2 text-(--md-sys-color-on-surface)">
              <XCircleIcon class="w-4 h-4 text-(--md-sys-color-error)" />
              {{ t("upload.result.failed") }}
            </p>
            <p class="text-xs mt-1 text-(--md-sys-color-on-surface-variant) break-words">{{ selectedFile.error }}</p>
          </div>
          <AnzuButton
            variant="tonal"
            size="sm"
            class="shrink-0"
            :disabled="uploading"
            @click="$emit('retry-current')"
          >
            <template #icon>
              <ArrowPathIcon class="w-4 h-4" />
            </template>
            {{ t("upload.actions.retry") }}
          </AnzuButton>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto px-6 pb-4 space-y-4">
        <div class="grid grid-cols-2 gap-3">
          <AnzuInput
            v-model="selectedFile.customName"
            :label="t('upload.customFileName')"
            :placeholder="t('upload.customFileNamePlaceholder')"
            class="col-span-2"
          />

          <AnzuInput
            v-model="selectedFile.description"
            :label="t('common.labels.description')"
          />

          <AnzuComboBox
            v-model="selectedTagOption"
            :items="tagItems"
            :placeholder="t('tags.selectPlaceholder')"
            :aria-label="t('tags.selectLabel')"
            @change="handleTagPick"
          />

          <AnzuTags
            v-model="selectedFile.tags"
            :label="t('common.labels.tags')"
            :max-tags="10"
          />

          <AnzuTags
            v-model="selectedFile.routes"
            :label="t('upload.route')"
            :max-tags="5"
          />
        </div>

        <AnzuCheckbox
          v-if="files.length > 1"
          v-model="applyToAll"
          :label="t('upload.applyToAll')"
        />
      </div>
    </template>

    <div
      v-if="!selectedFile"
      class="flex-1 flex items-center justify-center text-(--md-sys-color-on-surface-variant)"
    >
      Select a media file to edit details
    </div>
    <div class="p-4">
      <div class="mb-4">
        <p class="mb-2 text-xs font-medium text-(--md-sys-color-on-surface-variant)">
          {{ t("upload.convertSection") }}
        </p>
        <div class="flex items-center gap-2 mb-2">
          <AnzuCheckbox
            :model-value="enableConvert"
            @update:model-value="$emit('update:enableConvert', $event)"
            :label="t('upload.convert') + ' (All)'"
            :disabled="hasVideoFile"
          />
        </div>
        <p
          v-if="hasVideoFile"
          class="mb-2 text-xs text-(--md-sys-color-on-surface-variant)"
        >
          {{ t("upload.videoConvertDisabled") }}
        </p>
        <div
          v-if="enableConvert"
          class="flex flex-col gap-4 text-sm animate-fade-in-up"
        >
          <div class="w-full">
            <label
              class="text-xs text-(--md-sys-color-on-surface-variant) block mb-1"
              >Format</label
            >
            <AnzuSelector
              :model-value="targetFormat"
              @update:model-value="$emit('update:targetFormat', String($event))"
              :options="formatOptions"
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label
                class="text-xs text-(--md-sys-color-on-surface-variant) block mb-1"
                >{{ t("upload.quality") }}</label
              >
              <AnzuInput
                :model-value="quality"
                @update:model-value="$emit('update:quality', String($event))"
                type="number"
                placeholder="80"
              />
            </div>
            <div>
              <label
                class="text-xs text-(--md-sys-color-on-surface-variant) block mb-1"
                >{{ t("upload.effort") }}</label
              >
              <AnzuInput
                :model-value="effort"
                @update:model-value="$emit('update:effort', String($event))"
                type="number"
                placeholder="4"
              />
            </div>
          </div>
        </div>
      </div>
      <AnzuButton
        @click="$emit('upload')"
        :status="uploading ? 'loading' : 'default'"
        class="w-full"
        :disabled="uploading || !hasFiles"
      >
        {{ t("upload.submit") }} ({{ totalFiles }})
      </AnzuButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import {
  CheckCircleIcon,
  XCircleIcon,
  ClipboardIcon,
  ArrowTopRightOnSquareIcon,
  ArrowPathIcon,
  LinkIcon,
} from "@heroicons/vue/24/outline";
import { formatFileSize } from "~/utils/format";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuSplitButton from "~/components/AnzuSplitButton.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuTags from "~/components/AnzuTags.vue";
import AnzuCheckbox from "~/components/AnzuCheckbox.vue";
import AnzuComboBox from "~/components/AnzuComboBox.vue";
import AnzuSelector from "~/components/AnzuSelector.vue";
import { useNotification } from "~/composables/useNotification";
import { NotificationType } from "~/types/notification";
import type { UploadFileItem } from "~/types/upload";
import type { TagSummary } from "~/types/image";

const props = defineProps<{
  selectedFile: UploadFileItem | null;
  files: UploadFileItem[];
  tagList: TagSummary[];
  enableConvert: boolean;
  targetFormat: string;
  quality: string;
  effort: string;
  hasVideoFile: boolean;
  uploading: boolean;
  hasFiles: boolean;
  totalFiles: number;
  successCount: number;
  failedCount: number;
  allDone: boolean;
}>();

const emit = defineEmits<{
  (e: "update:enableConvert", value: boolean): void;
  (e: "update:targetFormat", value: string): void;
  (e: "update:quality", value: string): void;
  (e: "update:effort", value: string): void;
  (e: "upload"): void;
  (e: "clear"): void;
  (e: "copy-all"): void;
  (e: "retry-current"): void;
}>();

const { t } = useI18n();
const { notify } = useNotification();

const formatOptions = [
  { label: "webp", value: "webp" },
  { label: "avif", value: "avif" },
];

const selectedTagOption = ref<string | null>(null);
const applyToAll = ref(false);


watch(
  () => props.selectedFile?.tags,
  (newTags) => {
    if (applyToAll.value && newTags && props.files) {
      props.files.forEach((f) => {
        if (f !== props.selectedFile) {
          f.tags = [...newTags];
        }
      });
    }
  },
  { deep: true },
);

watch(applyToAll, (val) => {
  if (val && props.selectedFile && props.files) {
    props.files.forEach((f) => {
      if (f !== props.selectedFile) {
        f.tags = [...props.selectedFile.tags];
      }
    });
  }
});

const tagItems = computed(() =>
  props.tagList.map((item) => ({
    value: item.tag,
    label: `${item.tag} (${item.count})`,
  }))
);

const handleTagPick = (value: string | number | null) => {
  if (!value || !props.selectedFile) return;
  const tag = String(value);
  if (!props.selectedFile.tags.includes(tag)) {
    props.selectedFile.tags.push(tag);
  }
  selectedTagOption.value = null;
};

const isVideoFile = (file: File) => file.type.startsWith("video/");
const hasLocalMedia = (item: UploadFileItem) =>
  item.file !== null && !!item.previewUrl;

const selectAllText = (e: FocusEvent) => {
  const target = e.target as HTMLInputElement | null;
  target?.select();
};

const copyCurrentLink = async (format: "url" | "markdown" = "url") => {
  const url = props.selectedFile?.resultUrl;
  if (!url) return;
  const name = props.selectedFile?.displayName || "image";
  const text = format === "markdown" ? `![${name}](${url})` : url;
  try {
    await navigator.clipboard.writeText(text);
    notify({
      message: t("common.actions.copySuccess"),
      type: NotificationType.SUCCESS,
    });
  } catch {
    // clipboard API unavailable
  }
};
</script>

<style scoped>
@reference "tailwindcss";

.animate-fade-in-up {
  animation: fadeInUp 0.3s ease-out;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
