<template>
  <div class="h-full flex flex-col max-w-6xl mx-auto w-full">
    <h1 class="mb-4 text-3xl font-bold text-center">{{ t("upload.title") }}</h1>

    <div v-if="files.length === 0" class="flex flex-col lg:flex-row gap-4">
      <div
        class="flex-1 flex flex-col items-center justify-center rounded-lg bg-(--md-sys-color-surface-container-lowest) transition-all min-h-60 lg:min-h-72 relative group p-8 cursor-pointer"
        :class="[isDragging ? 'ring-2 ring-(--md-sys-color-primary)' : '']"
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
        <CloudArrowUpIcon class="h-14 w-14 mb-3 text-(--md-sys-color-primary)" />
        <p class="mb-1 text-lg font-medium text-(--md-sys-color-on-surface)">
          {{ t("upload.dragDrop") }}
        </p>
        <p class="text-sm text-(--md-sys-color-on-surface-variant)">
          {{ t("upload.orSelect") }}
        </p>
      </div>

      <div
        class="flex-1 p-6 rounded-lg"
      >
        <UrlSourcePanel
          :server-mode-acknowledged="serverModeAcknowledged"
          :loading="urlLoading"
          @update:server-mode-acknowledged="(v) => (serverModeAcknowledged = v)"
          @add="handleAddUrl"
        />
      </div>
    </div>

    <div v-else class="flex flex-col lg:flex-row gap-3 mb-4">
      <div
        class="flex-1 flex items-center gap-3 rounded-lg bg-(--md-sys-color-surface-container-lowest) px-3 py-2 cursor-pointer transition-colors hover:bg-black/5 dark:hover:bg-white/5"
        :class="[isDragging ? 'ring-2 ring-(--md-sys-color-primary)' : '']"
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
        <CloudArrowUpIcon class="h-5 w-5 shrink-0 text-(--md-sys-color-primary)" />
        <p class="text-sm text-(--md-sys-color-on-surface-variant) truncate">
          {{ t("upload.dragDrop") }} · {{ t("upload.orSelect") }}
        </p>
      </div>

      <div class="flex-1">
        <UrlSourcePanel
          compact
          :server-mode-acknowledged="serverModeAcknowledged"
          :loading="urlLoading"
          @update:server-mode-acknowledged="(v) => (serverModeAcknowledged = v)"
          @add="handleAddUrl"
        />
      </div>
    </div>

    <div
      v-if="files.length > 0"
      class="flex flex-col lg:flex-row gap-4 flex-1 min-h-0 animate-fade-in-up"
    >
      <div
        class="lg:w-1/2 flex flex-col min-h-125 lg:h-150 relative transition-colors"
        :class="[isDragging ? 'ring-2 ring-inset ring-(--md-sys-color-primary)' : '']"
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
          class="px-2 py-3 flex justify-between items-center"
        >
          <span class="font-medium text-sm"
            >{{ files.length }} {{ t("common.labels.files") }} ({{
              formatFileSize(totalSize)
            }})</span
          >
          <AnzuButton variant="text" size="sm" @click="clearAll">{{
            t("common.actions.clear")
          }}</AnzuButton>
        </div>

        <div class="flex-1 overflow-y-auto py-3 px-2">
          <div class="grid grid-cols-3 sm:grid-cols-4 gap-3">
            <FileGridItem
              v-for="(item, index) in files"
              :key="index"
              :item="item"
              :index="index"
              :is-selected="selectedIndex === index"
              :uploading="uploading"
              @select="selectFile"
              @copy="copySingleLink"
              @retry="retryItem"
              @remove="removeItem"
            />
            <div
              class="relative aspect-square rounded-lg bg-(--md-sys-color-surface-container-lowest) flex items-center justify-center cursor-pointer hover:bg-black/5 dark:hover:bg-white/5 transition-colors"
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
              <PlusIcon class="w-8 h-8 text-(--md-sys-color-on-surface-variant)" />
            </div>
          </div>
        </div>
      </div>

      <div
        class="lg:w-1/2 flex flex-col h-auto lg:h-150 overflow-hidden"
      >
        <FileEditor
          :selected-file="selectedFile"
          :files="files"
          :tag-list="tagList?.data ?? []"
          v-model:enable-convert="enableConvert"
          v-model:target-format="targetFormat"
          v-model:quality="quality"
          v-model:effort="effort"
          :has-video-file="hasVideoFile"
          :uploading="uploading"
          :has-files="files.length > 0"
          :total-files="files.length"
          :success-count="successCount"
          :failed-count="failedCount"
          :all-done="allDone"
          @upload="startUpload"
          @clear="clearAll"
          @copy-all="copyAllSuccessLinks"
          @retry-current="retryCurrent"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted, watch } from "vue";
