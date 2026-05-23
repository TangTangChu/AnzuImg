<template>
    <div
        v-if="totalPages > 1"
        class="mt-3 flex items-center justify-end gap-1 text-xs text-(--md-sys-color-on-surface-variant)"
    >
        <AnzuButton
            variant="text"
            size="md"
            class="w-9! min-w-9! px-0!"
            :disabled="page <= 1 || loading"
            :aria-label="t('common.actions.paginationPrevious')"
            @click="$emit('update', page - 1)"
        >
            <template #icon>
                <ChevronLeftIcon class="h-4 w-4" />
            </template>
        </AnzuButton>
        <span class="px-2 text-(--md-sys-color-on-surface)">{{ page }} / {{ totalPages }}</span>
        <AnzuButton
            variant="text"
            size="md"
            class="w-9! min-w-9! px-0!"
            :disabled="page >= totalPages || loading"
            :aria-label="t('common.actions.paginationNext')"
            @click="$emit('update', page + 1)"
        >
            <template #icon>
                <ChevronRightIcon class="h-4 w-4" />
            </template>
        </AnzuButton>
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { ChevronLeftIcon, ChevronRightIcon } from "@heroicons/vue/24/outline";
import AnzuButton from "~/components/AnzuButton.vue";

const props = defineProps<{
    page: number;
    total: number;
    size: number;
    loading?: boolean;
}>();

defineEmits<{
    (e: "update", page: number): void;
}>();

const { t } = useI18n();

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.size)));
</script>
