<template>
    <Teleport to="body">
        <Transition name="dialog-overlay">
            <div v-if="visible && showOverlay" class="fixed inset-0 z-50 bg-black/50"
                :class="{ 'cursor-pointer': overlayClickable }" @click="handleOverlayClick" />
        </Transition>

        <Transition :name="positionTransition">
            <div v-if="visible" ref="dialogRef" class="fixed z-50 flex flex-col overflow-hidden" :class="[
                dialogClasses,
                positionClasses,
                sizeClasses,
                variantClasses,
                customClass,
            ]" role="dialog" aria-modal="true" :aria-labelledby="title ? 'dialog-title' : undefined"
                :aria-describedby="message ? 'dialog-message' : undefined" @keydown.esc="handleEsc">
                <!-- Header -->
                <div v-if="showHeader"
                    class="flex items-center justify-between border-b border-(--md-sys-color-outline-variant) bg-(--md-sys-color-surface-container) px-6 py-4">
                    <div class="flex items-center gap-3">
                        <component v-if="iconComponent" :is="iconComponent" class="h-5 w-5 shrink-0"
                            :class="iconColorClass" />
                        <h2 v-if="title" id="dialog-title"
                            class="text-lg font-semibold text-(--md-sys-color-on-surface)">
                            {{ title }}
                        </h2>
                    </div>

                    <button v-if="showCloseButton" @click="handleClose"
                        class="flex h-8 w-8 items-center justify-center rounded-full text-base leading-none text-(--md-sys-color-on-surface-variant) transition-colors hover:bg-(--md-sys-color-surface-container-high) hover:text-(--md-sys-color-on-surface)"
                        aria-label="关闭对话框">
                        &times;
                    </button>
                </div>

                <!-- Content -->
                <div class="flex-1 overflow-auto bg-(--md-sys-color-surface-container) p-6" :class="{
                    'pt-4': !showHeader,
                    'max-h-[calc(100vh-16rem)]': maxHeight === undefined,
                }" :style="contentStyles">
                    <div v-if="type === DialogType.CUSTOM && component">
                        <component :is="component" v-bind="componentProps" @close="handleClose" />
                    </div>
                    <template v-else>
                        <p v-if="message" id="dialog-message" class="text-(--md-sys-color-on-surface-variant)">
                            {{ message }}
                        </p>
                        <slot v-else />
                    </template>
                </div>

                <!-- Actions -->
                <div v-if="showActions"
                    class="flex items-center justify-end gap-2 bg-(--md-sys-color-surface-container) px-6 py-4">
                    <AnzuButton v-for="(action, index) in effectiveActions" :key="index" @click="handleAction(action)"
                        :variant="action.variant || 'text'" :disabled="action.disabled"
                        :status="action.loading ? 'loading' : 'default'" :class="{
                            'order-last': action.primary,
                        }">
                        <template v-if="action.icon">
                            <component :is="action.icon" class="mr-2 h-4 w-4" />
                        </template>
                        {{ action.text }}
                    </AnzuButton>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<script setup lang="ts">
import {
    ref,
    computed,
    onMounted,
    onUnmounted,
    watch,
    nextTick,
    type Ref,
} from "vue";
import {
    DialogType,
    DialogSize,
    DialogPosition,
    DialogVariant,
    type DialogAction,
    type DialogOptions,
} from "@/types/dialog";
import AnzuButton from "./AnzuButton.vue";
import {
    ExclamationTriangleIcon,
    CheckCircleIcon,
    XCircleIcon,
    InformationCircleIcon,
    QuestionMarkCircleIcon,
} from "@heroicons/vue/24/outline";
import { useClickAway } from "@/composables/useClickAway";

const props = withDefaults(
    defineProps<
        DialogOptions & {
            visible: boolean;
        }
    >(),
    {
        title: undefined,
        message: undefined,
        type: DialogType.ALERT,
        size: DialogSize.MD,
        position: DialogPosition.CENTER,
        variant: DialogVariant.DEFAULT,
        actions: () => [],
        showCloseButton: true,
        closeOnClickOutside: true,
        closeOnEsc: true,
        showOverlay: true,
        overlayClickable: true,
        persistent: false,
        maxWidth: undefined,
        minWidth: undefined,
        maxHeight: undefined,
        minHeight: undefined,
        customClass: "",
        icon: undefined,
        component: undefined,
        componentProps: () => ({}),
        visible: false,
    }
);

const emit = defineEmits<{
    (e: "update:visible", value: boolean): void;
    (e: "close"): void;
    (e: "action", action: DialogAction): void;
    (e: "confirm", value?: any): void;
    (e: "cancel"): void;
}>();

