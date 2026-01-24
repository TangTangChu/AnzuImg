<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useAuth } from '~/composables/useAuth';

const { checkInit } = useAuth();
const router = useRouter();
const route = useRoute();
const checking = ref(true);

onMounted(async () => {
  try {
    const initialized = await checkInit();
    
    if (!initialized && route.path !== '/setup') {
      router.push('/setup');
    } else if (initialized && route.path === '/setup') {
      router.push('/login');
    }
  } catch (error) {
    console.error('Failed to check initialization status:', error);
  } finally {
    checking.value = false;
  }
});
</script>
