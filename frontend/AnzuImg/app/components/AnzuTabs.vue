<template>
    <div class="w-full" :class="containerClasses">
        <!-- 标签头区域 -->
        <div class="relative border-b border-(--md-sys-color-outline)" :class="headerClasses">
            <div ref="scrollerRef" class="relative overflow-x-auto scrollbar-none">
                <div class="inline-flex min-w-full" role="tablist">
                    <button v-for="(tab, index) in normalizedTabs" :key="tab._value"
                        class="relative flex items-center gap-2 px-4 py-3 text-sm font-medium transition-all duration-200 outline-none select-none focus-visible:ring-2 focus-visible:ring-(--md-sys-color-primary) focus-visible:ring-offset-2"
                        :class="getTabClasses(tab, index)" :disabled="disabled || tab.disabled" role="tab"
                        :aria-selected="isTabActive(index)" :aria-controls="`anzu-tabs-panel-${uid}-${index}`"
                        @click="selectTab(index, $event)" @keydown="handleTabKeydown($event, index)"
                        :ref="el => setTabRef(el, index)">
                        <!-- 图标插槽 -->
                        <slot v-if="$slots['tab-icon']" :name="`tab-icon-${index}`" :tab="tab" :index="index">
                            <component v-if="tab.icon" :is="tab.icon" class="w-4 h-4" />
                        </slot>

                        <!-- 标签文本 -->
                        <span class="truncate max-w-30">
                            {{ tab._label }}
                        </span>

                        <!-- 徽章 -->
                        <span v-if="tab.badge !== undefined"
                            class="inline-flex items-center justify-center min-w-5 h-5 px-1 text-xs font-medium rounded-full"
                            :class="[
                                isTabActive(index)
                                    ? 'bg-(--md-sys-color-primary) text-(--md-sys-color-on-primary)'
                                    : 'bg-(--md-sys-color-surface-variant) text-(--md-sys-color-on-surface-variant)'
                            ]">
                            {{ tab.badge }}
                        </span>
                    </button>
                </div>

                <!-- 指示器 -->
                <div v-if="showIndicator && activeTabIndex !== -1"
                    class="absolute bottom-0 h-0.5 bg-(--md-sys-color-primary) transition-all duration-200 pointer-events-none"
                    :style="indicatorStyle" />
            </div>
        </div>
        <div class="p-4">
            <div v-for="(tab, index) in normalizedTabs" :key="tab._value" :id="`anzu-tabs-panel-${uid}-${index}`"
                class="outline-none" :class="{ 'hidden': !isTabActive(index) }" role="tabpanel"
                :aria-labelledby="`anzu-tabs-tab-${uid}-${index}`" :hidden="!isTabActive(index)">
                <slot :name="`tab-content-${index}`" :tab="tab" :index="index">
                    {{ tab.content }}
                </slot>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from "vue";

interface Tab {
    label?: string;
    value?: string | number;
    disabled?: boolean;
    icon?: any;
    badge?: string | number;
    content?: any;
    [key: string]: any;
}

interface NormalizedTab extends Tab {
    _index: number;
    _value: string | number;
    _label: string;
}

interface Props {
    modelValue?: string | number;
    tabs?: Tab[];
    variant?: "primary" | "secondary" | "surface" | "filled" | "outlined";
    align?: "start" | "center" | "end" | "stretch";
    disabled?: boolean;
    showIndicator?: boolean;
    indicatorPosition?: "bottom" | "top";
    lazy?: boolean;
    keepAlive?: boolean;
    labelKey?: string;
    valueKey?: string;
}

const props = withDefaults(defineProps<Props>(), {
    modelValue: undefined,
    tabs: () => [],
    variant: "primary",
    align: "start",
    disabled: false,
    showIndicator: true,
    indicatorPosition: "bottom",
    lazy: false,
    keepAlive: false,
    labelKey: "label",
    valueKey: "value",
});

const emit = defineEmits<{
    (e: "update:modelValue", value: string | number): void;
    (e: "change", value: string | number, tab: Tab, index: number): void;
    (e: "tab-click", tab: Tab, index: number, event: MouseEvent): void;
}>();

const scrollerRef = ref<HTMLElement | null>(null);
const activeTabIndex = ref(-1);
const indicatorStyle = ref({});
const uid = Math.random().toString(36).substring(2, 9);
const tabRefs = ref<HTMLElement[]>([]);

const setTabRef = (el: any, index: number) => {
    if (el) tabRefs.value[index] = el;
};

const normalizedTabs = computed<NormalizedTab[]>(() => {
    return props.tabs.map((tab, index) => ({
        ...tab,
        _index: index,
        _value: tab.value ?? index,
        _label: getTabLabel(tab),
    }));
});

const activeTab = computed(() => {
    if (activeTabIndex.value >= 0 && activeTabIndex.value < normalizedTabs.value.length) {
        return normalizedTabs.value[activeTabIndex.value];
    }
    return null;
});

const containerClasses = computed(() => [
    props.variant === 'filled' ? 'bg-(--md-sys-color-surface)' : '',
    props.variant === 'outlined' ? 'border border-(--md-sys-color-outline) rounded-lg' : '',
    props.disabled ? 'opacity-50 pointer-events-none' : ''
]);