import { CloudArrowUpIcon, PlusIcon } from "@heroicons/vue/24/outline";
import { useAuth } from "~/composables/useAuth";
import { formatFileSize } from "~/utils/format";
import AnzuButton from "~/components/AnzuButton.vue";
import { useNotification } from "~/composables/useNotification";
import { NotificationType } from "~/types/notification";
import { parseApiError } from "~/utils/api-error";
import type { TagListResponse } from "~/types/image";
import type {
  UploadFileItem,
  UploadResultItem,
  UrlSourceMetadata,
} from "~/types/upload";
import FileGridItem from "~/components/upload/FileGridItem.vue";
import FileEditor from "~/components/upload/FileEditor.vue";
import UrlSourcePanel from "~/components/upload/UrlSourcePanel.vue";

const { t } = useI18n();
useAuth();
const { notify } = useNotification();
const { apiUrl } = useApi();

const isDragging = ref(false);
const uploading = ref(false);
const urlLoading = ref(false);
const files = ref<UploadFileItem[]>([]);
const selectedIndex = ref<number>(0);
const fileInput = ref<HTMLInputElement | null>(null);
const addInput = ref<HTMLInputElement | null>(null);

const serverModeAcknowledged = ref(false);

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
  return files.value.reduce((acc, item) => acc + item.displaySize, 0);
});

const isVideoFile = (file: File) => file.type.startsWith("video/");

const loadMediaDimensions = (item: UploadFileItem) => {
  if (!item.file || !item.previewUrl) return;
  const url = item.previewUrl;
  if (isVideoFile(item.file)) {
    const el = document.createElement("video");
    el.preload = "metadata";
    el.onloadedmetadata = () => {
      item.displayWidth = el.videoWidth;
      item.displayHeight = el.videoHeight;
      el.src = "";
    };
    el.src = url;
  } else {
    const img = new Image();
    img.onload = () => {
      item.displayWidth = img.naturalWidth;
      item.displayHeight = img.naturalHeight;
    };
    img.src = url;
  }
};

const hasVideoFile = computed(() =>
  files.value.some((item) => item.file !== null && isVideoFile(item.file)),
);

const successCount = computed(
  () => files.value.filter((f) => f.status === "success").length,
);
const failedCount = computed(
  () => files.value.filter((f) => f.status === "error").length,
);
const allDone = computed(
  () =>
    files.value.length > 0 &&
    files.value.every((f) => f.status !== "pending"),
);

watch(hasVideoFile, (value) => {
  if (value) {
    enableConvert.value = false;
  }
});

onUnmounted(() => {
  files.value.forEach((item) => {
    if (item.previewUrl) URL.revokeObjectURL(item.previewUrl);
  });
});

const triggerMainInput = () => {
  fileInput.value?.click();
};

const triggerAddInput = () => {
  addInput.value?.click();
};

