<template>
  <div v-if="visible" class="fixed inset-0 z-50">
    <div class="fixed inset-0 bg-black/80"></div>
    <div class="fixed inset-0 flex flex-col" @click="handleClose">
      <AnzuButton @click.stop="handleClose" :aria-label="t('common.actions.close')"
        class="absolute top-4 left-4 w-10! h-10! p-0! min-w-0! rounded-full z-20">
        <XMarkIcon class="h-6 w-6" />
      </AnzuButton>
      <div class="flex-1 flex items-center justify-center p-4 overflow-hidden" @mouseup="stopDrag"
        @mouseleave="stopDrag">
        <div class="relative w-full h-full flex items-center justify-center" @wheel="handleWheel">
          <img v-if="displayImage" :src="`/i/${displayImage.hash}`" :alt="displayImage.file_name" ref="imageElement"
            class="max-h-full max-w-full object-contain transition-transform duration-0 origin-center will-change-transform"
            :class="isDragging ? 'cursor-grabbing' : 'cursor-grab'" @mousedown.prevent="startDrag" @mousemove="onDrag"
            @dblclick="resetZoom" @dragstart.prevent @click.stop @load="handleImageLoad" />
          <div v-if="!imageLoaded" class="absolute inset-0 flex items-center justify-center pointer-events-none">
            <AnzuProgressRing status="loading" :size="48" />
          </div>
        </div>
      </div>

      <!-- Info Box -->
      <div v-if="displayImage"
        class="bg-(--md-sys-color-surface-container) border-t border-(--md-sys-color-outline-variant) w-full h-72 shrink-0 z-20"
        @click.stop>
        <div class="h-full overflow-y-auto p-4 md:p-6 custom-scrollbar">
          <div class="max-w-7xl mx-auto">
            <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4 mb-4">
              <div class="min-w-0 flex-1 w-full">
                <h3 class="truncate text-xl font-semibold text-(--md-sys-color-on-surface)"
                  :title="displayImage.file_name">
                  {{ displayImage.file_name }}
                </h3>
                <div v-if="displayImage.uploaded_by_token_name"
                  class="mt-1 inline-flex items-center gap-2 text-xs text-(--md-sys-color-on-surface-variant)">
                  <span
                    class="px-1.5 py-0.5 rounded bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)">
                    {{ t("gallery.uploadedByToken") }}
                  </span>
                  <span class="font-medium">{{
                    displayImage.uploaded_by_token_name
                  }}</span>
                </div>
                <div class="flex flex-wrap items-center gap-4 mt-2 text-sm text-(--md-sys-color-on-surface-variant)">
                  <div class="flex items-center gap-1.5">
                    <DocumentIcon class="h-4 w-4" />
                    <span>{{
                      displayImage.size
                        ? formatFileSize(displayImage.size)
                        : "-"
                    }}</span>
                  </div>
                  <div class="flex items-center gap-1.5">
                    <ArrowsPointingOutIcon class="h-4 w-4" />
                    <span>{{
                      displayImage.width && displayImage.height
                        ? `${displayImage.width} × ${displayImage.height}`
                        : "-"
                    }}</span>
                  </div>
                  <div class="flex items-center gap-1.5">
                    <PhotoIcon class="h-4 w-4" />
                    <span>{{ mimeType }}</span>
                  </div>
                  <div class="flex items-center gap-1.5">
                    <CalendarIcon class="h-4 w-4" />
                    <span>{{
                      displayImage.created_at
                        ? formatDate(displayImage.created_at)
                        : "-"
                    }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-2 shrink-0 self-end sm:self-auto">
                <div class="flex items-center" v-if="!isEditing">
                  <AnzuButton :disabled="!hasPrevious" @click="handlePrevious" variant="text"
                    class="w-9! h-9! p-0! min-w-0! rounded-full">
                    <ChevronLeftIcon class="h-5 w-5" />
                  </AnzuButton>
                  <AnzuButton :disabled="!hasNext" @click="handleNext" variant="text"
                    class="w-9! h-9! p-0! min-w-0! rounded-full">
                    <ChevronRightIcon class="h-5 w-5" />
                  </AnzuButton>
                </div>
                <template v-if="!isEditing">
                  <AnzuButton @click="startEdit" variant="text" class="w-10! h-10! p-0! min-w-0! rounded-full"
                    title="Edit">
                    <PencilIcon class="h-5 w-5" />
                  </AnzuButton>
                  <AnzuButton @click="handleDelete" variant="text" class="w-10! h-10! p-0! min-w-0! rounded-full"
                    :title="t('common.actions.delete')">
                    <TrashIcon class="h-5 w-5" />
                  </AnzuButton>
                  <AnzuButton @click="handleCopyLink" variant="text" class="w-10! h-10! p-0! min-w-0! rounded-full"
                    :title="t('common.actions.copyLink')">
                    <LinkIcon class="h-5 w-5" />
                  </AnzuButton>
                  <AnzuButton @click="handleDownload" variant="text" class="w-10! h-10! p-0! min-w-0! rounded-full"
                    :title="t('common.actions.delete')">
                    <ArrowDownTrayIcon class="h-5 w-5" />
                  </AnzuButton>
                </template>
                <template v-else>
                  <AnzuButton @click="saveEdit" :loading="saving" variant="filled" class="h-9! px-4! rounded-full">
                    {{ t("common.actions.confirm") }}
                  </AnzuButton>
                  <AnzuButton @click="cancelEdit" variant="text" class="h-9! px-4! rounded-full">
                    {{ t("common.actions.cancel") }}
                  </AnzuButton>
                </template>
              </div>
            </div>

            <div class="border-t border-(--md-sys-color-outline-variant) my-4 opacity-50"></div>
            <div v-if="!isEditing" class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div class="md:col-span-2">
                <div v-if="loadingDetail" class="flex items-center">
                  <AnzuProgressRing status="loading" :size="20" />
                  <span
                    class="ml-2 text-sm text-(--md-sys-color-on-surface-variant)">{{ t("gallery.loadingDetails") }}</span>
                </div>
                <div v-else-if="displayImage.description"
                  class="text-sm text-(--md-sys-color-on-surface) whitespace-pre-wrap leading-relaxed">
                  {{ displayImage.description }}
                </div>
                <div v-else class="text-sm text-(--md-sys-color-on-surface-variant) italic">
                  {{ "No description" }}
                </div>
              </div>
              <div class="space-y-3">
                <div v-if="hasRoutes">
                  <span class="text-xs font-bold text-(--md-sys-color-primary) uppercase">Routes</span>
                  <div class="flex flex-wrap gap-2 mt-1">
                    <a v-for="route in detailedImage!.routes" :key="route" :href="`/i/r/${route}`" target="_blank"
                      class="px-2 py-1 rounded bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container) text-xs">
                      {{ route }}
                    </a>
                  </div>
                </div>

                <div>
                  <span class="text-xs font-bold text-(--md-sys-color-primary) uppercase">{{ "Tags" }}</span>
                  <div v-if="displayImage.tags && displayImage.tags.length > 0" class="mt-1">
                    <TagList :tags="displayImage.tags" class="text-sm" />
                  </div>
                  <div v-else class="mt-1 text-sm text-(--md-sys-color-on-surface-variant) italic">
                    No tags
                  </div>
                </div>
              </div>
            </div>

            <!-- Edit Form -->
            <div v-else class="grid gap-4">
              <AnzuInput v-model="editForm.file_name" :label="t('upload.customFileName')" />
              <AnzuInput v-model="editForm.description" :label="t('common.labels.description')" />
              <AnzuTags v-model="editForm.tags" :label="t('common.labels.tags')" />
              <AnzuTags v-model="editForm.routes" :label="t('upload.route')" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick } from "vue";
