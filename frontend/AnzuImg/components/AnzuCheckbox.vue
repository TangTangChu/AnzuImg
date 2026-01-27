<template>
  <label class="flex items-center gap-2 cursor-pointer select-none group">
    <div class="relative flex items-center justify-center w-5 h-5 rounded border transition-colors" :class="[
      modelValue
        ? 'bg-(--md-sys-color-primary) border-(--md-sys-color-primary)'
        : 'bg-transparent border-(--md-sys-color-outline) group-hover:border-(--md-sys-color-primary)'
    ]">
      <input type="checkbox" :checked="modelValue" @change="updateInput"
        class="absolute inset-0 opacity-0 cursor-pointer" />
      <svg v-if="modelValue" class="w-3.5 h-3.5 text-(--md-sys-color-on-primary) pointer-events-none"
        viewBox="0 0 14 14" fill="none">
        <path d="M11.6666 3.5L5.24992 9.91667L2.33325 7" stroke="currentColor" stroke-width="2" stroke-linecap="round"
          stroke-linejoin="round" />
      </svg>
    </div>
    <span v-if="label" class="text-sm font-medium text-(--md-sys-color-on-surface)">{{ label }}</span>
  </label>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: boolean;
  label?: string;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
}>();

const updateInput = (event: Event) => {
  const target = event.target as HTMLInputElement;
  emit('update:modelValue', target.checked);
};
</script>

<style scoped>
@reference "tailwindcss";
</style>
