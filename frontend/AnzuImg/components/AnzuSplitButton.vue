<template>
    <div class="inline-flex">
        <AnzuButton
            :variant="variant"
            :size="size"
            :disabled="disabled"
            :status="status"
            class="rounded-r-none"
            @click="$emit('click')"
        >
            <template v-if="$slots.icon" #icon>
                <slot name="icon" />
            </template>
            <template v-if="$slots.default" #default>
                <slot />
            </template>
        </AnzuButton>
        <AnzuDropdown
            v-model="menuOpen"
            align="right"
            width-class="w-32"
            offset-class="mt-1"
            panel-class="min-w-32"
        >
            <template #trigger="{ toggle }">
                <AnzuButton
                    :variant="variant"
                    :size="size"
                    :disabled="disabled"
                    class="w-7 shrink-0 rounded-l-none border-l border-white/20 px-0"
                    @click="toggle"
                >
                    <template #icon>
                        <ChevronDownIcon class="h-3.5 w-3.5" />
                    </template>
                </AnzuButton>
            </template>
            <template #menu="{ close }">
                <slot v-if="$slots.menu" name="menu" :close="close" />
                <div v-else class="p-1">
                    <AnzuButton
                        v-for="item in items"
                        :key="item.key"
                        variant="text"
                        size="sm"
                        class="w-full justify-start"
                        @click="$emit('select', item.key); close()"
                    >
                        {{ item.label }}
                    </AnzuButton>
                </div>
            </template>
        </AnzuDropdown>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { ChevronDownIcon } from "@heroicons/vue/24/outline";
import AnzuButton from "./AnzuButton.vue";
import AnzuDropdown from "./AnzuDropdown.vue";

export interface SplitButtonItem {
    key: string;
    label: string;
}

withDefaults(
    defineProps<{
        variant?: "filled" | "outlined" | "text" | "elevated" | "tonal";
        size?: "sm" | "md" | "lg";
        disabled?: boolean;
        status?: "default" | "loading" | "success" | "error" | "disabled";
        items?: SplitButtonItem[];
    }>(),
    {
        variant: "filled",
        size: "md",
        disabled: false,
        status: "default",
        items: () => [],
    },
);

defineEmits<{
    (e: "click"): void;
    (e: "select", key: string): void;
}>();

const menuOpen = ref(false);
</script>
