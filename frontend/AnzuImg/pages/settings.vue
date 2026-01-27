<template>
  <h1 class="mb-6 text-3xl font-bold text-center">{{ t('settings.title') }}</h1>
  <!-- 修改密码 -->
  <div class="mb-12 max-w-3xl mx-auto">
    <h2 class="mb-4 text-xl font-semibold">{{ t('settings.changePassword.title') }}
    </h2>
    <p class="mb-6 text-(--md-sys-color-on-surface-variant)">{{ t('settings.changePassword.description') }}</p>

    <form @submit.prevent="handleChangePassword" class="flex flex-col gap-4" autocomplete="on">
      <input type="text" name="username" autocomplete="username" value="anzuimg" style="display: none;" />

      <AnzuInput v-model="passwordForm.currentPassword" type="password"
        :label="t('settings.changePassword.currentPassword')"
        :placeholder="t('settings.changePassword.currentPasswordPlaceholder')" :disabled="changingPassword"
        name="current-password" autocomplete="current-password" />

      <AnzuInput v-model="passwordForm.newPassword" type="password" :label="t('settings.changePassword.newPassword')"
        :placeholder="t('settings.changePassword.newPasswordPlaceholder')" :disabled="changingPassword"
        name="new-password" autocomplete="new-password" />

      <AnzuInput v-model="passwordForm.confirmPassword" type="password"
        :label="t('settings.changePassword.confirmPassword')"
        :placeholder="t('settings.changePassword.confirmPasswordPlaceholder')" :disabled="changingPassword"
        name="confirm-new-password" autocomplete="new-password" />

      <AnzuButton type="submit" :status="changingPassword ? 'loading' : 'default'" class="w-full sm:w-auto">
        {{ t('settings.changePassword.submit') }}
      </AnzuButton>
    </form>
  </div>

  <!-- PassKey管理 -->
  <div class="mb-12 max-w-3xl mx-auto">
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold">{{ t('settings.passkeyManagement.title') }}
        </h2>
        <p class="mt-1 text-(--md-sys-color-on-surface-variant)">{{ t('settings.passkeyManagement.description') }}</p>
      </div>
      <AnzuButton @click="handleRegisterPasskey" :status="registeringPasskey ? 'loading' : 'default'"
        variant="outlined">
        {{ t('settings.passkeyManagement.registerNew') }}
      </AnzuButton>
    </div>

    <!-- PassKey列表 -->
    <div v-if="loadingPasskeys" class="flex justify-center py-8">
      <AnzuProgressRing :size="48" />
    </div>

    <div v-else-if="passkeys.length === 0"
      class="rounded-lg border border-(--md-sys-color-outline-variant) p-8 text-center">
      <p class="text-(--md-sys-color-on-surface-variant)">{{ t('settings.passkeyManagement.noPasskeys') }}</p>
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-1 lg:grid-cols-2">
      <div v-for="passkey in passkeys" :key="passkey.id"
        class="relative flex flex-col justify-between rounded-xl border border-(--md-sys-color-outline-variant) p-4 transition-colors">

        <div class="flex items-start justify-between mb-3">
          <div class="flex items-start gap-3 overflow-hidden">
            <div
              class="rounded-full bg-(--md-sys-color-secondary-container) p-2.5 text-(--md-sys-color-on-secondary-container) shrink-0">
              <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                  :d="getDeviceIcon(passkey.UserAgent || '')" />
              </svg>
            </div>
            <div class="min-w-0">
              <h3 class="font-semibold text-(--md-sys-color-on-surface) truncate"
                :title="passkey.DeviceName || `Passkey #${passkey.ID}`">
                {{ passkey.DeviceName || `Passkey #${passkey.ID}` }}
              </h3>
              <div class="flex flex-wrap gap-2 mt-1.5">
                <span v-if="passkey.IPAddress"
                  class="text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-surface-variant) text-(--md-sys-color-on-surface-variant)">
                  {{ passkey.IPAddress }}
                </span>
                <span class="text-xs text-(--md-sys-color-on-surface-variant) flex items-center"
                  :title="formatDate(passkey.UpdatedAt || passkey.CreatedAt)">
                  {{ formatRelativeTime(passkey.UpdatedAt || passkey.CreatedAt, locale) }}
                </span>
              </div>
            </div>
          </div>

          <AnzuButton @click="() => handleDeletePasskey(passkey.CredentialID)" variant="text"
            class="min-w-0! p-2! h-9! w-9! shrink-0 -mr-2 -mt-2 text-(--md-sys-color-error)"
            :status="deletingPasskeyId === passkey.CredentialID ? 'loading' : 'default'">
            <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
              v-if="deletingPasskeyId !== passkey.CredentialID">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </AnzuButton>
        </div>

        <div class="mt-2 pt-3 border-t border-(--md-sys-color-outline-variant) border-opacity-50 space-y-1">
          <div v-if="passkey.UserAgent" class="text-xs text-(--md-sys-color-on-surface-variant) opacity-75 truncate"
            :title="passkey.UserAgent">
            {{ passkey.UserAgent }}
          </div>
          <div class="flex items-center justify-between gap-2">
            <div class="text-xs text-(--md-sys-color-on-surface-variant) opacity-60 font-mono truncate flex-1"
              :title="passkey.CredentialID">
              ID: {{ passkey.CredentialID }}
            </div>
            <div class="text-xs text-(--md-sys-color-on-surface-variant) opacity-60 shrink-0">
              {{ formatDate(passkey.CreatedAt).split(' ')[0] }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- API Token Management -->
  <div class="max-w-3xl mx-auto">
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold">API Token</h2>
        <p class="mt-1 text-(--md-sys-color-on-surface-variant)">{{ t('settings.apiTokens.description') }}</p>
      </div>
      <AnzuButton @click="showCreateTokenDialog = true" variant="outlined">
        {{ t('settings.apiTokens.createNew') }}
      </AnzuButton>
    </div>

    <!-- Token List -->
    <div v-if="loadingTokens" class="flex justify-center py-8">
      <AnzuProgressRing :size="48" />
    </div>
    <div v-else-if="apiTokens.length === 0"
      class="rounded-lg border border-(--md-sys-color-outline-variant) p-8 text-center">
      <p class="text-(--md-sys-color-on-surface-variant)">{{ t('settings.apiTokens.noTokens') }}</p>
    </div>
    <div v-else class="grid gap-4 sm:grid-cols-1">
      <div v-for="token in apiTokens" :key="token.id"
        class="flex flex-col justify-between rounded-xl border border-(--md-sys-color-outline-variant) p-4 transition-colors">
        <div class="flex items-start justify-between">
          <div>
            <h3 class="font-semibold text-(--md-sys-color-on-surface)">{{ token.name }}</h3>
            <div class="mt-1 flex flex-wrap gap-2">
              <span v-if="!token.ip_allowlist || token.ip_allowlist.length === 0"
                class="text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-error-container) text-(--md-sys-color-on-error-container)">
                Any IP
              </span>
              <span v-else v-for="ip in token.ip_allowlist" :key="ip"
                class="text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-surface-variant) text-(--md-sys-color-on-surface-variant)">
                {{ ip }}
              </span>
            </div>
          </div>
          <AnzuButton @click="() => handleDeleteToken(token.id)" variant="text"
            class="min-w-0! p-2! h-9! w-9! shrink-0 text-(--md-sys-color-error)">
            <TrashIcon class="h-5 w-5" />
          </AnzuButton>
        </div>
        <div
          class="mt-2 pt-2 border-t border-(--md-sys-color-outline-variant) border-opacity-50 flex justify-between text-xs text-(--md-sys-color-on-surface-variant) opacity-75">
          <span>{{ t('settings.apiTokens.created') }}: {{ formatDate(token.created_at) }}</span>
          <span>{{ t('settings.apiTokens.lastUsed') }}:
            {{ token.last_used_at ? formatRelativeTime(token.last_used_at, locale) : t('settings.apiTokens.neverUsed') }}</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Create Token Dialog -->
  <AnzuDialog v-model:visible="showCreateTokenDialog" :title="t('settings.apiTokens.createNew')" :actions="[
    { text: t('common.actions.cancel'), variant: 'text', handler: () => { showCreateTokenDialog = false } },
    { text: t('settings.apiTokens.createNew'), primary: true, variant: 'filled', handler: handleCreateToken, loading: creatingToken }
  ]">
    <div class="flex flex-col gap-4">
      <AnzuInput v-model="tokenForm.name" :label="t('settings.apiTokens.name')"
        :placeholder="t('settings.apiTokens.namePlaceholder')" name="token-name" autocomplete="off" />
      <AnzuTags v-model="tokenForm.ipAllowlist" :label="t('settings.apiTokens.ipAllowlist')" :max-tags="10"
        :hint="t('settings.apiTokens.ipAllowlistTip')" />
    </div>
  </AnzuDialog>

  <!-- Show Token Dialog -->
  <AnzuDialog v-model:visible="showTokenResultDialog" :title="t('settings.apiTokens.tokenCreatedTitle')" :actions="[
    { text: t('common.actions.close'), variant: 'filled', handler: () => { showTokenResultDialog = false } }
  ]">
    <div class="flex flex-col gap-4">
      <p class="text-(--md-sys-color-on-surface-variant)">{{ t('settings.apiTokens.tokenCreatedMessage') }}</p>
      <div class="relative">
        <AnzuInput :model-value="createdTokenRaw" readonly name="api-token" autocomplete="off" />
        <AnzuButton @click="copyToken" variant="tonal" class="absolute right-1 top-1 bottom-1">
          {{ t('settings.apiTokens.copy') }}
        </AnzuButton>
      </div>
    </div>
  </AnzuDialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AnzuButton from '~/components/AnzuButton.vue'
