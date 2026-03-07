<template>
  <div class="min-h-screen bg-(--md-sys-color-surface-container-lowest) text-(--md-sys-color-on-background) transition-colors duration-300">
    <AnzuNavBar />
    <main class="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-4 pt-24 md:pb-6 md:pt-28">
      <slot />
    </main>
    <AnzuNotification ref="notificationRef" />
    <AnzuDialogContainer />
    <PageFooter />
  </div>
</template>

<script setup lang="ts">
import AnzuNavBar from '~/components/AnzuNavBar.vue';
import AnzuNotification from '~/components/AnzuNotification.vue';
import AnzuDialogContainer from '~/components/AnzuDialogContainer.vue';
import { useNotification } from '~/composables/useNotification';

const { notificationRef } = useNotification();
</script>
