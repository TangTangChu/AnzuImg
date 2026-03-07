<template>
  <div
    class="relative aspect-square rounded-lg overflow-hidden border-2 cursor-pointer transition-all group select-none"
    :class="[
      isSelected
        ? 'border-(--md-sys-color-primary) ring-2 ring-(--md-sys-color-primary/20)'
        : 'border-transparent hover:border-(--md-sys-color-outline)',
      item.status === 'error' ? 'border-red-500!' : '',
      item.status === 'success' ? 'border-green-500!' : '',
    ]"
    @click="$emit('select', index)"
  >
    <img
      v-if="!isVideoFile(item.file)"
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
    <div
      class="absolute inset-0 flex items-center justify-center bg-black/40"
      v-if="item.status !== 'pending'"
    >
      <div v-if="item.status === 'success'" class="text-green-400">
        <svg
          class="w-8 h-8"
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
      </div>
      <div v-if="item.status === 'error'" class="text-red-400">
        <svg
          class="w-8 h-8"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </div>
    </div>

    <div
      class="absolute inset-0 bg-primary/10 opacity-0 group-hover:opacity-100 transition-opacity"
      :class="{ 'opacity-100': isSelected }"
    ></div>

    <div
      class="absolute top-1 left-1 bg-black/50 text-white text-xs px-1.5 py-0.5 rounded opacity-0 group-hover:opacity-100 transition-opacity"
      :class="{ 'opacity-100': isSelected }"
    >
      {{ index + 1 }}
    </div>
  </div>
</template>

<script setup lang="ts">
import type { UploadFileItem } from "~/types/upload";

const props = defineProps<{
  item: UploadFileItem;
  index: number;
  isSelected: boolean;
}>();

defineEmits<{
  (e: "select", index: number): void;
}>();

const isVideoFile = (file: File) => file.type.startsWith("video/");
</script>
