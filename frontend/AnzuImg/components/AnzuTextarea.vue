<template>
    <div class="w-full">
        <label
            v-if="label"
            class="mb-1 block text-sm font-medium text-(--md-sys-color-on-surface-variant)"
        >
            {{ label }}
        </label>
        <textarea
            :value="modelValue"
            :placeholder="placeholder"
            :disabled="disabled"
            :rows="rows"
            class="w-full rounded-lg bg-black/5 px-3 py-2 text-sm text-(--md-sys-color-on-surface) outline-none transition-[background-color,box-shadow] duration-200 ease-out hover:bg-black/8 focus:bg-black/8 placeholder:text-(--md-sys-color-on-surface-variant)/50 dark:bg-white/5 dark:hover:bg-white/10 dark:focus:bg-white/10 focus:ring-2 focus:ring-(--md-sys-color-on-surface)/15 disabled:cursor-not-allowed disabled:opacity-50"
            @input="handleInput"
            @blur="$emit('blur', $event)"
            @focus="$emit('focus', $event)"
        />
    </div>
</template>

<script setup lang="ts">
interface Props {
    modelValue?: string;
    label?: string;
    placeholder?: string;
    disabled?: boolean;
    rows?: number;
}

const props = withDefaults(defineProps<Props>(), {
    modelValue: "",
    disabled: false,
    rows: 3,
});

const emit = defineEmits<{
    (e: "update:modelValue", value: string): void;
    (e: "blur", event: FocusEvent): void;
    (e: "focus", event: FocusEvent): void;
}>();

const handleInput = (event: Event) => {
    const target = event.target as HTMLTextAreaElement;
    emit("update:modelValue", target.value);
};
</script>