import AnzuButton from "./AnzuButton.vue";
import AnzuProgressRing from "./AnzuProgressRing.vue";
import AnzuAlert from "./AnzuAlert.vue";
import AnzuInput from "./AnzuInput.vue";
import AnzuTags from "./AnzuTags.vue";
import TagList from "./TagList.vue";
import {
  XMarkIcon,
  PencilIcon,
  CheckIcon,
  DocumentIcon,
  ArrowsPointingOutIcon,
  CalendarIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  LinkIcon,
  PhotoIcon,
  TrashIcon,
  ArrowDownTrayIcon,
} from "@heroicons/vue/24/outline";
import type {
  ImageModalProps,
  ImageModalEmits,
  ImageDetail,
} from "~/types/image";

import { formatDate, formatFileSize } from "~/utils/format";
const { t } = useI18n();

const props = withDefaults(defineProps<ImageModalProps>(), {
  image: null,
  visible: false,
  currentIndex: 0,
  totalImages: 0,
  hasPrevious: false,
  hasNext: false,
});

const emit = defineEmits<ImageModalEmits>();

const imageLoaded = ref(false);
const detailedImage = ref<ImageDetail | null>(null);
const loadingDetail = ref(false);
const detailError = ref<string | null>(null);
const imageElement = ref<HTMLImageElement | null>(null);