const processFiles = (newFiles: FileList | File[] | null) => {
  if (!newFiles) return;
  const arr = Array.isArray(newFiles) ? newFiles : Array.from(newFiles);
  if (arr.length === 0) return;
  const newItems: UploadFileItem[] = arr.map((file) => ({
    file,
    previewUrl: URL.createObjectURL(file),
    description: "",
    tags: [],
    routes: [],
    customName: "",
    status: "pending" as const,
    source: "file",
    displayName: file.name,
    displaySize: file.size,
    displayMime: file.type,
  }));

  const startIndex = files.value.length;
  files.value = [...files.value, ...newItems];
  if (newItems.length > 0 && startIndex === 0) {
    selectedIndex.value = 0;
  }
  newItems.forEach(loadMediaDimensions);
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
  files.value.forEach((item) => {
    if (item.previewUrl) URL.revokeObjectURL(item.previewUrl);
  });
  files.value = [];
  selectedIndex.value = 0;
};

const removeItem = (index: number) => {
  if (uploading.value) return;
  const target = files.value[index];
  if (!target) return;
  if (target.previewUrl) URL.revokeObjectURL(target.previewUrl);
  files.value = files.value.filter((_, i) => i !== index);
  if (files.value.length === 0) {
    selectedIndex.value = 0;
    return;
  }
  if (selectedIndex.value >= files.value.length) {
    selectedIndex.value = files.value.length - 1;
  } else if (selectedIndex.value > index) {
    selectedIndex.value = selectedIndex.value - 1;
  }
};

const urlBasename = (url: string): string => {
  try {
    const parsed = new URL(url);
    const segments = parsed.pathname.split("/").filter(Boolean);
    if (segments.length > 0) {
      const last = segments[segments.length - 1];
      try {
        return decodeURIComponent(last);
      } catch {
        return last;
      }
    }
    return parsed.host;
  } catch {
    return url;
  }
};

const handleAddUrl = async (url: string, mode: "browser" | "server") => {
  if (mode === "browser") {
    await addBrowserUrl(url);
  } else {
    addServerUrl(url);
  }
};

const addBrowserUrl = async (url: string) => {
  urlLoading.value = true;
  try {
    const resp = await fetch(url, { mode: "cors" });
    if (!resp.ok) {
      throw new Error(`HTTP ${resp.status}`);
    }
    const blob = await resp.blob();
    const name = urlBasename(url) || "remote-file";
    const file = new File([blob], name, {
      type: blob.type || "application/octet-stream",
    });
    processFiles([file]);
  } catch {
    notify({
      message: t("upload.url.fetchFailed"),
      type: NotificationType.ERROR,
    });
  } finally {
    urlLoading.value = false;
  }
};

const addServerUrl = (url: string) => {
  const item: UploadFileItem = {
    file: null,
    previewUrl: "",
    description: "",
    tags: [],
    routes: [],
    customName: "",
    status: "pending",
    source: "url-server",
    sourceUrl: url,
    displayName: urlBasename(url) || url,
    displaySize: 0,
    displayMime: "",
  };
  const startIndex = files.value.length;
  files.value = [...files.value, item];
  if (startIndex === 0) {
    selectedIndex.value = 0;
  }
};

const resetItemStatus = (item: UploadFileItem) => {
  item.status = "pending";
  item.error = undefined;
  item.resultUrl = undefined;
};

