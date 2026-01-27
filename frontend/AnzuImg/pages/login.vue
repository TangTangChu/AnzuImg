<template>
  <div class="flex min-h-[calc(100vh-8rem)] items-center justify-center">
    <div class="w-full max-w-md rounded-xl">
      <h1 class="mb-6 text-center text-2xl font-bold">
        {{ t("login.title") }}
      </h1>

      <form @submit.prevent="handleLogin" class="flex flex-col gap-4">
        <AnzuInput v-model="password" type="password" :label="t('common.labels.password')"
          placeholder="Enter password" />

        <AnzuButton type="submit" :status="loading ? 'loading' : 'default'" class="w-full">
          {{ t("common.actions.login") }}
        </AnzuButton>
      </form>

      <AnzuDivider>OR</AnzuDivider>

      <AnzuButton variant="outlined" class="w-full" :status="loading ? 'loading' : 'default'"
        @click="handlePasskeyLogin">
        {{ t("login.passkeyButton") }}
      </AnzuButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import { useAuth } from "~/composables/useAuth";
import { useNotification } from "~/composables/useNotification";
import { NotificationType } from "~/types/notification";

const { t } = useI18n();
const password = ref("");
const loading = ref(false);
const { login, loginWithPasskey } = useAuth();
const router = useRouter();
const { notify } = useNotification();

const handleLogin = async () => {
  if (!password.value) return;
  loading.value = true;

  const success = await login(password.value);
  if (success) {
    notify({
      message: t("login.success"),
      type: NotificationType.SUCCESS,
    });
    router.push("/gallery");
  } else {
    notify({
      message: "Login failed. Check your Password.",
      type: NotificationType.ERROR,
    });
  }
  loading.value = false;
};

const handlePasskeyLogin = async () => {
  loading.value = true;

  const success = await loginWithPasskey();
  if (success) {
    notify({
      message: t("login.success"),
      type: NotificationType.SUCCESS,
    });
    router.push("/gallery");
  } else {
    notify({
      message: "Passkey login failed.",
      type: NotificationType.ERROR,
    });
  }
  loading.value = false;
};
</script>
