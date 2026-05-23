<template>
  <div
    class="relative aspect-square rounded-lg overflow-hidden border-2 cursor-pointer transition-all group select-none"
    :class="[
      isSelected
        ? 'border-(--md-sys-color-primary) ring-2 ring-(--md-sys-color-primary/20)'
        : 'border-transparent hover:border-(--md-sys-color-outline)',
      item.status === 'error' ? 'border-(--md-sys-color-error)!' : '',
      item.status === 'success' ? 'border-green-500!' : '',
    ]"
    @click="$emit('select', index)"
  >
    <template v-if="hasLocalMedia">
      <img
        v-if="!isVideoFile(item.file!)"
        :src="item.previewUrl"
        class="w-full h-full object-cover"
        draggable="false"
      />
      <video
        v-else
        :src="item.previewUrl"
        class="w-full h-full object-cover"
        muted
        playsinline
        preload="metadata"
      ></video>
    </template>
    <div
      v-else
      class="w-full h-full flex flex-col items-center justify-center border border-dashed border-(--md-sys-color-outline-variant) text-(--md-sys-color-on-surface-variant) px-2 text-center"
    >
      <LinkIcon class="w-7 h-7 mb-1" />
      <p class="text-[10px] leading-tight break-all line-clamp-2">
        {{ hostLabel }}
      </p>
      <p class="text-[9px] opacity-70 mt-0.5">{{ t("upload.url.serverItemHint") }}</p>
    </div>

    <div
      v-if="item.status !== 'pending'"
      class="absolute inset-0 flex items-center justify-center bg-black/40 pointer-events-none"
    >
      <CheckIcon v-if="item.status === 'success'" class="w-8 h-8 text-green-400" />
      <XMarkIcon v-else-if="item.status === 'error'" class="w-8 h-8 text-red-400" />
    </div>

    <div
      class="absolute top-1.5 right-1.5 flex items-center gap-1 transition-opacity"
      :class="actionsVisible ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'"
    >
      <button
        v-if="item.status === 'success' && item.resultUrl"
        type="button"
        class="h-6 w-6 rounded-full bg-black/55 text-white flex items-center justify-center hover:bg-black/75 transition-colors"
        :title="t('common.actions.copyLink')"
        @click.stop="$emit('copy', index)"
      >
        <ClipboardIcon class="w-3.5 h-3.5" />
      </button>
      <button
        v-if="item.status === 'error'"
        type="button"
        class="h-6 w-6 rounded-full bg-black/55 text-white flex items-center justify-center hover:bg-black/75 transition-colors"
        :title="t('upload.actions.retry')"
        @click.stop="$emit('retry', index)"
      >
        <ArrowPathIcon class="w-3.5 h-3.5" />
      </button>
      <button
        type="button"
        class="h-6 w-6 rounded-full bg-black/55 text-white flex items-center justify-center hover:bg-(--md-sys-color-error) transition-colors"
        :title="t('upload.actions.removeItem')"
        @click.stop="$emit('remove', index)"
      >
        <XMarkIcon class="w-3.5 h-3.5" />
      </button>
    </div>

    <div
      class="absolute top-1 left-1 bg-black/50 text-white text-xs px-1.5 py-0.5 rounded transition-opacity"
      :class="actionsVisible ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'"
    >
      {{ index + 1 }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  LinkIcon,
  CheckIcon,
  XMarkIcon,
  ClipboardIcon,
  ArrowPathIcon,
} from "@heroicons/vue/24/outline";
import type { UploadFileItem } from "~/types/upload";

const props = defineProps<{
  item: UploadFileItem;
  index: number;
  isSelected: boolean;
}>();

defineEmits<{
  (e: "select", index: number): void;
  (e: "copy", index: number): void;
  (e: "retry", index: number): void;
  (e: "remove", index: number): void;
}>();

const { t } = useI18n();

const hasLocalMedia = computed(
  () => props.item.file !== null && !!props.item.previewUrl,
);

const isVideoFile = (file: File) => file.type.startsWith("video/");

const actionsVisible = computed(() => props.isSelected);

const hostLabel = computed(() => {
  const url = props.item.sourceUrl;
  if (!url) return "";
  try {
    return new URL(url).host;
  } catch {
    return url;
  }
});
</script>
