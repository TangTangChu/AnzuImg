<template>
    <div class="w-full">
        <label v-if="label" class="mb-1 block text-sm font-medium text-(--md-sys-color-on-surface-variant)">
            {{ label }}
        </label>
        
        <!-- 标签显示区域 -->
        <div 
            class="flex flex-wrap items-center gap-2 rounded-lg border bg-transparent px-3 py-2 transition-all duration-200"
            :class="[
                errorMessage
                    ? 'border-(--md-sys-color-error) ring-1 ring-(--md-sys-color-error)'
                    : isFocused
                        ? 'border-(--md-sys-color-primary) ring-1 ring-(--md-sys-color-primary)'
                        : 'border-(--md-sys-color-outline) hover:border-(--md-sys-color-outline-variant)',
                disabled ? 'cursor-not-allowed opacity-50' : 'cursor-text'
            ]"
            @focusin="isFocused = true"
            @focusout="isFocused = false"
            @click="focusInput"
        >
            <div 
                v-for="(tag, index) in modelValue" 
                :key="getTagKey(tag, index)" 
                class="inline-flex items-center gap-1 rounded-full bg-(--md-sys-color-secondary-container) px-3 py-1 text-sm text-(--md-sys-color-on-secondary-container) transition-colors"
                :class="[
                    removable ? 'pr-2' : '',
                    disabled ? 'opacity-50 cursor-not-allowed' : ''
                ]"
            >
                <span class="truncate max-w-50">
                    {{ getTagLabel(tag) }}
                </span>
                <button 
                    v-if="removable && !disabled" 
                    type="button" 
                    class="flex h-4 w-4 items-center justify-center rounded-full text-(--md-sys-color-on-secondary-container)/60 transition-colors hover:bg-(--md-sys-color-on-secondary-container)/10 hover:text-(--md-sys-color-on-secondary-container)"
                    @click.stop="removeTag(index)" 
                    :aria-label="getRemoveAriaLabel(tag)"
                >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                        stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <!-- 输入框 -->
            <div 
                v-if="!disabled && (maxTags === undefined || modelValue.length < maxTags)"
                class="relative flex-1 min-w-30"
            >
                <input 
                    ref="inputRef" 
                    v-model="inputValue" 
                    type="text" 
                    class="w-full bg-transparent p-0 text-sm text-(--md-sys-color-on-surface) outline-none placeholder:text-(--md-sys-color-on-surface-variant)/50"
                    :placeholder="placeholderText" 
                    :disabled="disabled" 
                    @keydown="handleKeydown" 
                    @blur="handleBlur"
                    @focus="$emit('focus', $event)"
                />
            </div>
        </div>

        <!-- 辅助信息 -->
        <div v-if="errorMessage" class="mt-1 text-xs text-(--md-sys-color-error)">
            {{ errorMessage }}
        </div>
        <div v-else-if="hint" class="mt-1 text-xs text-(--md-sys-color-on-surface-variant)">
            {{ hint }}
        </div>

        <!-- 计数器 -->
        <div 
            v-if="maxTags !== undefined" 
            class="mt-2 text-xs"
            :class="modelValue.length >= maxTags ? 'text-(--md-sys-color-error)' : 'text-(--md-sys-color-on-surface-variant)'"
        >
            {{ t('tags.counter', { current: modelValue.length, max: maxTags }) }}
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";

type TagType = string | number | Record<string, any>;

interface Props {
    modelValue: TagType[];
    label?: string;
    labelKey?: string;
    valueKey?: string;
    removable?: boolean;
    maxTags?: number;
    placeholder?: string;
    disabled?: boolean;
    error?: boolean | string;
    hint?: string;
    addOnBlur?: boolean;
    addOnKeys?: string[];
    removeLabel?: string;
    allowDuplicates?: boolean;
    validate?: (tag: string) => string | boolean;
    useI18n?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
    modelValue: () => [],
    label: "",
    labelKey: "label",
    valueKey: "value",
    removable: true,
    maxTags: undefined,
    placeholder: "",
    disabled: false,
    error: false,
    hint: "",
    addOnBlur: true,
    addOnKeys: () => [",", "Enter", "Tab"],
    removeLabel: "",
    allowDuplicates: false,
    validate: () => true,
    useI18n: true,
});

