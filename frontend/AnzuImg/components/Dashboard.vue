<template>
  <div class="space-y-12 max-w-3xl mx-auto mb-12">
    
    <!-- 态势感知 -->
    <div>
      <h2 class="mb-4 text-xl font-semibold text-(--md-sys-color-on-surface)">
        {{ t("dashboard.situationalAwareness") }}
      </h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <!-- Total Images -->
        <div
          class="rounded-xl border border-(--md-sys-color-outline-variant) p-4 flex items-center gap-4"
        >
          <div
            class="rounded-full bg-(--md-sys-color-primary-container) p-3 text-(--md-sys-color-on-primary-container)"
          >
            <PhotoIcon class="w-6 h-6" />
          </div>
          <div>
            <div class="text-2xl font-bold text-(--md-sys-color-on-surface)">
              {{ loading ? "-" : stats?.total_images || 0 }}
            </div>
            <div class="text-sm text-(--md-sys-color-on-surface-variant)">
              {{ t("dashboard.totalImages") }}
            </div>
          </div>
        </div>

        <!-- Total Size -->
        <div
          class="rounded-xl border border-(--md-sys-color-outline-variant) p-4 flex items-center gap-4"
        >
          <div
            class="rounded-full bg-(--md-sys-color-secondary-container) p-3 text-(--md-sys-color-on-secondary-container)"
          >
            <ServerStackIcon class="w-6 h-6" />
          </div>
          <div>
            <div class="text-2xl font-bold text-(--md-sys-color-on-surface)">
              {{ loading ? "-" : formatFileSize(stats?.total_size || 0) }}
            </div>
            <div class="text-sm text-(--md-sys-color-on-surface-variant)">
              {{ t("dashboard.totalSize") }}
            </div>
          </div>
        </div>

        <!-- Security Events -->
        <div
          class="rounded-xl border border-(--md-sys-color-outline-variant) p-4 flex items-center gap-4 sm:col-span-2"
        >
          <div
            class="rounded-full bg-(--md-sys-color-tertiary-container) p-3 text-(--md-sys-color-on-tertiary-container)"
          >
            <ShieldCheckIcon class="w-6 h-6" />
          </div>
          <div>
            <div class="text-2xl font-bold text-(--md-sys-color-on-surface)">
              {{ loading ? "-" : stats?.security_events_24h || 0 }}
            </div>
            <div class="text-sm text-(--md-sys-color-on-surface-variant)">
              {{ t("dashboard.securityEvents24h") }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 风险处置 -->
    <div v-if="!loading">
      <h2 class="mb-4 text-xl font-semibold text-(--md-sys-color-on-surface)">
        {{ t("dashboard.riskDisposition") }}
      </h2>

      <!-- 有风险 -->
      <div v-if="(stats?.login_failures_24h || 0) > 0" class="grid grid-cols-1 gap-4">
        <div
          class="rounded-xl border border-(--md-sys-color-error) p-4 flex items-center gap-4 bg-(--md-sys-color-error-container) text-(--md-sys-color-on-error-container)"
        >
          <div
            class="rounded-full bg-(--md-sys-color-error) p-3 text-(--md-sys-color-on-error)"
          >
            <NoSymbolIcon class="w-6 h-6" />
          </div>
          <div>
            <div class="text-2xl font-bold">
              {{ stats?.login_failures_24h || 0 }}
            </div>
            <div class="text-sm font-medium opacity-90">
              {{ t("dashboard.loginFailures24h") }}
            </div>
          </div>
        </div>
      </div>

      <!-- 无风险 -->
      <div v-else class="rounded-xl border border-(--md-sys-color-outline-variant) p-8 text-center">
         <p class="text-(--md-sys-color-on-surface-variant)">
            {{ t("dashboard.safeSystem") }}
         </p>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { 
    PhotoIcon, 
    ServerStackIcon, 
    ShieldCheckIcon, 
    NoSymbolIcon
} from "@heroicons/vue/24/outline";
import { useStats } from "~/composables/useStats";
import { formatFileSize } from "~/utils/format";
import type { SystemStats } from "~/types/stats";

const { t } = useI18n();
const { fetchStats } = useStats();
const stats = ref<SystemStats | null>(null);
const loading = ref(true);

onMounted(async () => {
  loading.value = true;
  try {
      stats.value = await fetchStats();
  } catch (e) {
      console.error("Failed to fetch stats", e);
  } finally {
      loading.value = false;
  }
});
</script>