import AnzuInput from '~/components/AnzuInput.vue'
import AnzuTags from '~/components/AnzuTags.vue'
import AnzuDialog from '~/components/AnzuDialog.vue'
import AnzuProgressRing from '~/components/AnzuProgressRing.vue'
import { useAuth } from '~/composables/useAuth'
import { useNotification } from '~/composables/useNotification'
import { useDialog } from '~/composables/useDialog'
import { NotificationType } from '~/types/notification'
import { DialogVariant } from '~/types/dialog'
import { formatDate, formatRelativeTime } from "~/utils/format";
import { validatePassword } from '~/utils/password';
import { TrashIcon } from "@heroicons/vue/24/outline";
import type { APIToken } from '~/types/api_token';

const { t, locale } = useI18n()
const { changePassword, logout, listPasskeys, deletePasskey, registerPasskey, checkPasskeyExists, createAPIToken, listAPITokens, deleteAPIToken } = useAuth()
const { notify } = useNotification()
const { confirm } = useDialog()

const passwordForm = ref({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})
const changingPassword = ref(false)

// PassKey管理
const passkeys = ref<any[]>([])
const loadingPasskeys = ref(false)
const registeringPasskey = ref(false)
const deletingPasskeyId = ref<string | null>(null)

// API Tokens state
const apiTokens = ref<APIToken[]>([]);
const loadingTokens = ref(false);
const showCreateTokenDialog = ref(false);
const creatingToken = ref(false);
const tokenForm = ref({ name: '', ipAllowlist: [] as string[] });
const showTokenResultDialog = ref(false);
const createdTokenRaw = ref('');