const uploadItems = async (indices: number[]) => {
  if (indices.length === 0) return;
  const targets = indices
    .map((i) => ({ index: i, item: files.value[i] }))
    .filter((entry): entry is { index: number; item: UploadFileItem } => !!entry.item);
  if (targets.length === 0) return;

  uploading.value = true;
  targets.forEach(({ item }) => resetItemStatus(item));

  const localTargets = targets.filter(({ item }) => item.file !== null);
  const urlTargets = targets.filter(({ item }) => item.source === "url-server");

  const formData = new FormData();
  const fileMetadata = localTargets.map(({ index, item }) => ({
    client_index: index,
    description: item.description,
    tags: item.tags,
    routes: item.routes,
    custom_name: item.customName,
  }));

  localTargets.forEach(({ item }) => {
    formData.append("file", item.file as File);
  });
  if (fileMetadata.length > 0) {
    formData.append("metadata", JSON.stringify(fileMetadata));
  }

  if (urlTargets.length > 0) {
    const urlSources: UrlSourceMetadata[] = urlTargets.map(({ index, item }) => ({
      url: item.sourceUrl as string,
      client_index: index,
      description: item.description,
      tags: item.tags,
      routes: item.routes,
      custom_name: item.customName,
    }));
    formData.append("url_sources", JSON.stringify(urlSources));
  }

  if (enableConvert.value && !hasVideoFile.value) {
    formData.append("convert", "true");
    formData.append("target_format", targetFormat.value);
    if (quality.value) formData.append("quality", quality.value);
    if (effort.value) formData.append("effort", effort.value);
  }

  const targetIndexSet = new Set(targets.map((t) => t.index));

  try {
    const data = await $fetch<UploadResultItem[]>(apiUrl("/api/v1/images"), {
      method: "POST",
      body: formData,
    });

    if (Array.isArray(data)) {
      let successResultCount = 0;
      const claimed = new Set<number>();
      const remaining = targets.map((t) => t.index);
      let fallbackCursor = 0;

      const claimFallbackIndex = (): number => {
        while (
          fallbackCursor < remaining.length &&
          claimed.has(remaining[fallbackCursor] ?? -1)
        ) {
          fallbackCursor++;
        }
        const next = remaining[fallbackCursor];
        if (next === undefined) return -1;
        fallbackCursor++;
        return next;
      };

      data.forEach((res) => {
        let targetIndex = -1;
        if (
          typeof res.client_index === "number" &&
          Number.isInteger(res.client_index) &&
          res.client_index >= 0 &&
          res.client_index < files.value.length &&
          targetIndexSet.has(res.client_index) &&
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
              window.location.origin,
            ).toString();
          } catch {
            targetItem.resultUrl = rawUrl || `${window.location.origin}/`;
          }
          successResultCount++;
        } else {
          targetItem.status = "error";
          targetItem.error = res.message || "Unknown error";
        }
      });

      targets.forEach(({ index }) => {
        const item = files.value[index];
        if (item && !claimed.has(index) && item.status === "pending") {
          item.status = "error";
          item.error = "Upload failed";
        }
      });

      if (successResultCount === targets.length) {
        notify({
          message: t("upload.success"),
          type: NotificationType.SUCCESS,
        });
      } else {
        notify({
          message: `Uploaded with ${targets.length - successResultCount} errors`,
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
    targets.forEach(({ item }) => {
      if (item.status === "pending") {
        item.status = "error";
        item.error = parsed.displayMessage;
      }
    });
  } finally {
    uploading.value = false;
  }
};

const startUpload = () =>
  uploadItems(
    files.value.reduce<number[]>((acc, item, i) => {
      if (item.status !== "success") acc.push(i);
      return acc;
    }, []),
  );

const retryItem = (index: number) => {
  if (uploading.value) return;
  uploadItems([index]);
};

const retryCurrent = () => {
  if (uploading.value) return;
  if (selectedIndex.value < 0 || selectedIndex.value >= files.value.length) return;
  uploadItems([selectedIndex.value]);
};

const copySingleLink = async (index: number) => {
  const url = files.value[index]?.resultUrl;
  if (!url) return;
  try {
    await navigator.clipboard.writeText(url);
    notify({
      message: t("common.actions.copySuccess"),
      type: NotificationType.SUCCESS,
    });
  } catch {
    notify({
      message: url,
      type: NotificationType.INFO,
    });
  }
};

const copyAllSuccessLinks = async () => {
  const urls = files.value
    .filter((f) => f.status === "success" && f.resultUrl)
    .map((f) => f.resultUrl as string);
  if (urls.length === 0) {
    notify({
      message: t("upload.summary.copyAllEmpty"),
      type: NotificationType.INFO,
    });
    return;
  }
  try {
    await navigator.clipboard.writeText(urls.join("\n"));
    notify({
      message: t("upload.summary.copyAllDone", { count: urls.length }),
      type: NotificationType.SUCCESS,
    });
  } catch {
    notify({
      message: urls.join("\n"),
      type: NotificationType.INFO,
    });
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
