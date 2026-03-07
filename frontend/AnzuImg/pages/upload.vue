<template>
  <div class="h-full flex flex-col max-w-6xl mx-auto w-full">
    <h1 class="mb-6 text-3xl font-bold text-center">{{ t("upload.title") }}</h1>
    <div
      v-if="files.length === 0"
      class="flex-1 flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-(--md-sys-color-outline-variant) transition-colors min-h-100 relative group p-12 cursor-pointer"
      :class="[isDragging ? 'border-(--md-sys-color-primary)' : '']"
      @dragenter.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @dragover.prevent
      @drop.prevent="handleDrop"
      @click="triggerMainInput"
    >
      <input
        type="file"
        ref="fileInput"
        class="hidden"
        @change="handleFileSelect"
        accept="image/*,video/*"
        multiple
      />

      <svg
        class="mx-auto mb-4 h-16 w-16 text-(--md-sys-color-primary)"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
        />
      </svg>
      <p class="mb-2 text-xl font-medium text-(--md-sys-color-on-surface)">
        {{ t("upload.dragDrop") }}
      </p>
      <p class="text-sm text-(--md-sys-color-on-surface-variant)">
        {{ t("upload.orSelect") }}
      </p>
    </div>
    <div
      v-else
      class="flex flex-col lg:flex-row gap-6 flex-1 min-h-0 animate-fade-in-up"
    >
      <!-- Left Panel: File Grid -->
      <div
        class="lg:w-1/2 flex flex-col min-h-125 lg:h-150 rounded-xl overflow-hidden border border-(--md-sys-color-outline-variant) relative transition-colors"
        :class="[
          isDragging
            ? 'border-dashed border-(--md-sys-color-primary) bg-(--md-sys-color-primary)/5'
            : '',
        ]"
        @dragenter.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @dragover.prevent
        @drop.prevent="handleDrop"
      >
        <div
          v-if="isDragging"
          class="absolute inset-0 z-50 flex items-center justify-center bg-(--md-sys-color-surface)/80 pointer-events-none"
        >
          <p class="text-xl font-medium text-(--md-sys-color-primary)">
            {{ t("upload.dragDrop") }}
          </p>
        </div>

        <div
          class="p-4 border-b border-(--md-sys-color-outline-variant) flex justify-between items-center"
        >
          <span class="font-medium"
            >{{ files.length }} {{ t("common.labels.files") }} ({{
              formatFileSize(totalSize)
            }})</span
          >
          <AnzuButton variant="text" @click="clearAll">{{
            t("common.actions.clear")
          }}</AnzuButton>
        </div>

        <div class="flex-1 overflow-y-auto p-4">
          <div class="grid grid-cols-3 sm:grid-cols-4 gap-3">
            <FileGridItem
              v-for="(item, index) in files"
              :key="index"
              :item="item"
              :index="index"
              :is-selected="selectedIndex === index"
              @select="selectFile"
            />
            <div
              class="relative aspect-square rounded-lg border-2 border-dashed border-(--md-sys-color-outline-variant) flex items-center justify-center cursor-pointer hover:bg-(--md-sys-color-surface-container) hover:border-(--md-sys-color-primary) transition-colors"
              @click="triggerAddInput"
            >
              <input
                type="file"
                ref="addInput"
                class="hidden"
                @change="handleAddFile"
                accept="image/*,video/*"
                multiple
              />
              <svg
                class="w-8 h-8 text-(--md-sys-color-on-surface-variant)"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 4v16m8-8H4"
                />
              </svg>
            </div>
          </div>
        </div>
      </div>
      <div
        class="lg:w-1/2 flex flex-col h-auto lg:h-150 rounded-xl border border-(--md-sys-color-outline-variant) overflow-hidden"
      >
        <FileEditor
          :selected-file="selectedFile"
          :tag-list="tagList?.data ?? []"
          v-model:enable-convert="enableConvert"
          v-model:target-format="targetFormat"
          v-model:quality="quality"
          v-model:effort="effort"
          :has-video-file="hasVideoFile"
          :uploading="uploading"
          :has-files="files.length > 0"
          :total-files="files.length"
          @upload="startUpload"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted, watch } from "vue";
import { useAuth } from "~/composables/useAuth";
import { formatFileSize } from "~/utils/format";
import AnzuButton from "~/components/AnzuButton.vue";
import { useNotification } from "~/composables/useNotification";
import { NotificationType } from "~/types/notification";
import { parseApiError } from "~/utils/api-error";
import type { TagListResponse } from "~/types/image";
import type { UploadFileItem, UploadResultItem } from "~/types/upload";
import FileGridItem from "~/components/upload/FileGridItem.vue";
import FileEditor from "~/components/upload/FileEditor.vue";

const { t } = useI18n();
useAuth();
const { notify } = useNotification();
const { apiUrl } = useApi();

const isDragging = ref(false);
const uploading = ref(false);
const files = ref<UploadFileItem[]>([]);
const selectedIndex = ref<number>(0);
const fileInput = ref<HTMLInputElement | null>(null);
const addInput = ref<HTMLInputElement | null>(null);

const enableConvert = ref(false);
const targetFormat = ref("webp");
const quality = ref("");
const effort = ref("");