const emit = defineEmits<{
    (e: "update:modelValue", value: TagType[]): void;
    (e: "change", value: TagType[], added?: TagType | undefined, removed?: TagType | undefined): void;
    (e: "add", tag: TagType): void;
    (e: "remove", tag: TagType, index: number): void;
    (e: "focus", event: FocusEvent): void;
    (e: "blur", event: FocusEvent): void;
    (e: "error", message: string): void;
}>();

const { t } = useI18n();

const inputRef = ref<HTMLInputElement | null>(null);
const inputValue = ref("");
const isFocused = ref(false);
const errorMessage = computed(() =>
    typeof props.error === "string" ? props.error : ""
);

const placeholderText = computed(() =>
    props.placeholder || t("tags.addPlaceholder")
);

const removeLabelText = computed(() =>
    props.removeLabel || t("tags.remove")
);

const getRemoveAriaLabel = (tag: TagType) => {
    if (props.removeLabel) return props.removeLabel;
    return t("tags.removeLabel", { tag: getTagLabel(tag) });
};

const getTagLabel = (tag: TagType): string => {
    if (typeof tag === "string" || typeof tag === "number") {
        return String(tag);
    }
    return tag?.[props.labelKey] ?? String(tag[props.valueKey] ?? "");
};

const getTagKey = (tag: TagType, index: number): string | number => {
    if (typeof tag === "object" && tag !== null) {
        return tag[props.valueKey] ?? index;
    }
    return tag;
};

const addTag = (tagValue: string) => {
    const trimmed = tagValue.trim();
    if (!trimmed) return;

    const validationResult = props.validate(trimmed);
    if (typeof validationResult === "string") {
        emit("error", validationResult);
        return;
    }
    if (validationResult === false) {
        emit("error", t("tags.validationError"));
        return;
    }

    // 检查重复
    if (!props.allowDuplicates) {
        const exists = props.modelValue.some(tag =>
            getTagLabel(tag).toLowerCase() === trimmed.toLowerCase()
        );
        if (exists) {
            emit("error", t("tags.duplicateError"));
            return;
        }
    }

    if (props.maxTags !== undefined && props.modelValue.length >= props.maxTags) {
        emit("error", t("tags.maxTagsError", { max: props.maxTags }));
        return;
    }

    const newTag = trimmed;
    const newTags = [...props.modelValue, newTag];

    emit("update:modelValue", newTags);
    emit("change", newTags, newTag, undefined);
    emit("add", newTag);

    inputValue.value = "";
};

const removeTag = (index: number) => {
    if (props.disabled || index < 0 || index >= props.modelValue.length) return;

    const removedTag = props.modelValue[index];
    if (removedTag === undefined) return;

    const newTags = props.modelValue.filter((_, i) => i !== index);

    emit("update:modelValue", newTags);
    emit("change", newTags, undefined, removedTag);
    emit("remove", removedTag, index);
};

const handleKeydown = (event: KeyboardEvent) => {
    if (props.disabled) return;

    const key = event.key;

    // 添加标签的按键
    if (props.addOnKeys.includes(key)) {
        event.preventDefault();
        if (inputValue.value.trim()) {
            addTag(inputValue.value);
        }
        return;
    }

    if (key === "Backspace" && !inputValue.value && props.modelValue.length > 0) {
        event.preventDefault();
        removeTag(props.modelValue.length - 1);
        return;
    }

    if (props.maxTags !== undefined && props.modelValue.length >= props.maxTags) {
        event.preventDefault();
        return;
    }
};

const handleBlur = (event: FocusEvent) => {
    if (props.addOnBlur && inputValue.value.trim()) {
        addTag(inputValue.value);
    }
    emit("blur", event);
};

const focusInput = () => {
    if (inputRef.value && !props.disabled) {
        inputRef.value.focus();
    }
};

defineExpose({
    focusInput,
    addTag: (tag: string) => addTag(tag),
    removeTag: (index: number) => removeTag(index),
    clear: () => {
        emit("update:modelValue", []);
        emit("change", [], undefined as TagType | undefined, undefined as TagType | undefined);
    },
});

watch(errorMessage, (newError) => {
    if (newError) {
        console.warn("AnzuTags error:", newError);
    }
});
</script>

<style scoped>
@reference "tailwindcss";
</style>