const headerClasses = computed(() => [
    props.align === 'center' ? 'text-center' : '',
    props.align === 'end' ? 'text-right' : '',
]);


function getTabLabel(tab: Tab): string {
    if (typeof tab === "string" || typeof tab === "number") {
        return String(tab);
    }
    return tab?.[props.labelKey] ?? tab?.["label"] ?? String(tab[props.valueKey] ?? "");
}

function isTabActive(index: number): boolean {
    return activeTabIndex.value === index;
}

function getTabClasses(tab: NormalizedTab, index: number) {
    const isActive = isTabActive(index);
    const classes = [];

    if (!isActive && !tab.disabled) {
        classes.push('text-(--md-sys-color-on-surface-variant) hover:text-(--md-sys-color-on-surface)');
    } else if (tab.disabled || props.disabled) {
        classes.push('cursor-not-allowed opacity-50');
    }

    if (props.align === 'stretch') {
        classes.push('flex-1 justify-center');
    }

    if (isActive) {
        switch (props.variant) {
            case 'filled':
                classes.push('bg-(--md-sys-color-surface-container-highest) text-(--md-sys-color-primary) rounded-t-lg');
                break;
            case 'outlined':
                classes.push('border border-(--md-sys-color-outline) border-b-transparent bg-(--md-sys-color-surface) text-(--md-sys-color-primary) rounded-t-lg');
                break;
            default:
                classes.push('text-(--md-sys-color-primary)');
        }
    }

    return classes;
}

const selectTab = (index: number, event?: MouseEvent) => {
    if (props.disabled || index < 0 || index >= normalizedTabs.value.length) {
        return;
    }

    const tab = normalizedTabs.value[index];
    if (!tab || tab.disabled) {
        return;
    }

    if (activeTabIndex.value === index) {
        if (event) {
            emit("tab-click", tab, index, event);
        }
        return;
    }

    activeTabIndex.value = index;
    const value = tab._value;

    emit("update:modelValue", value);
    emit("change", value, tab, index);

    if (event) {
        emit("tab-click", tab, index, event);
    }

    nextTick(updateIndicatorPosition);
};

const updateIndicatorPosition = () => {
    if (!props.showIndicator || activeTabIndex.value === -1) return;

    const scroller = scrollerRef.value;
    const activeEl = tabRefs.value[activeTabIndex.value];

    if (!scroller || !activeEl) return;

    const scrollerRect = scroller.getBoundingClientRect();
    const tabRect = activeEl.getBoundingClientRect();

    const left = tabRect.left - scrollerRect.left + scroller.scrollLeft;
    const width = tabRect.width;

    indicatorStyle.value = {
        left: `${left}px`,
        width: `${width}px`,
        transform: "translateX(0)",
    };
};

const handleTabKeydown = (event: KeyboardEvent, index: number) => {
    if (props.disabled) return;

    const key = event.key;
    const tabsCount = normalizedTabs.value.length;
    let targetIndex = -1;

    switch (key) {
        case "ArrowLeft":
        case "ArrowUp":
            event.preventDefault();
            targetIndex = index - 1;
            while (targetIndex >= 0 && normalizedTabs.value[targetIndex]?.disabled) {
                targetIndex--;
            }
            break;

        case "ArrowRight":
        case "ArrowDown":
            event.preventDefault();
            targetIndex = index + 1;
            while (targetIndex < tabsCount && normalizedTabs.value[targetIndex]?.disabled) {
                targetIndex++;
            }
            break;

        case "Home":
            event.preventDefault();
            targetIndex = normalizedTabs.value.findIndex(t => !t.disabled);
            break;

        case "End":
            event.preventDefault();
            // Find last enabled tab
            for (let i = tabsCount - 1; i >= 0; i--) {
                if (!normalizedTabs.value[i]?.disabled) {
                    targetIndex = i;
                    break;
                }
            }
            break;

        case "Enter":
        case " ":
            event.preventDefault();
            selectTab(index);
            return;
    }

    if (targetIndex !== -1 && targetIndex >= 0 && targetIndex < tabsCount) {
        selectTab(targetIndex);
        nextTick(() => {
            tabRefs.value[targetIndex]?.focus();
        });
    }
};

// Watchers
watch(
    () => props.modelValue,
    (newValue) => {
        if (newValue === undefined) return;
        const index = normalizedTabs.value.findIndex(tab => tab._value === newValue);
        if (index !== -1 && index !== activeTabIndex.value) {
            activeTabIndex.value = index;
            nextTick(updateIndicatorPosition);
        }
    },
    { immediate: true },
);

watch(
    normalizedTabs,
    (tabs) => {
        if (activeTabIndex.value === -1 && tabs.length > 0) {
            const firstEnabled = tabs.findIndex(t => !t.disabled);
            if (firstEnabled !== -1) {
                selectTab(firstEnabled);
            }
        }
        nextTick(updateIndicatorPosition);
    }
);

onMounted(() => {
    nextTick(updateIndicatorPosition);
    window.addEventListener("resize", updateIndicatorPosition);
});

onUnmounted(() => {
    window.removeEventListener("resize", updateIndicatorPosition);
});

defineExpose({
    selectTab,
    activeTabIndex,
    activeTab,
});
</script>

<style scoped>
@reference "tailwindcss";
</style>
