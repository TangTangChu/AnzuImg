<script lang="ts" setup>
import { ref } from "vue";
import { MoonIcon, SunIcon } from "@heroicons/vue/24/outline";
import { useColorPalette } from "@/composables/useColorPalette";
import AnzuDropdown from "@/components/AnzuDropdown.vue";

const { isDark } = useTheme();
const { primaryColor, setPrimaryColor } = useColorPalette();
const { t } = useI18n();

const showMenu = ref(false);

const Themes = [
    { name: "矢泽妮可", color: "#ff4f91" },
    { name: "园田海未", color: "#1769ff" },
    { name: "高坂穗乃果", color: "#f38500" },
    { name: "南小鸟", color: "#cebfbf" },
    { name: "星空凛", color: "#fff832" },
    { name: "小泉花阳", color: "#6ae673" },
    { name: "西木野真姬", color: "#ff503e" },
    { name: "东条希", color: "#c455f6" },
    { name: "绚濑绘里", color: "#7aeeff" },
    { name: "初音ミク", color: "#39C5BB" },
];

const updateColor = (event: Event) => {
    const target = event.target as HTMLInputElement;
    setPrimaryColor(target.value);
};

const selectPreset = (color: string) => {
    setPrimaryColor(color);
};

const setTheme = (nextIsDark: boolean) => {
    isDark.value = nextIsDark;

    if (import.meta.client) {
        localStorage.setItem("dark-theme", nextIsDark ? "dark" : "light");
        document.documentElement.classList.toggle("dark", nextIsDark);
    }
};
</script>

<template>
    <AnzuDropdown v-model="showMenu" width-class="w-72" align="right">
        <template #trigger>
            <button
                class="flex h-7.5 w-7.5 items-center justify-center rounded-full text-(--md-sys-color-on-surface) transition-colors duration-200 hover:bg-(--md-sys-color-surface-container-high)"
                @click="showMenu = !showMenu" :title="t('common.themeMenu.tooltip')" aria-haspopup="true"
                :aria-expanded="showMenu ? 'true' : 'false'" type="button">
                <MoonIcon v-if="isDark" class="box-border p-1" :style="{ color: primaryColor }" />
                <SunIcon v-else class="box-border p-1" :style="{ color: primaryColor }" />
            </button>
        </template>

        <template #menu>
            <div class="p-4" role="dialog" :aria-label="t('common.themeMenu.title')">
                <div class="mb-4">
                    <label class="mb-2 block text-xs font-medium text-(--md-sys-color-on-surface-variant)">
                        {{ t("common.themeMenu.mode") }}
                    </label>

                    <div class="grid grid-cols-2 gap-2">
                        <button type="button"
                            class="flex items-center justify-center gap-2 rounded-lg px-2 py-2 text-sm transition-colors"
                            :class="[
                                !isDark
                                    ? 'bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)'
                                    : 'text-(--md-sys-color-on-surface) hover:bg-(--md-sys-color-surface-container-high)',
                            ]" @click="setTheme(false)">
                            <SunIcon class="h-4 w-4" />
                            {{ t("common.themeMenu.light") }}
                        </button>

                        <button type="button"
                            class="flex items-center justify-center gap-2 rounded-lg px-2 py-2 text-sm transition-colors"
                            :class="[
                                isDark
                                    ? 'bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)'
                                    : 'text-(--md-sys-color-on-surface) hover:bg-(--md-sys-color-surface-container-high)',
                            ]" @click="setTheme(true)">
                            <MoonIcon class="h-4 w-4" />
                            {{ t("common.themeMenu.dark") }}
                        </button>
                    </div>
                </div>
                <div>
                    <label class="mb-2 block text-xs font-medium text-(--md-sys-color-on-surface-variant)">
                        {{ t("common.themeColorPicker.title") }}
                    </label>

                    <div class="mb-3">
                        <label class="mb-2 block text-xs font-medium text-(--md-sys-color-on-surface-variant)">
                            {{ t("common.themeColorPicker.customColor") }}
                        </label>
                        <div class="flex items-center gap-2">
                            <div class="h-8 w-8 rounded-full border border-(--md-sys-color-outline-variant)"
                                :style="{ background: primaryColor }"></div>
                            <input type="color" :value="primaryColor" @input="updateColor"
                                class="absolute h-8 w-8 cursor-pointer opacity-0" :aria-label="t('common.themeColorPicker.colorInputLabel')
                                    " />
                            <span class="font-mono text-sm text-(--md-sys-color-on-surface)">{{ primaryColor }}</span>
                        </div>
                    </div>

                    <div>
                        <label class="mb-2 block text-xs font-medium text-(--md-sys-color-on-surface-variant)">
                            {{ t("common.themeColorPicker.Themes") }}
                        </label>
                        <div class="custom-scrollbar flex max-h-56 flex-col gap-1 overflow-y-auto pr-1">
                            <button v-for="theme in Themes" :key="theme.name" @click="selectPreset(theme.color)"
                                class="group flex w-full items-center rounded-lg px-2 py-1.5 transition-colors hover:bg-(--md-sys-color-secondary-container)"
                                :class="{
                                    'bg-(--md-sys-color-secondary-container)':
                                        primaryColor === theme.color,
                                }" type="button" :aria-label="t('common.themeColorPicker.themeItem', {
                                    name: theme.name,
                                })
                                    ">
                                <div class="mr-3 h-6 w-6 shrink-0 rounded-full border border-(--md-sys-color-outline-variant) shadow-sm transition-transform group-hover:scale-110"
                                    :style="{ backgroundColor: theme.color }"></div>
                                <span class="truncate text-sm text-(--md-sys-color-on-surface)">{{ theme.name }}</span>
                                <div v-if="primaryColor === theme.color"
                                    class="ml-auto h-2 w-2 rounded-full bg-(--md-sys-color-primary)"></div>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </template>
    </AnzuDropdown>
</template>