const loadTokens = async () => {
  loadingTokens.value = true;
  apiTokens.value = await listAPITokens();
  loadingTokens.value = false;
};

const handleCreateToken = async () => {
  if (!tokenForm.value.name) return;
  creatingToken.value = true;
  const res = await createAPIToken(tokenForm.value.name, tokenForm.value.ipAllowlist);
  creatingToken.value = false;
  if (res) {
    createdTokenRaw.value = res.raw_token;
    showCreateTokenDialog.value = false;
    showTokenResultDialog.value = true;
    loadTokens();
    tokenForm.value = { name: '', ipAllowlist: [] };
    notify({ message: t('settings.apiTokens.createSuccess'), type: NotificationType.SUCCESS });
  } else {
    notify({ message: t('settings.apiTokens.createFailed'), type: NotificationType.ERROR });
  }
};

const handleDeleteToken = async (id: number) => {
  const result = await confirm(t('common.actions.deleteConfirm'), {
    title: t('common.actions.delete'),
    variant: DialogVariant.DESTRUCTIVE,
    actions: [
      { text: t('common.actions.cancel'), variant: 'text' },
      { text: t('common.actions.delete'), primary: true, variant: 'filled' }
    ]
  });
  if (!result) return;

  if (await deleteAPIToken(id)) {
    loadTokens();
    notify({ message: t('common.actions.deleteSuccess'), type: NotificationType.SUCCESS });
  } else {
    notify({ message: t('common.actions.deleteFailed'), type: NotificationType.ERROR });
  }
};