const { data: tagList } = await useFetch<TagListResponse>(
  apiUrl("/api/v1/tags")
);

const selectedFile = computed(() => {
  if (files.value.length === 0) return null;
  return files.value[selectedIndex.value] || null;
});

const totalSize = computed(() => {
  return files.value.reduce((acc, item) => acc + item.file.size, 0);
});

const isVideoFile = (file: File) => file.type.startsWith("video/");

const hasVideoFile = computed(() =>
  files.value.some((item) => isVideoFile(item.file))
);

watch(hasVideoFile, (value) => {
  if (value) {
    enableConvert.value = false;
  }
});

onUnmounted(() => {
  files.value.forEach((item) => URL.revokeObjectURL(item.previewUrl));
});

const triggerMainInput = () => {
  fileInput.value?.click();
};

const triggerAddInput = () => {
  addInput.value?.click();
};

const processFiles = (newFiles: FileList | null) => {
  if (!newFiles) return;
  const newItems = Array.from(newFiles).map((file) => ({
    file,
    previewUrl: URL.createObjectURL(file),
    description: "",
    tags: [],
    routes: [],
    customName: "",
    status: "pending" as const,
  }));

  const startIndex = files.value.length;
  files.value = [...files.value, ...newItems];
  if (newItems.length > 0 && startIndex === 0) {
    selectedIndex.value = 0;
  }
};

const handleDrop = (e: DragEvent) => {
  isDragging.value = false;
  processFiles(e.dataTransfer?.files || null);
};

const handleFileSelect = (e: Event) => {
  const input = e.target as HTMLInputElement;
  processFiles(input.files);
  input.value = "";
};

const handleAddFile = (e: Event) => {
  const input = e.target as HTMLInputElement;
  processFiles(input.files);
  input.value = "";
};

const selectFile = (index: number) => {
  selectedIndex.value = index;
};

const clearAll = () => {
  files.value.forEach((item) => URL.revokeObjectURL(item.previewUrl));
  files.value = [];
  selectedIndex.value = 0;
};

const startUpload = async () => {
  if (files.value.length === 0) return;

  uploading.value = true;

  // Reset
  files.value.forEach((f) => {
    f.status = "pending";
    f.error = undefined;
    f.resultUrl = undefined;
  });

  const formData = new FormData();
  // Append all
  files.value.forEach((item) => {
    formData.append("file", item.file);
  });

  // Append metadata
  const metadata = files.value.map((f, index) => ({
    client_index: index,
    description: f.description,
    tags: f.tags,
    routes: f.routes,
    custom_name: f.customName,
  }));
  formData.append("metadata", JSON.stringify(metadata));

  // Global settings
  if (enableConvert.value && !hasVideoFile.value) {
    formData.append("convert", "true");
    formData.append("target_format", targetFormat.value);
    if (quality.value) formData.append("quality", quality.value);
    if (effort.value) formData.append("effort", effort.value);
  }

  try {
    const data = await $fetch<UploadResultItem[]>(apiUrl("/api/v1/images"), {
      method: "POST",
      body: formData,
    });

    if (Array.isArray(data)) {
      let successCount = 0;
      const claimed = new Set<number>();
      let fallbackCursor = 0;

      const claimFallbackIndex = (): number => {
        while (
          fallbackCursor < files.value.length &&
          claimed.has(fallbackCursor)
        ) {
          fallbackCursor++;
        }
        return fallbackCursor < files.value.length ? fallbackCursor++ : -1;
      };

      data.forEach((res) => {
        let targetIndex = -1;
        if (
          typeof res.client_index === "number" &&
          Number.isInteger(res.client_index) &&
          res.client_index >= 0 &&
          res.client_index < files.value.length &&
          !claimed.has(res.client_index)
        ) {
          targetIndex = res.client_index;
        } else {
          targetIndex = claimFallbackIndex();
        }

        if (targetIndex < 0 || !files.value[targetIndex]) return;

        const targetItem = files.value[targetIndex];
        if (!targetItem) return;

        claimed.add(targetIndex);

        if (res.success) {
          targetItem.status = "success";
          const rawUrl = typeof res.url === "string" ? res.url : "";
          try {
            targetItem.resultUrl = new URL(
              rawUrl,
              window.location.origin
            ).toString();
          } catch {
            targetItem.resultUrl = rawUrl || `${window.location.origin}/`;
          }
          successCount++;
        } else {
          targetItem.status = "error";
          targetItem.error = res.message || "Unknown error";
        }
      });

      files.value.forEach((f, index) => {
        if (!claimed.has(index) && f.status === "pending") {
          f.status = "error";
          f.error = "Upload failed";
        }
      });

      if (successCount === files.value.length) {
        notify({
          message: t("upload.success"),
          type: NotificationType.SUCCESS,
        });
      } else {
        notify({
          message: `Uploaded with ${files.value.length - successCount} errors`,
          type: NotificationType.WARNING,
        });
      }
    }
  } catch (e: any) {
    const parsed = parseApiError(e, "Upload failed");
    notify({
      message: parsed.displayMessage,
      type: NotificationType.ERROR,
    });
  } finally {
    uploading.value = false;
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
