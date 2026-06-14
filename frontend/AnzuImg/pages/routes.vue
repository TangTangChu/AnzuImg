<template>
  <h1 class="mb-6 text-3xl font-bold text-center">{{ t("routes.title") }}</h1>

  <div class="max-w-6xl mx-auto">
    <div v-if="routes.length === 0 && !loading" class="rounded-lg bg-black/5 p-6 text-center dark:bg-white/5">
      <p class="text-(--md-sys-color-on-surface-variant)">No routes found</p>
    </div>

    <div v-else-if="routes.length === 0 && loading" class="flex justify-center py-12">
      <AnzuProgressRing :size="48" />
    </div>

    <div v-else>
      <div
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4"
        :class="loading ? 'opacity-50 pointer-events-none' : ''"
      >
        <div
          v-for="item in routes"
          :key="item.id"
          class="group flex flex-col rounded-lg overflow-hidden transition-colors hover:bg-black/5 dark:hover:bg-white/5"
        >
          <div class="h-48 w-full p-2 flex items-center justify-center">
            <img
              v-if="item.image"
              :src="`/i/${item.image.hash}/thumbnail`"
              class="max-w-full max-h-full object-contain rounded-sm shadow"
              loading="lazy"
              :alt="item.image.file_name"
            />
            <div
              v-else
              class="flex h-full w-full items-center justify-center text-(--md-sys-color-on-surface-variant)"
            >
              <span class="text-sm">No Image</span>
            </div>
          </div>
          <div class="flex flex-col flex-1 p-4 gap-3">
            <div class="flex items-center justify-between gap-2">
              <div
                class="text-xl font-bold text-(--md-sys-color-primary) break-all select-all"
              >
                /{{ item.route }}
              </div>
              <AnzuSplitButton
                variant="text"
                class="shrink-0"
                @click="copyAs(item.route, 'url')"
              >
                <template #icon>
                  <ClipboardIcon class="h-4 w-4" />
                </template>
                <template #menu="{ close }">
                  <AnzuButton variant="text" size="sm" class="w-full justify-start" @click="copyAs(item.route, 'url'); close()">
                    {{ t("common.actions.copyUrl") }}
                  </AnzuButton>
                  <AnzuButton variant="text" size="sm" class="w-full justify-start" @click="copyAs(item.route, 'markdown'); close()">
                    {{ t("common.actions.copyMarkdown") }}
                  </AnzuButton>
                </template>
              </AnzuSplitButton>
            </div>

            <div class="min-w-0 flex flex-col gap-1">
              <p
                class="text-sm text-(--md-sys-color-on-surface) truncate"
                :title="item.image?.file_name"
              >
                {{ item.image?.file_name || "Unknown File" }}
              </p>
              <p
                class="text-xs text-(--md-sys-color-on-surface-variant) font-mono truncate"
                :title="item.image?.hash"
              >
                {{ item.image?.hash }}
              </p>
            </div>
            <div
              class="mt-auto flex items-center justify-between pt-3 border-t border-(--md-sys-color-outline-variant)/50"
            >
              <span class="text-xs text-(--md-sys-color-on-surface-variant)">
                {{ formatDate(item.created_at) }}
              </span>

              <div class="flex items-center gap-1">
                <a :href="`/i/r/${item.route}`" target="_blank" class="block">
                  <AnzuButton
                    variant="text"
                    :title="t('common.actions.open')"
                  >
                    <template #icon>
                      <ArrowTopRightOnSquareIcon class="h-4 w-4" />
                    </template>
                  </AnzuButton>
                </a>

                <AnzuButton
                  variant="text"
                  @click="deleteRoute(item.route)"
                  :title="t('common.actions.delete')"
                >
                  <template #icon>
                    <TrashIcon class="h-4 w-4" />
                  </template>
                </AnzuButton>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="mt-8 flex justify-center">
        <AnzuPagination
          :current-page="currentPage"
          :total-pages="totalPages"
          base-url="/routes"
          :loading="loading"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch, nextTick } from "vue";
