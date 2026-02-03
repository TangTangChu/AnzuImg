<template>
    <h1 class="mb-6 text-3xl font-bold text-center">{{ t("routes.title") }}</h1>

    <div v-if="loading" class="flex justify-center p-8">
        <AnzuProgressRing status="loading" />
    </div>

    <div v-else-if="routes.length === 0"
        class="flex flex-col items-center justify-center p-8 text-(--md-sys-color-on-surface-variant)">
        <p>{{ t("routes.noRoutes", "No routes found") }}</p>
    </div>

    <div v-else>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            <div v-for="route in routes" :key="route.id"
                class="group flex flex-col border border-(--md-sys-color-outline-variant) rounded-md overflow-hidden transition-colors">
                <div class="h-48 w-full p-2 flex items-center justify-center">
                    <img v-if="route.image" :src="`/i/${route.image.hash}/thumbnail`"
                        class="max-w-full max-h-full object-contain rounded-sm shadow" loading="lazy"
                        :alt="route.image.file_name" />
                    <div v-else
                        class="flex h-full w-full items-center justify-center text-(--md-sys-color-on-surface-variant)">
                        <span class="text-sm">No Image</span>
                    </div>
                </div>
                <div class="flex flex-col flex-1 p-4 gap-3">
                    <div class="flex items-center justify-between gap-2">
                        <div class="text-xl font-bold text-(--md-sys-color-primary) break-all select-all">
                            /{{ route.route }}
                        </div>
                        <AnzuButton variant="text" class="w-8! h-8! p-0! min-w-0! shrink-0"
                            @click="copyLink(route.route)" :title="t('common.actions.copyLink')">
                            <LinkIcon class="h-5 w-5" />
                        </AnzuButton>
                    </div>

                    <div class="min-w-0 flex flex-col gap-1">
                        <p class="text-sm text-(--md-sys-color-on-surface) truncate" :title="route.image?.file_name">
                            {{ route.image?.file_name || "Unknown File" }}
                        </p>
                        <p class="text-xs text-(--md-sys-color-on-surface-variant) font-mono truncate"
                            :title="route.image?.hash">
                            {{ route.image?.hash }}
                        </p>
                    </div>
                    <div
                        class="mt-auto flex items-center justify-between pt-3 border-t border-(--md-sys-color-outline-variant)/50">
                        <span class="text-xs text-(--md-sys-color-on-surface-variant)">
                            {{ formatDate(route.created_at) }}
                        </span>

                        <div class="flex items-center gap-1">
                            <a :href="`/i/r/${route.route}`" target="_blank" class="block">
                                <AnzuButton variant="text" class="w-8! h-8! p-0! min-w-0!"
                                    :title="t('common.actions.open')">
                                    <ArrowTopRightOnSquareIcon class="h-4 w-4" />
                                </AnzuButton>
                            </a>

                            <AnzuButton variant="text" class="w-8! h-8! p-0! min-w-0!" @click="deleteRoute(route.route)"
                                :title="t('common.actions.delete')">
                                <TrashIcon class="h-4 w-4" />
                            </AnzuButton>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div class="mt-8 flex justify-center">
            <AnzuPagination :current-page="currentPage" :total-pages="totalPages" base-url="/routes" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import { useAuth } from "~/composables/useAuth";
import { formatDate } from "~/utils/format";
import { useNotification } from "~/composables/useNotification";
import { useDialog } from "~/composables/useDialog";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuPagination from "~/components/AnzuPagination.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import {
    LinkIcon,
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
    } catch (e) {
        notify({
            message: "Failed to load routes",
            type: NotificationType.ERROR,
        });
    } finally {
        loading.value = false;
    }
};

watch(currentPage, () => {
    fetchRoutes();
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
        if (e.message === "Dialog closed" || e.message === "All dialogs closed")
            return;
        notify({
            message: t("common.actions.deleteFailed"),
            type: NotificationType.ERROR,
        });
    }
};

const copyLink = (routePath: string) => {
    const url = `${window.location.origin}/i/r/${routePath}`;
    navigator.clipboard.writeText(url).then(() => {
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
