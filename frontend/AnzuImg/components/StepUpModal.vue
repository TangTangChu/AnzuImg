<template>
    <div class="flex flex-col gap-4 p-2 min-w-80">
        <p class="text-sm text-(--md-sys-color-on-surface-variant)">
            {{ t("auth.stepUp.description") }}
        </p>

        <form v-if="hasPassword" @submit.prevent="onPasswordSubmit" class="flex flex-col gap-3">
            <AnzuInput
                v-model="password"
                type="password"
                :label="t('auth.stepUp.passwordLabel')"
                :placeholder="t('auth.stepUp.passwordPlaceholder')"
                name="stepup-password"
                autocomplete="current-password"
            />
            <p v-if="error" class="text-xs text-(--md-sys-color-error)">{{ error }}</p>
            <div class="flex items-center gap-3">
                <AnzuButton type="submit" :status="running ? 'loading' : 'default'" :disabled="running || !password">
                    {{ t("auth.stepUp.confirm") }}
                </AnzuButton>
                <AnzuButton
                    v-if="hasPasskey"
                    variant="text"
                    :disabled="running"
                    @click="onPasskeySubmit"
                >
                    {{ t("auth.stepUp.usePasskey") }}
                </AnzuButton>
            </div>
        </form>

        <div v-else-if="hasPasskey" class="flex flex-col gap-3">
            <p class="text-xs text-(--md-sys-color-on-surface-variant)">
                {{ t("auth.stepUp.passkeyHint") }}
            </p>
            <p v-if="error" class="text-xs text-(--md-sys-color-error)">{{ error }}</p>
            <AnzuButton :status="running ? 'loading' : 'default'" :disabled="running" @click="onPasskeySubmit">
                {{ t("auth.stepUp.usePasskey") }}
            </AnzuButton>
        </div>

        <div class="flex justify-end pt-2">
            <AnzuButton variant="text" :disabled="running" @click="cancel">
                {{ t("common.actions.cancel") }}
            </AnzuButton>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
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
const password = ref("");

const hasPassword = computed(() => props.available.includes("password"));
const hasPasskey = computed(() => props.available.includes("passkey"));

const onPasswordSubmit = async () => {
    if (running.value || !password.value) return;
    running.value = true;
    error.value = "";
    const ok = await props.runPassword(password.value);
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
</script>
