<template>
    <div class="space-y-8">
        <div
            v-if="!allowWebModify"
            class="rounded-lg border border-(--md-sys-color-outline-variant) bg-(--md-sys-color-surface-variant) p-3 text-sm text-(--md-sys-color-on-surface-variant)"
        >
            {{ t("settings.systemConfig.disabledNotice") }}
        </div>

        <div v-for="group in groupedSchema" :key="group.name">
            <h3 class="mb-3 text-lg font-semibold">
                {{ t(`settings.systemConfig.groups.${group.name}`) }}
            </h3>
            <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div
                    v-for="field in group.fields"
                    :key="field.key"
                    class="rounded-xl border border-(--md-sys-color-outline-variant) p-4"
                >
                    <div class="mb-2 flex items-center justify-between gap-2">
                        <label class="text-sm font-medium">
                            {{ t(`settings.systemConfig.fields.${field.key}.label`) }}
                        </label>
                        <span
                            v-if="isOverridden(field.key)"
                            class="text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)"
                        >
                            {{ t("settings.systemConfig.overridden") }}
                        </span>
                    </div>
                    <p
                        v-if="te(`settings.systemConfig.fields.${field.key}.hint`)"
                        class="mb-2 text-xs text-(--md-sys-color-on-surface-variant)"
                    >
                        {{ t(`settings.systemConfig.fields.${field.key}.hint`) }}
                    </p>

                    <component
                        v-if="resolveControl(field) === 'enum'"
                        :is="AnzuComboBox"
                        :model-value="String(local[field.key] ?? '')"
                        :items="(field.options ?? []).map((o) => ({ label: o, value: o }))"
                        :disabled="!allowWebModify"
                        @update:modelValue="(v: any) => onInput(field.key, v)"
                    />
                    <AnzuCheckbox
                        v-else-if="resolveControl(field) === 'bool'"
                        :model-value="!!local[field.key]"
                        :disabled="!allowWebModify"
                        @update:modelValue="(v: any) => onInput(field.key, !!v)"
                    />
                    <AnzuTags
                        v-else-if="resolveControl(field) === 'list'"
                        :model-value="(local[field.key] as string[]) ?? []"
                        :disabled="!allowWebModify"
                        @update:modelValue="(v: any) => onInput(field.key, v)"
                    />
                    <textarea
                        v-else-if="resolveControl(field) === 'multiline'"
                        :value="String(local[field.key] ?? '')"
                        :disabled="!allowWebModify"
                        rows="3"
                        class="w-full rounded-md border border-(--md-sys-color-outline-variant) bg-transparent p-2 text-sm focus:outline-none focus:border-(--md-sys-color-primary)"
                        @input="(e: any) => onInput(field.key, e.target.value)"
                    />
                    <AnzuInput
                        v-else-if="resolveControl(field) === 'int'"
                        :model-value="String(local[field.key] ?? '')"
                        type="number"
                        :min="field.min"
                        :max="field.max"
                        :disabled="!allowWebModify"
                        @update:modelValue="(v: any) => onInput(field.key, v)"
                    />
                    <AnzuInput
                        v-else
                        :model-value="String(local[field.key] ?? '')"
                        :disabled="!allowWebModify"
                        @update:modelValue="(v: any) => onInput(field.key, v)"
                    />
                </div>
            </div>
        </div>

        <div
            v-if="allowWebModify"
            class="sticky bottom-2 flex items-center justify-end gap-2 rounded-xl bg-(--md-sys-color-surface) p-3 shadow"
        >
            <span
                v-if="dirtyKeys.length > 0"
                class="text-xs text-(--md-sys-color-on-surface-variant)"
            >
                {{ t("settings.systemConfig.dirtyCount", { count: dirtyKeys.length }) }}
            </span>
            <AnzuButton
                variant="text"
                :disabled="dirtyKeys.length === 0 || saving"
                @click="resetDirty"
            >
                {{ t("settings.systemConfig.discard") }}
            </AnzuButton>
            <AnzuButton
                variant="text"
                :disabled="overriddenKeys.length === 0 || saving"
                @click="onResetSelected"
            >
                {{ t("settings.systemConfig.resetOverridden") }}
            </AnzuButton>
            <AnzuButton
                :status="saving ? 'loading' : 'default'"
                :disabled="dirtyKeys.length === 0 || saving"
                @click="onSave"
            >
                {{ t("settings.systemConfig.save") }}
            </AnzuButton>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuComboBox from "~/components/AnzuComboBox.vue";
import AnzuTags from "~/components/AnzuTags.vue";
import AnzuCheckbox from "~/components/AnzuCheckbox.vue";
import type { FieldSchema, FieldValue } from "~/types/settings";

const { t, te } = useI18n();

const props = defineProps<{
    schema: FieldSchema[];
    values: FieldValue[];
    allowWebModify: boolean;
    saving: boolean;
}>();

const emit = defineEmits<{
    (e: "save", values: Record<string, string>): void;
    (e: "reset", keys: string[]): void;
}>();

const local = ref<Record<string, unknown>>({});
const initial = ref<Record<string, unknown>>({});

const overriddenKeys = computed(() =>
    props.values.filter((v) => v.overridden_in_db).map((v) => v.key),
);

const isOverridden = (key: string) =>
    overriddenKeys.value.includes(key);

const groupedSchema = computed(() => {
    const map = new Map<string, FieldSchema[]>();
    for (const f of props.schema) {
        const arr = map.get(f.group) ?? [];
        arr.push(f);
        map.set(f.group, arr);
    }
    return Array.from(map.entries()).map(([name, fields]) => ({ name, fields }));
});

const resolveControl = (f: FieldSchema): string => {
    if (f.type === "enum") return "enum";
    if (f.type === "bool") return "bool";
    if (f.type === "list") return "list";
    if (f.type === "multiline") return "multiline";
    if (f.type === "int" || f.type === "int64") return "int";
    return "string";
};

const dirtyKeys = computed(() => {
    const out: string[] = [];
    for (const f of props.schema) {
        if (
            JSON.stringify(initial.value[f.key]) !==
            JSON.stringify(local.value[f.key])
        ) {
            out.push(f.key);
        }
    }
    return out;
});

watch(
    () => props.values,
    (vs) => {
        const next: Record<string, unknown> = {};
        for (const v of vs) next[v.key] = v.value;
        local.value = { ...next };
        initial.value = { ...next };
    },
    { immediate: true },
);

const onInput = (key: string, v: unknown) => {
    local.value = { ...local.value, [key]: v };
};

const resetDirty = () => {
    local.value = { ...initial.value };
};

const serialize = (key: string): string => {
    const v = local.value[key];
    if (Array.isArray(v)) return JSON.stringify(v);
    if (typeof v === "boolean") return v ? "true" : "false";
    return String(v ?? "");
};

const onSave = () => {
    const payload: Record<string, string> = {};
    for (const k of dirtyKeys.value) payload[k] = serialize(k);
    emit("save", payload);
};

const onResetSelected = () => {
    emit("reset", overriddenKeys.value);
};
</script>
