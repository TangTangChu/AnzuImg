<template>
  <div class="my-6 flex items-center gap-2">
    <div class="h-px flex-1 bg-(--md-sys-color-outline-variant)"></div>

    <span class="text-xs text-(--md-sys-color-outline) px-2 whitespace-nowrap">
      <slot />
    </span>

    <div class="h-px flex-1 bg-(--md-sys-color-outline-variant)"></div>
  </div>
</template>