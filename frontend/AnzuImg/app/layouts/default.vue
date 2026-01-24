<template>
  <div class="min-h-screen bg-(--md-sys-color-surface-container-lowest) text-(--md-sys-color-on-background) transition-colors duration-300">
    <AnzuNavBar />
    <main class="container mx-auto p-4 pt-24 md:p-6 md:pt-28">
      <slot />
    </main>
    <AnzuNotification ref="notificationRef" />
    <AnzuDialogContainer />
  </div>
</template>

<script setup lang="ts">
import AnzuNavBar from '~/components/AnzuNavBar.vue';
import AnzuNotification from '~/components/AnzuNotification.vue';
import AnzuDialogContainer from '~/components/AnzuDialogContainer.vue';
import { useNotification } from '~/composables/useNotification';

const { notificationRef } = useNotification();
</script>