const copyToken = () => {
  navigator.clipboard.writeText(createdTokenRaw.value);
  notify({ message: t('settings.apiTokens.copySuccess'), type: NotificationType.SUCCESS });
};

// 修改密码
const handleChangePassword = async () => {
  if (!passwordForm.value.currentPassword || !passwordForm.value.newPassword || !passwordForm.value.confirmPassword) {
    notify({
      message: t('settings.changePassword.fillAllFields'),
      type: NotificationType.WARNING
    })
    return
  }

  const validation = validatePassword(passwordForm.value.newPassword, t)
  if (!validation.valid) {
    notify({
      message: validation.error!,
      type: NotificationType.WARNING
    })
    return
  }

  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    notify({
      message: t('settings.changePassword.passwordMatchError'),
      type: NotificationType.WARNING
    })
    return
  }

  changingPassword.value = true

  const success = await changePassword(passwordForm.value.currentPassword, passwordForm.value.newPassword)
  if (success) {
    notify({
      message: t('settings.changePassword.success'),
      type: NotificationType.SUCCESS
    })
    // 修改密码成功后自动登出
    setTimeout(() => {
      logout()
    }, 1500)
  } else {
    notify({
      message: t('settings.changePassword.failed'),
      type: NotificationType.ERROR
    })
  }

  changingPassword.value = false
}

// 注册新PassKey
const handleRegisterPasskey = async () => {
  registeringPasskey.value = true

  const success = await registerPasskey()
  if (success) {
    notify({
      message: t('settings.passkeyManagement.registerSuccess'),
      type: NotificationType.SUCCESS
    })
    loadPasskeys()
  } else {
    notify({
      message: t('settings.passkeyManagement.registerFailed'),
      type: NotificationType.ERROR
    })
  }

  registeringPasskey.value = false
}

// 删除PassKey
const handleDeletePasskey = async (credentialId: string) => {
  try {
    const result = await confirm(t('common.actions.deleteConfirm'), {
      title: t('common.actions.delete'),
      variant: DialogVariant.DESTRUCTIVE,
      actions: [
        { text: t('common.actions.cancel'), variant: 'text' },
        {
          text: t('common.actions.delete'),
          primary: true,
          variant: 'filled',
          loading: false
        }
      ]
    })

    if (!result) return

    deletingPasskeyId.value = credentialId

    const success = await deletePasskey(credentialId)
    if (success) {
      notify({
        message: t('common.actions.deleteSuccess'),
        type: NotificationType.SUCCESS
      })
      loadPasskeys()
    } else {
      notify({
        message: t('common.actions.deleteFailed'),
        type: NotificationType.ERROR
      })
    }
  } catch (e: any) {
    if (e.message === 'Dialog closed' || e.message === 'All dialogs closed') {
      return
    }
    notify({
      message: t('common.actions.deleteFailed'),
      type: NotificationType.ERROR
    })
  } finally {
    deletingPasskeyId.value = null
  }
}

// 加载PassKey列表
const loadPasskeys = async () => {
  loadingPasskeys.value = true
  passkeys.value = await listPasskeys()
  loadingPasskeys.value = false
}

const getDeviceIcon = (ua: string = '') => {
  ua = ua.toLowerCase();
  if (ua.includes('mobile') || ua.includes('android') || ua.includes('iphone') || ua.includes('ipad')) {
    return "M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3";
  }
  return "M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25";
}

onMounted(() => {
  loadPasskeys()
  loadTokens()
})
</script>