const dialogRef = ref<HTMLElement | null>(null);

const { t } = useI18n();

const showHeader = computed(() => props.title || props.icon || props.showCloseButton);
const showActions = computed(() => props.actions && props.actions.length > 0);

const iconComponent = computed(() => {
    if (props.icon) return props.icon;

    const iconMap: Record<string, any> = {
        [DialogVariant.DESTRUCTIVE]: ExclamationTriangleIcon,
        [DialogVariant.SUCCESS]: CheckCircleIcon,
        [DialogVariant.WARNING]: ExclamationTriangleIcon,
        [DialogVariant.INFO]: InformationCircleIcon,
        [DialogType.CONFIRM]: QuestionMarkCircleIcon,
        [DialogType.PROMPT]: QuestionMarkCircleIcon,
    };

    return iconMap[props.variant] || iconMap[props.type] || null;
});

const iconColorClass = computed(() => {
    const colorMap: Record<string, string> = {
        [DialogVariant.DESTRUCTIVE]: "text-(--md-sys-color-error)",
        [DialogVariant.SUCCESS]: "text-(--md-sys-color-primary)",
        [DialogVariant.WARNING]: "text-(--md-sys-color-tertiary)",
        [DialogVariant.INFO]: "text-(--md-sys-color-secondary)",
        [DialogType.CONFIRM]: "text-(--md-sys-color-primary)",
        [DialogType.PROMPT]: "text-(--md-sys-color-primary)",
    };

    return colorMap[props.variant] || colorMap[props.type] || "text-(--md-sys-color-primary)";
});

const effectiveActions = computed(() => {
    if (props.actions && props.actions.length > 0) {
        return props.actions;
    }

    const defaultActions: Record<string, DialogAction[]> = {
        [DialogType.ALERT]: [
            {
                text: t("common.labels.ok"),
                primary: true,
                variant: "filled",
            },
        ],
        [DialogType.CONFIRM]: [
            {
                text: t("common.actions.cancel"),
                variant: "text",
            },
            {
                text: t("common.actions.confirm"),
                primary: true,
                variant: "filled",
            },
        ],
        [DialogType.PROMPT]: [
            {
                text: t("common.actions.cancel"),
                variant: "text",
            },
            {
                text: t("common.labels.ok"),
                primary: true,
                variant: "filled",
            },
        ],
    };

    return defaultActions[props.type] || [];
});

const positionClasses = computed(() => {
    const classes: Record<string, string> = {
        [DialogPosition.CENTER]: "top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2",
        [DialogPosition.TOP]: "top-4 left-1/2 -translate-x-1/2",
        [DialogPosition.BOTTOM]: "bottom-4 left-1/2 -translate-x-1/2",
        [DialogPosition.LEFT]: "top-1/2 left-4 -translate-y-1/2",
        [DialogPosition.RIGHT]: "top-1/2 right-4 -translate-y-1/2",
        [DialogPosition.TOP_LEFT]: "top-4 left-4",
        [DialogPosition.TOP_RIGHT]: "top-4 right-4",
        [DialogPosition.BOTTOM_LEFT]: "bottom-4 left-4",
        [DialogPosition.BOTTOM_RIGHT]: "bottom-4 right-4",
    };

    return classes[props.position] || classes[DialogPosition.CENTER];
});

const positionTransition = computed(() => {
    const transitions: Record<string, string> = {
        [DialogPosition.CENTER]: "dialog-center",
        [DialogPosition.TOP]: "dialog-top",
        [DialogPosition.BOTTOM]: "dialog-bottom",
        [DialogPosition.LEFT]: "dialog-left",
        [DialogPosition.RIGHT]: "dialog-right",
        [DialogPosition.TOP_LEFT]: "dialog-top-left",
        [DialogPosition.TOP_RIGHT]: "dialog-top-right",
        [DialogPosition.BOTTOM_LEFT]: "dialog-bottom-left",
        [DialogPosition.BOTTOM_RIGHT]: "dialog-bottom-right",
    };

    return transitions[props.position] || "dialog-center";
});

const sizeClasses = computed(() => {
    const classes: Record<string, string> = {
        [DialogSize.SM]: "w-80 rounded-xl",
        [DialogSize.MD]: "w-96 rounded-xl",
        [DialogSize.LG]: "w-xl rounded-xl",
        [DialogSize.XL]: "w-3xl rounded-xl",
        [DialogSize.FULL]: "inset-4 rounded-xl",
    };

    return classes[props.size] || classes[DialogSize.MD];
});