const scale = ref(1);
const position = ref({ x: 0, y: 0 });
const isDragging = ref(false);
const lastMousePosition = ref({ x: 0, y: 0 });

const isEditing = ref(false);
const editForm = ref({
  file_name: "",
  description: "",
  tags: [] as string[],
  routes: [] as string[],
});
const saving = ref(false);

const updateTransform = () => {
  if (imageElement.value) {
    imageElement.value.style.transform = `translate3d(${position.value.x}px, ${position.value.y}px, 0) scale(${scale.value})`;
  }
};

const resetZoom = () => {
  scale.value = 1;
  position.value = { x: 0, y: 0 };
  updateTransform();
};

watch(
  () => props.image,
  async (newImage) => {
    imageLoaded.value = false;
    detailedImage.value = null;
    detailError.value = null;
    isEditing.value = false;
    resetZoom();

    if (newImage && newImage.hash) {
      await fetchImageDetail(newImage.hash);
    }
  },
  { immediate: true },
);

const handleWheel = (e: WheelEvent) => {
  e.preventDefault();
  const delta = e.deltaY > 0 ? 0.9 : 1.1;
  const newScale = Math.max(0.1, Math.min(scale.value * delta, 10));
  scale.value = newScale;
  updateTransform();
};

const startDrag = (e: MouseEvent) => {
  isDragging.value = true;
  lastMousePosition.value = { x: e.clientX, y: e.clientY };
};

const onDrag = (e: MouseEvent) => {
  if (!isDragging.value) return;
  const deltaX = e.clientX - lastMousePosition.value.x;
  const deltaY = e.clientY - lastMousePosition.value.y;

  position.value.x += deltaX;
  position.value.y += deltaY;

  updateTransform();

  lastMousePosition.value = { x: e.clientX, y: e.clientY };
};

const stopDrag = () => {
  isDragging.value = false;
};

const handleImageLoad = () => {
  imageLoaded.value = true;
  resetZoom();
};

const startEdit = () => {
  if (!displayImage.value) return;
  editForm.value = {
    file_name: displayImage.value.file_name,
    description: displayImage.value.description || "",
    tags: [...(displayImage.value.tags || [])],
    routes: [...(detailedImage.value?.routes || [])],
  };
  isEditing.value = true;
};

const cancelEdit = () => {
  isEditing.value = false;
};

const saveEdit = async () => {
  if (!displayImage.value) return;
  saving.value = true;
  try {
    const updated = await $fetch<ImageDetail>(
      `/api/v1/images/${displayImage.value.hash}`,
      {
        method: "PATCH",
        body: {
          file_name: editForm.value.file_name,
          description: editForm.value.description,
          tags: editForm.value.tags,
          routes: editForm.value.routes,
        },
      },
    );

    if (detailedImage.value) {
      detailedImage.value = {
        ...detailedImage.value,
        ...updated,
        routes: editForm.value.routes,
      };
    }
    isEditing.value = false;
  } catch (e) {
    console.error(e);
  } finally {
    saving.value = false;
  }
};

const fetchImageDetail = async (hash: string) => {
  loadingDetail.value = true;
  detailError.value = null;

  try {
    const data = await $fetch<ImageDetail>(`/api/v1/images/${hash}/info`, {});
    detailedImage.value = data;
  } catch (error: any) {
    console.error("Failed to fetch image details:", error);
    detailError.value = error.data?.error || t("gallery.detailLoadFailed");
  } finally {
    loadingDetail.value = false;
  }
};

const displayImage = computed(() => {
  return detailedImage.value || props.image;
});

const hasRoutes = computed(() => {
  return detailedImage.value?.routes && detailedImage.value.routes.length > 0;
});

const mimeType = computed(() => {
  return (
    detailedImage.value?.mime_type || detailedImage.value?.mime || "Unknown"
  );
});

const handleClose = () => {
  emit("update:visible", false);
  emit("close");
};

const handlePrevious = () => {
  if (props.hasPrevious) {
    emit("previous");
  }
};

const handleNext = () => {
  if (props.hasNext) {
    emit("next");
  }
};

const handleCopyLink = () => {
  emit("copy-link");
};

const handleDownload = () => {
  emit("download");
};

const handleDelete = () => {
  if (props.image) {
    emit("delete", props.image.hash);
  }
};
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: var(--md-sys-color-outline-variant);
  border-radius: 3px;
}
</style>
