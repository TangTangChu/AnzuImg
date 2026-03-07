<template>
  <div class="flex-1 flex flex-col min-h-0 h-full">
    <div
      v-if="selectedFile"
      class="p-4 border-b border-(--md-sys-color-outline-variant) flex gap-4 items-center"
    >
      <div
        class="h-16 w-16 rounded overflow-hidden shrink-0 border border-(--md-sys-color-outline-variant)"
      >
        <img
          v-if="!isVideoFile(selectedFile.file)"
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
      </div>
      <div class="flex-1 min-w-0">
        <h3 class="font-bold truncate text-(--md-sys-color-on-surface)">
          {{ selectedFile.file.name }}
        </h3>
        <p class="text-xs text-(--md-sys-color-on-surface-variant)">
          {{ formatFileSize(selectedFile.file.size) }} ·
          {{ selectedFile.file.type }}
        </p>
      </div>
    </div>

    <div v-if="selectedFile" class="flex-1 overflow-y-auto p-6 space-y-4">
      <AnzuInput
        v-model="selectedFile.customName"
        :label="t('upload.customFileName')"
        :placeholder="t('upload.customFileNamePlaceholder')"
      />

      <AnzuInput
        v-model="selectedFile.description"
        :label="t('common.labels.description')"
      />

      <div class="flex items-center gap-2">
        <AnzuComboBox
          v-model="selectedTagOption"
          :items="tagItems"
          :placeholder="t('tags.selectPlaceholder')"
          :aria-label="t('tags.selectLabel')"
          @change="handleTagPick"
        />
        <AnzuButton
          class="shrink-0 whitespace-nowrap"
          variant="tonal"
          :disabled="!selectedTagOption"
          @click="addSelectedTag"
        >
          {{ t("tags.add") }}
        </AnzuButton>
      </div>

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
    <div
      v-if="selectedFile?.status === 'success'"
      class="p-4 border-t border-(--md-sys-color-outline-variant) bg-(--md-sys-color-primary-container)/40 text-(--md-sys-color-on-primary-container)"
    >
      <p class="font-bold text-sm flex items-center gap-2">
        <svg
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 13l4 4L19 7"
          />
        </svg>
        {{ t("upload.success") }}
      </p>
      <a
        :href="selectedFile.resultUrl"
        target="_blank"
        class="text-xs underline break-all mt-1 block"
        >{{ selectedFile.resultUrl }}</a
      >
    </div>
    <div
      v-else-if="selectedFile?.status === 'error'"
      class="p-4 bg-red-500/10 border-t border-red-500/20 text-red-600"
    >
      <p class="font-bold text-sm flex items-center gap-2">
        <svg
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        Upload Failed
      </p>
      <p class="text-xs mt-1">{{ selectedFile.error }}</p>
    </div>
    <div
      v-if="!selectedFile"
      class="flex-1 flex items-center justify-center text-(--md-sys-color-on-surface-variant)"
    >
      Select a media file to edit details
    </div>
    <div class="p-4 border-t border-(--md-sys-color-outline-variant)">
      <div
        class="mb-4 p-3 rounded-lg border border-(--md-sys-color-outline-variant)"
      >
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
            <AnzuComboBox
              :model-value="targetFormat"
              @update:model-value="$emit('update:targetFormat', String($event))"
              :items="['webp', 'avif']"
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
import { ref, computed } from "vue";
import { formatFileSize } from "~/utils/format";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuTags from "~/components/AnzuTags.vue";
import AnzuCheckbox from "~/components/AnzuCheckbox.vue";
import AnzuComboBox from "~/components/AnzuComboBox.vue";
import type { UploadFileItem } from "~/types/upload";
import type { TagSummary } from "~/types/image";

const props = defineProps<{
  selectedFile: UploadFileItem | null;
  tagList: TagSummary[];
  enableConvert: boolean;
  targetFormat: string;
  quality: string;
  effort: string;
  hasVideoFile: boolean;
  uploading: boolean;
  hasFiles: boolean;
  totalFiles: number;
}>();

const emit = defineEmits<{
  (e: "update:enableConvert", value: boolean): void;
  (e: "update:targetFormat", value: string): void;
  (e: "update:quality", value: string): void;
  (e: "update:effort", value: string): void;
  (e: "upload"): void;
}>();

const { t } = useI18n();

const selectedTagOption = ref<string | null>(null);

const tagItems = computed(() =>
  props.tagList.map((item) => ({
    value: item.tag,
    label: `${item.tag} (${item.count})`,
  }))
);

const handleTagPick = (value: string | number | null) => {
  selectedTagOption.value = value ? String(value) : null;
};

const addSelectedTag = () => {
  if (!props.selectedFile || !selectedTagOption.value) return;
  if (!props.selectedFile.tags.includes(selectedTagOption.value)) {
    props.selectedFile.tags.push(selectedTagOption.value);
  }
  selectedTagOption.value = null;
};

const isVideoFile = (file: File) => file.type.startsWith("video/");
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
