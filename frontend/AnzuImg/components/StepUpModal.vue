<template>
    <div class="flex flex-col gap-4 p-2 min-w-80">
        <p class="text-sm text-(--md-sys-color-on-surface-variant)">
            {{ t("auth.stepUp.description") }}
        </p>

        <AnzuTabs v-if="hasMultipleMethods" v-model="activeMethod" :tabs="methodTabs">
            <template #tab-content-0>
                <PasswordForm
                    :running="running"
                    :error="error"
                    @submit="onPasswordSubmit"
                />
            </template>
            <template #tab-content-1>
                <PasskeyForm
                    :running="running"
                    :error="error"
                    @submit="onPasskeySubmit"
                />
            </template>
        </AnzuTabs>

        <PasswordForm
            v-else-if="onlyPassword"
            :running="running"
            :error="error"
            @submit="onPasswordSubmit"
        />
        <PasskeyForm
            v-else-if="onlyPasskey"
            :running="running"
            :error="error"
            @submit="onPasskeySubmit"
        />

        <div class="flex justify-end gap-2 pt-2">
            <AnzuButton variant="text" :disabled="running" @click="cancel">
                {{ t("common.actions.cancel") }}
            </AnzuButton>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, h, defineComponent } from "vue";
import AnzuTabs from "~/components/AnzuTabs.vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuInput from "~/components/AnzuInput.vue";

const { t } = useI18n();

const props = defineProps<{
    available: string[];
    maxAgeSeconds: number;
    runPassword: (password: string) => Promise<boolean>;
    runPasskey: () => Promise<boolean>;
    onResult?: (ok: boolean) => void;
}>();

const emit = defineEmits<{
    (e: "close"): void;
}>();

const running = ref(false);
const error = ref("");
const activeMethod = ref<string | number>(props.available[0] ?? "password");

const hasMultipleMethods = computed(() => props.available.length > 1);
const onlyPassword = computed(
    () => props.available.length === 1 && props.available[0] === "password",
);
const onlyPasskey = computed(
    () => props.available.length === 1 && props.available[0] === "passkey",
);
const methodTabs = computed(() => [
    { label: t("auth.stepUp.byPassword"), value: "password" },
    { label: t("auth.stepUp.byPasskey"), value: "passkey" },
]);

const onPasswordSubmit = async (password: string) => {
    if (running.value) return;
    running.value = true;
    error.value = "";
    const ok = await props.runPassword(password);
    running.value = false;
    if (ok) {
        props.onResult?.(true);
        emit("close");
    } else {
        error.value = t("auth.stepUp.passwordFailed");
    }
};

const onPasskeySubmit = async () => {
    if (running.value) return;
    running.value = true;
    error.value = "";
    const ok = await props.runPasskey();
    running.value = false;
    if (ok) {
        props.onResult?.(true);
        emit("close");
    } else {
        error.value = t("auth.stepUp.passkeyFailed");
    }
};

const cancel = () => {
    props.onResult?.(false);
    emit("close");
};

const PasswordForm = defineComponent({
    components: { AnzuInput, AnzuButton },
    props: {
        running: Boolean,
        error: String,
    },
    emits: ["submit"],
    setup(p, { emit }) {
        const password = ref("");
        const submit = () => {
            if (!password.value) return;
            emit("submit", password.value);
        };
        return () =>
            h("form", {
                onSubmit: (e: Event) => {
                    e.preventDefault();
                    submit();
                },
                class: "flex flex-col gap-3",
            }, [
                h(AnzuInput, {
                    type: "password",
                    label: t("auth.stepUp.passwordLabel"),
                    placeholder: t("auth.stepUp.passwordPlaceholder"),
                    modelValue: password.value,
                    "onUpdate:modelValue": (v: string) => (password.value = v),
                    name: "stepup-password",
                    autocomplete: "current-password",
                }),
                p.error
                    ? h(
                          "p",
                          { class: "text-xs text-(--md-sys-color-error)" },
                          p.error,
                      )
                    : null,
                h(
                    AnzuButton,
                    {
                        type: "submit",
                        status: p.running ? "loading" : "default",
                        disabled: p.running || !password.value,
                    },
                    () => t("auth.stepUp.confirm"),
                ),
            ]);
    },
});

const PasskeyForm = defineComponent({
    components: { AnzuButton },
    props: {
        running: Boolean,
        error: String,
    },
    emits: ["submit"],
    setup(p, { emit }) {
        return () =>
            h("div", { class: "flex flex-col gap-3 items-stretch" }, [
                h(
                    "p",
                    { class: "text-xs text-(--md-sys-color-on-surface-variant)" },
                    t("auth.stepUp.passkeyHint"),
                ),
                p.error
                    ? h(
                          "p",
                          { class: "text-xs text-(--md-sys-color-error)" },
                          p.error,
                      )
                    : null,
                h(
                    AnzuButton,
                    {
                        status: p.running ? "loading" : "default",
                        disabled: p.running,
                        onClick: () => emit("submit"),
                    },
                    () => t("auth.stepUp.usePasskey"),
                ),
            ]);
    },
});
</script>