const variantClasses = computed(() => {
    const classes: Record<string, string> = {
        [DialogVariant.DESTRUCTIVE]: "border-(--md-sys-color-error-container)",
        [DialogVariant.SUCCESS]: "border-(--md-sys-color-primary-container)",
        [DialogVariant.WARNING]: "border-(--md-sys-color-tertiary-container)",
        [DialogVariant.INFO]: "border-(--md-sys-color-secondary-container)",
    };

    return classes[props.variant] || "";
});

const dialogClasses = computed(() => {
    return "bg-(--md-sys-color-surface-container) border border-(--md-sys-color-outline-variant) shadow-lg";
});

const contentStyles = computed(() => {
    const styles: Record<string, string> = {};

    if (props.maxWidth) styles.maxWidth = props.maxWidth;
    if (props.minWidth) styles.minWidth = props.minWidth;
    if (props.maxHeight) styles.maxHeight = props.maxHeight;
    if (props.minHeight) styles.minHeight = props.minHeight;

    return styles;
});

const handleClose = () => {
    if (props.persistent) return;

    emit("update:visible", false);
    emit("close");
    emit("cancel");
};

const handleOverlayClick = () => {
    if (props.overlayClickable && props.closeOnClickOutside) {
        handleClose();
    }
};

const handleEsc = (event: KeyboardEvent) => {
    if (props.closeOnEsc && !props.persistent) {
        event.preventDefault();
        handleClose();
    }
};

const handleAction = async (action: DialogAction) => {
    if (action.disabled || action.loading) return;

    emit("action", action);

    if (action.handler) {
        try {
            await action.handler();
        } catch (error) {
            console.error("Dialog action handler error:", error);
        }
    }

    if (props.type === DialogType.CONFIRM && action.primary) {
        emit("confirm");
    }

};

// 点击外部关闭
useClickAway(dialogRef as Ref<HTMLElement | null>, () => {
    if (props.closeOnClickOutside && !props.persistent) {
        handleClose();
    }
});

onMounted(() => {
    if (props.visible) {
        nextTick(() => {
            dialogRef.value?.focus();
        });
    }
});

watch(
    () => props.visible,
    (newValue) => {
        if (newValue) {
            nextTick(() => {
                dialogRef.value?.focus();
            });
        }
    }
);

const handleKeydown = (event: KeyboardEvent) => {
    if (!props.visible) return;

    if (event.key === "Escape" && props.closeOnEsc && !props.persistent) {
        event.preventDefault();
        handleClose();
    }

    if (event.key === "Enter" && !event.shiftKey) {
        const primaryAction = effectiveActions.value.find((action) => action.primary);
        if (primaryAction && !primaryAction.disabled && !primaryAction.loading) {
            event.preventDefault();
            handleAction(primaryAction);
        }
    }
};

onMounted(() => {
    window.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
    window.removeEventListener("keydown", handleKeydown);
});
</script>

<style scoped>
@reference "tailwindcss";

.dialog-overlay-enter-active,
.dialog-overlay-leave-active {
    transition: opacity 0.2s ease;
}

.dialog-overlay-enter-from,
.dialog-overlay-leave-to {
    opacity: 0;
}

.dialog-center-enter-active,
.dialog-center-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-center-enter-from,
.dialog-center-leave-to {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.95);
}

.dialog-top-enter-active,
.dialog-top-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-top-enter-from,
.dialog-top-leave-to {
    opacity: 0;
    transform: translate(-50%, -20px);
}

.dialog-bottom-enter-active,
.dialog-bottom-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-bottom-enter-from,
.dialog-bottom-leave-to {
    opacity: 0;
    transform: translate(-50%, 20px);
}

.dialog-left-enter-active,
.dialog-left-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-left-enter-from,
.dialog-left-leave-to {
    opacity: 0;
    transform: translate(-20px, -50%);
}

.dialog-right-enter-active,
.dialog-right-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-right-enter-from,
.dialog-right-leave-to {
    opacity: 0;
    transform: translate(20px, -50%);
}

.dialog-top-left-enter-active,
.dialog-top-left-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-top-left-enter-from,
.dialog-top-left-leave-to {
    opacity: 0;
    transform: translateY(-20px);
}

.dialog-top-right-enter-active,
.dialog-top-right-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-top-right-enter-from,
.dialog-top-right-leave-to {
    opacity: 0;
    transform: translateY(-20px);
}

.dialog-bottom-left-enter-active,
.dialog-bottom-left-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-bottom-left-enter-from,
.dialog-bottom-left-leave-to {
    opacity: 0;
    transform: translateY(20px);
}

.dialog-bottom-right-enter-active,
.dialog-bottom-right-leave-active {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.dialog-bottom-right-enter-from,
.dialog-bottom-right-leave-to {
    opacity: 0;
    transform: translateY(20px);
}
</style>
