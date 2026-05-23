<template>
    <div class="space-y-10">
        <div
            v-if="!allowWebModify"
            class="flex items-start gap-2 rounded-lg border border-(--md-sys-color-outline-variant) p-3 text-sm text-(--md-sys-color-on-surface-variant)"
        >
            <InformationCircleIcon class="w-5 h-5 shrink-0 mt-0.5 text-(--md-sys-color-on-surface-variant)" />
            <span>{{ t("settings.systemConfig.disabledNotice") }}</span>
        </div>

        <div
            v-for="(group, groupIdx) in groupedSchema"
            :key="group.name"
            :class="groupIdx > 0 ? 'pt-8 border-t border-(--md-sys-color-outline-variant)' : ''"
        >
            <h3 class="mb-5 text-lg font-semibold">
                {{ t(`settings.systemConfig.groups.${group.name}`) }}
            </h3>
            <div class="grid grid-cols-1 gap-x-8 gap-y-5 md:grid-cols-2">
                <div
                    v-for="field in group.regular"
                    :key="field.key"
                    class="flex flex-col gap-1.5"
                    :class="isWideField(field) ? 'md:col-span-2' : ''"
                >
                    <div class="flex items-center justify-between gap-2">
                        <label class="text-sm font-medium text-(--md-sys-color-on-surface)">
                            {{ t(`settings.systemConfig.fields.${field.key}.label`) }}
                        </label>
                        <span
                            v-if="isOverridden(field.key)"
                            class="inline-flex items-center gap-1 text-xs text-(--md-sys-color-on-surface-variant)"
                        >
                            <span class="w-1.5 h-1.5 rounded-full bg-(--md-sys-color-primary)"></span>
                            {{ t("settings.systemConfig.overridden") }}
                        </span>
                    </div>
                    <p
                        v-if="te(`settings.systemConfig.fields.${field.key}.hint`)"
                        class="text-xs text-(--md-sys-color-on-surface-variant)"
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
                        class="w-full rounded-lg border border-(--md-sys-color-outline) bg-transparent px-3 py-2 text-sm text-(--md-sys-color-on-surface) outline-none transition-[border-color] duration-200 ease-out hover:border-(--md-sys-color-outline-variant) focus:border-(--md-sys-color-primary)"
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

                <div
                    v-if="group.compactBools.length > 0"
                    class="md:col-span-2 flex flex-wrap gap-x-6 gap-y-3"
                >
                    <div
                        v-for="field in group.compactBools"
                        :key="field.key"
                        class="inline-flex items-center gap-1.5"
                    >
                        <AnzuCheckbox
                            :model-value="!!local[field.key]"
                            :label="t(`settings.systemConfig.fields.${field.key}.label`)"
                            :disabled="!allowWebModify"
                            @update:modelValue="(v: any) => onInput(field.key, !!v)"
                        />
                        <span
                            v-if="isOverridden(field.key)"
                            :title="t('settings.systemConfig.overridden')"
                            class="w-1.5 h-1.5 rounded-full bg-(--md-sys-color-primary)"
                        ></span>
                    </div>
                </div>
            </div>
        </div>

        <div
            v-if="allowWebModify"
            class="sticky bottom-2 flex items-center justify-end gap-2 rounded-xl border border-(--md-sys-color-outline-variant) bg-(--md-sys-color-surface) px-4 py-2"
        >
            <span
                v-if="dirtyKeys.length > 0"
                class="text-xs text-(--md-sys-color-on-surface-variant) mr-auto"
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
import { InformationCircleIcon } from "@heroicons/vue/24/outline";
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

const resolveControl = (f: FieldSchema): string => {
    if (f.type === "enum") return "enum";
    if (f.type === "bool") return "bool";
    if (f.type === "list") return "list";
    if (f.type === "multiline") return "multiline";
    if (f.type === "int" || f.type === "int64") return "int";
    return "string";
};

const isWideField = (f: FieldSchema): boolean => {
    const c = resolveControl(f);
    return c === "list" || c === "multiline";
};

const isCompactBool = (f: FieldSchema): boolean => {
    return (
        resolveControl(f) === "bool" &&
        !te(`settings.systemConfig.fields.${f.key}.hint`)
    );
};

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
    return Array.from(map.entries()).map(([name, fields]) => {
        const compactBools = fields.filter((f) => isCompactBool(f));
        const regular = fields.filter((f) => !isCompactBool(f));
        return { name, regular, compactBools };
    });
});

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