import { useAuth } from "~/composables/useAuth";
import { formatDate } from "~/utils/format";
import { useNotification } from "~/composables/useNotification";
import { useDialog, isDialogDismissedError } from "~/composables/useDialog";
import { parseApiError } from "~/utils/api-error";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuSplitButton from "~/components/AnzuSplitButton.vue";
import AnzuPagination from "~/components/AnzuPagination.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import {
  ClipboardIcon,
  ArrowTopRightOnSquareIcon,
  TrashIcon,
} from "@heroicons/vue/24/outline";

const { t } = useI18n();
useAuth();
const { notify } = useNotification();
const { confirm } = useDialog();
const route = useRoute();
const { apiUrl } = useApi();

interface Route {
  id: number;
  image_id: number;
  route: string;
  created_at: string;
  image: {
    hash: string;
    file_name: string;
  };
}

interface RouteListResponse {
  data: Route[];
  total: number;
  page: number;
  size: number;
}

const routes = ref<Route[]>([]);
const loading = ref(true);
const limit = 20;
const totalRoutes = ref(0);

const currentPage = computed(() => {
  const p = Number(route.query.page);
  return Number.isNaN(p) || p < 1 ? 1 : p;
});

const totalPages = computed(() => {
  return Math.ceil(totalRoutes.value / limit);
});

const fetchRoutes = async () => {
  loading.value = true;
  try {
    const data = await $fetch<RouteListResponse>(apiUrl("/api/v1/routes"), {
      query: {
        page: currentPage.value,
        page_size: limit,
      },
    });
    routes.value = data.data;
    totalRoutes.value = data.total;
  } catch (e: any) {
    const parsed = parseApiError(e, "Failed to load routes");
    notify({
      message: parsed.displayMessage,
      type: NotificationType.ERROR,
    });
  } finally {
    loading.value = false;
  }
};

watch(currentPage, () => {
  fetchRoutes();
});

watch(loading, (isLoading, wasLoading) => {
  if (wasLoading && !isLoading) {
    nextTick(() => {
      requestAnimationFrame(() => {
        window.scrollTo({ top: 0, behavior: "smooth" });
      });
    });
  }
});

const deleteRoute = async (routePath: string) => {
  try {
    const result = await confirm(t("common.actions.deleteConfirm"), {
      title: t("common.actions.delete"),
      variant: DialogVariant.DESTRUCTIVE,
      actions: [
        { text: t("common.actions.cancel"), variant: "text" },
        {
          text: t("common.actions.delete"),
          primary: true,
          variant: "filled",
        },
      ],
    });

    if (!result) return;

    try {
      await $fetch(apiUrl(`/api/v1/routes/${routePath}`), {
        method: "DELETE",
      });
    } catch (error: any) {
      await $fetch(apiUrl(`/api/v1/routes/${routePath}/delete`), {
        method: "POST",
      });
    }
    notify({
      message: t("common.actions.deleteSuccess"),
      type: NotificationType.SUCCESS,
    });
    await fetchRoutes();
  } catch (e: any) {
    if (isDialogDismissedError(e)) return;
    const parsed = parseApiError(e, t("common.actions.deleteFailed"));
    notify({
      message: parsed.displayMessage,
      type: NotificationType.ERROR,
    });
  }
};

const copyAs = (routePath: string, format: "url" | "markdown" = "url") => {
  const url = `${window.location.origin}/i/r/${routePath}`;
  const text = format === "markdown" ? `[/i/r/${routePath}](${url})` : url;
  navigator.clipboard.writeText(text).then(() => {
    notify({
      message: t("common.actions.copySuccess"),
      type: NotificationType.SUCCESS,
    });
  });
};

onMounted(() => {
  fetchRoutes();
});
</script>

<style scoped>
@reference "tailwindcss";
</style>
