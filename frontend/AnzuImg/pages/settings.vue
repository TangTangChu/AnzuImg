<template>
  <h1 class="mb-6 text-3xl font-bold text-center">{{ t("settings.title") }}</h1>

  <Dashboard />

  <!-- 修改密码 -->
  <div class="mb-12 max-w-3xl mx-auto">
    <h2 class="mb-4 text-xl font-semibold">
      {{ t("settings.changePassword.title") }}
    </h2>
    <p class="mb-6 text-(--md-sys-color-on-surface-variant)">
      {{ t("settings.changePassword.description") }}
    </p>

    <form
      @submit.prevent="handleChangePassword"
      class="flex flex-col gap-4"
      autocomplete="on"
    >
      <input
        type="text"
        name="username"
        autocomplete="username"
        value="anzuimg"
        style="display: none"
      />

      <AnzuInput
        v-model="passwordForm.currentPassword"
        type="password"
        :label="t('settings.changePassword.currentPassword')"
        :placeholder="t('settings.changePassword.currentPasswordPlaceholder')"
        :disabled="changingPassword"
        name="current-password"
        autocomplete="current-password"
      />

      <AnzuInput
        v-model="passwordForm.newPassword"
        type="password"
        :label="t('settings.changePassword.newPassword')"
        :placeholder="t('settings.changePassword.newPasswordPlaceholder')"
        :disabled="changingPassword"
        name="new-password"
        autocomplete="new-password"
      />

      <AnzuInput
        v-model="passwordForm.confirmPassword"
        type="password"
        :label="t('settings.changePassword.confirmPassword')"
        :placeholder="t('settings.changePassword.confirmPasswordPlaceholder')"
        :disabled="changingPassword"
        name="confirm-new-password"
        autocomplete="new-password"
      />

      <AnzuButton
        type="submit"
        :status="changingPassword ? 'loading' : 'default'"
        class="w-full sm:w-auto"
      >
        {{ t("settings.changePassword.submit") }}
      </AnzuButton>
    </form>
  </div>

  <!-- PassKey管理 -->
  <div class="mb-12 max-w-3xl mx-auto">
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold">
          {{ t("settings.passkeyManagement.title") }}
        </h2>
        <p class="mt-1 text-(--md-sys-color-on-surface-variant)">
          {{ t("settings.passkeyManagement.description") }}
        </p>
      </div>
      <AnzuButton
        @click="handleRegisterPasskey"
        :status="registeringPasskey ? 'loading' : 'default'"
        variant="outlined"
      >
        {{ t("settings.passkeyManagement.registerNew") }}
      </AnzuButton>
    </div>

    <!-- PassKey列表 -->
    <div v-if="loadingPasskeys" class="flex justify-center py-8">
      <AnzuProgressRing :size="48" />
    </div>

    <div
      v-else-if="passkeys.length === 0"
      class="rounded-lg border border-(--md-sys-color-outline-variant) p-8 text-center"
    >
      <p class="text-(--md-sys-color-on-surface-variant)">
        {{ t("settings.passkeyManagement.noPasskeys") }}
      </p>
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-1 lg:grid-cols-2">
      <div
        v-for="passkey in passkeys"
        :key="passkey.ID"
        class="relative flex flex-col justify-between rounded-xl border border-(--md-sys-color-outline-variant) p-4 transition-colors min-w-0"
      >
        <div class="flex items-start justify-between mb-3">
          <div class="flex items-start gap-3 overflow-hidden min-w-0">
            <div
              class="rounded-full bg-(--md-sys-color-secondary-container) p-2.5 text-(--md-sys-color-on-secondary-container) shrink-0"
            >
              <svg
                class="h-6 w-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="1.5"
                  :d="getDeviceIcon(passkey.UserAgent || '')"
                />
              </svg>
            </div>
            <div class="min-w-0">
              <h3
                class="font-semibold text-(--md-sys-color-on-surface) truncate"
                :title="passkey.DeviceName || `Passkey #${passkey.ID}`"
              >
                {{ passkey.DeviceName || `Passkey #${passkey.ID}` }}
              </h3>
              <div class="flex flex-wrap gap-2 mt-1.5">
                <span
                  v-if="passkey.IPAddress"
                  class="text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-surface-variant) text-(--md-sys-color-on-surface-variant)"
                >
                  {{ passkey.IPAddress }}
                </span>
                <span
                  class="text-xs text-(--md-sys-color-on-surface-variant) flex items-center"
                  :title="formatDate(passkey.UpdatedAt || passkey.CreatedAt)"
                >
                  {{
                    formatRelativeTime(
                      passkey.UpdatedAt || passkey.CreatedAt,
                      locale,
                    )
                  }}
                </span>
              </div>
            </div>
          </div>

          <AnzuButton
            @click="() => handleDeletePasskey(passkey.CredentialID)"
            variant="text"
            class="min-w-0! p-2! h-9! w-9! shrink-0 -mr-2 -mt-2 text-(--md-sys-color-error)"
            :status="
              deletingPasskeyId === passkey.CredentialID ? 'loading' : 'default'
            "
          >
            <svg
              class="h-5 w-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              v-if="deletingPasskeyId !== passkey.CredentialID"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </AnzuButton>
        </div>

        <div
          class="mt-2 pt-3 border-t border-(--md-sys-color-outline-variant) border-opacity-50 space-y-1"
        >
          <div
            v-if="passkey.UserAgent"
            class="text-xs text-(--md-sys-color-on-surface-variant) opacity-75 truncate min-w-0"
            :title="passkey.UserAgent"
          >
            {{ passkey.UserAgent }}
          </div>
          <div class="flex items-center justify-between gap-2">
            <div
              class="text-xs text-(--md-sys-color-on-surface-variant) opacity-60 font-mono truncate flex-1 min-w-0"
              :title="passkey.CredentialID"
            >
              ID: {{ passkey.CredentialID }}
            </div>
            <div
              class="text-xs text-(--md-sys-color-on-surface-variant) opacity-60 shrink-0"
            >
              {{ formatDate(passkey.CreatedAt).split(" ")[0] }}
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
        <p class="mt-1 text-(--md-sys-color-on-surface-variant)">
          {{ t("settings.apiTokens.description") }}
        </p>
      </div>
      <AnzuButton @click="showCreateTokenDialog = true" variant="outlined">
        {{ t("settings.apiTokens.createNew") }}
      </AnzuButton>
    </div>

    <!-- Token List -->
    <div v-if="loadingTokens" class="flex justify-center py-8">
      <AnzuProgressRing :size="48" />
    </div>
    <div
      v-else-if="apiTokens.length === 0"
      class="rounded-lg border border-(--md-sys-color-outline-variant) p-8 text-center"
    >
      <p class="text-(--md-sys-color-on-surface-variant)">
        {{ t("settings.apiTokens.noTokens") }}
      </p>
    </div>
    <div v-else class="grid gap-4 sm:grid-cols-1">
      <div
        v-for="token in apiTokens"
        :key="token.id"
        class="flex flex-col justify-between rounded-xl border border-(--md-sys-color-outline-variant) p-4 transition-colors"
      >
        <div class="flex items-start justify-between">
          <div>
            <h3 class="font-semibold text-(--md-sys-color-on-surface)">
              {{ token.name }}
            </h3>
            <div class="mt-1 flex flex-wrap gap-2">
              <span
                class="text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)"
              >
                {{ getTokenTypeLabel(token.token_type) }}
              </span>
              <span
                v-if="!token.ip_allowlist || token.ip_allowlist.length === 0"
                class="text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-error-container) text-(--md-sys-color-on-error-container)"
              >
                {{ t("settings.apiTokens.anyIP") }}
              </span>
              <span
                v-else
                v-for="ip in token.ip_allowlist"
                :key="ip"
                class="text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-surface-variant) text-(--md-sys-color-on-surface-variant)"
              >
                {{ ip }}
              </span>
            </div>
          </div>
          <AnzuButton
            @click="() => handleDeleteToken(token.id)"
            variant="text"
            class="min-w-0! p-2! h-9! w-9! shrink-0 text-(--md-sys-color-error)"
          >
            <TrashIcon class="h-5 w-5" />
          </AnzuButton>
        </div>
        <div
          class="mt-2 pt-2 border-t border-(--md-sys-color-outline-variant) border-opacity-50 flex justify-between text-xs text-(--md-sys-color-on-surface-variant) opacity-75"
        >
          <span
            >{{ t("settings.apiTokens.created") }}:
            {{ formatDate(token.created_at) }}</span
          >
          <span
            >{{ t("settings.apiTokens.lastUsed") }}:
            {{
              token.last_used_at
                ? formatRelativeTime(token.last_used_at, locale)
                : t("settings.apiTokens.neverUsed")
            }}</span
          >
        </div>
      </div>
    </div>
  </div>

  <div class="max-w-3xl mx-auto mt-10">
    <AnzuTabs v-model="activeLogTab" :tabs="logTabs">
      <template #tab-content-0>
        <!-- Token Logs -->
        <div>
          <div class="mb-4 flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold">
                {{ t("settings.apiTokens.logsTitle") }}
              </h3>
              <p class="mt-1 text-(--md-sys-color-on-surface-variant)">
                {{ t("settings.apiTokens.logsDescription") }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <AnzuInput
                v-model.number="logRetentionDays"
                type="number"
                min="1"
                class="w-28"
                :placeholder="t('settings.apiTokens.cleanupDaysPlaceholder')"
              />
              <AnzuButton
                variant="text"
                :status="cleaningTokenLogs ? 'loading' : 'default'"
                @click="handleCleanupTokenLogs"
              >
                {{ t("settings.apiTokens.cleanupButton") }}
              </AnzuButton>
            </div>
          </div>

          <div class="mb-4 grid grid-cols-1 gap-4 sm:grid-cols-4">
            <AnzuInput
              v-model="tokenLogFilters.search"
              :placeholder="t('common.actions.search')"
              @keyup.enter="loadTokenLogs(1)"
            />
            <AnzuComboBox
              v-model="tokenLogFilters.type"
              :items="tokenLogTypeOptions"
              :placeholder="t('settings.apiTokens.tokenTypePlaceholder')"
              @update:modelValue="loadTokenLogs(1)"
            />
            <AnzuInput
              v-model="tokenLogFilters.startDate"
              type="date"
              @change="loadTokenLogs(1)"
            />
            <AnzuInput
              v-model="tokenLogFilters.endDate"
              type="date"
              @change="loadTokenLogs(1)"
            />
          </div>
          <div class="mb-2 flex justify-end">
             <AnzuButton variant="text" @click="() => loadTokenLogs(1)">
                {{ t("settings.apiTokens.refreshLogs") }}
              </AnzuButton>
          </div>

          <div v-if="loadingTokenLogs" class="flex justify-center py-6">
            <AnzuProgressRing :size="40" />
          </div>
          <div
            v-else-if="tokenLogs.length === 0"
            class="rounded-lg border border-(--md-sys-color-outline-variant) p-6 text-center"
          >
            <p class="text-(--md-sys-color-on-surface-variant)">
              {{ t("settings.apiTokens.logsEmpty") }}
            </p>
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="log in tokenLogs"
              :key="log.id"
              class="rounded-xl border border-(--md-sys-color-outline-variant) p-4"
            >
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                  <span
                    class="text-sm font-semibold text-(--md-sys-color-on-surface)"
                    >{{ getTokenLogActionLabel(log.action) }}</span
                  >
                  <span
                    class="text-xs px-1.5 py-0.5 rounded bg-(--md-sys-color-surface-variant) text-(--md-sys-color-on-surface-variant)"
                  >
                    {{ getTokenTypeLabel(log.token_type) }}
                  </span>
                </div>
                <span
                  class="text-xs text-(--md-sys-color-on-surface-variant)"
                  >{{ formatRelativeTime(log.created_at, locale) }}</span
                >
              </div>
              <div
                class="mt-2 flex flex-wrap gap-2 text-xs text-(--md-sys-color-on-surface-variant)"
              >
                <span
                  >{{ t("settings.apiTokens.logToken") }}:
                  {{ log.token_name }}</span
                >
                <span v-if="log.ip_address"
                  >{{ t("settings.apiTokens.logIP") }}:
                  {{ log.ip_address }}</span
                >
                <span v-if="log.image_hash" class="truncate"
                  >{{ t("settings.apiTokens.logImage") }}:
                  {{ log.image_hash }}</span
                >
              </div>
              <div
                class="mt-1 text-xs text-(--md-sys-color-on-surface-variant) opacity-70 truncate"
              >
                {{ log.method }} {{ log.path }}
              </div>
            </div>
            <div
              v-if="tokenLogsTotalPages > 1"
              class="mt-2 flex items-center justify-end gap-2 text-xs text-(--md-sys-color-on-surface-variant)"
            >
              <AnzuButton
                variant="text"
                :disabled="tokenLogsPage <= 1 || loadingTokenLogs"
                @click="() => handleTokenLogsPageChange(tokenLogsPage - 1)"
              >
                {{ t("common.actions.paginationPrevious") }}
              </AnzuButton>
              <span>{{ tokenLogsPage }} / {{ tokenLogsTotalPages }}</span>
              <AnzuButton
                variant="text"
                :disabled="
                  tokenLogsPage >= tokenLogsTotalPages || loadingTokenLogs
                "
                @click="() => handleTokenLogsPageChange(tokenLogsPage + 1)"
              >
                {{ t("common.actions.paginationNext") }}
              </AnzuButton>
            </div>
          </div>
        </div>
      </template>

      <template #tab-content-1>
        <!-- Security Logs -->
        <div>
          <div class="mb-4 flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold">
                {{ t("settings.securityLogs.title") }}
              </h3>
              <p class="mt-1 text-(--md-sys-color-on-surface-variant)">
                {{ t("settings.securityLogs.description") }}
              </p>
            </div>
          </div>
          
          <div class="mb-4 grid grid-cols-1 gap-4 sm:grid-cols-4">
            <AnzuInput
              v-model="securityLogFilters.search"
              :placeholder="t('common.actions.search')"
              @keyup.enter="loadSecurityLogs(1)"
            />
            <AnzuComboBox
              v-model="securityLogFilters.type"
              :items="securityLogTypeOptions"
              :placeholder="t('settings.apiTokens.tokenTypePlaceholder')"
              @update:modelValue="loadSecurityLogs(1)"
            />
            <AnzuInput
              v-model="securityLogFilters.startDate"
              type="date"
              @change="loadSecurityLogs(1)"
            />
            <AnzuInput
              v-model="securityLogFilters.endDate"
              type="date"
              @change="loadSecurityLogs(1)"
            />
          </div>
          <div class="mb-2 flex justify-end">
            <AnzuButton variant="text" @click="() => loadSecurityLogs(1)">
              {{ t("settings.securityLogs.refresh") }}
            </AnzuButton>
          </div>

          <div v-if="loadingSecurityLogs" class="flex justify-center py-6">
            <AnzuProgressRing :size="40" />
          </div>
          <div
            v-else-if="securityLogs.length === 0"
            class="rounded-lg border border-(--md-sys-color-outline-variant) p-6 text-center"
          >
            <p class="text-(--md-sys-color-on-surface-variant)">
              {{ t("settings.securityLogs.empty") }}
            </p>
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="log in securityLogs"
              :key="log.id"
              class="rounded-xl border border-(--md-sys-color-outline-variant) p-4"
            >
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                  <span
                    class="text-sm font-semibold text-(--md-sys-color-on-surface)"
                    >{{ getSecurityLogActionLabel(log.action) }}</span
                  >
                  <span
                    class="text-xs px-1.5 py-0.5 rounded"
                    :class="
                      log.level === 'warning' || log.level === 'error'
                        ? 'bg-(--md-sys-color-error-container) text-(--md-sys-color-on-error-container)'
                        : 'bg-(--md-sys-color-secondary-container) text-(--md-sys-color-on-secondary-container)'
                    "
                  >
                    {{ getSecurityLogLevelLabel(log.level) }}
                  </span>
                </div>
                <span
                  class="text-xs text-(--md-sys-color-on-surface-variant)"
                  >{{ formatRelativeTime(log.created_at, locale) }}</span
                >
              </div>
              <div
                class="mt-1 flex flex-wrap gap-2 text-xs text-(--md-sys-color-on-surface-variant) opacity-80"
              >
                <span v-if="log.ip_address"
                  >{{ t("settings.securityLogs.ip") }}:
                  {{ log.ip_address }}</span
                >
              </div>
              <div
                v-if="log.method || log.path"
                class="mt-1 text-xs text-(--md-sys-color-on-surface-variant) opacity-70 truncate"
              >
                {{ log.method }} {{ log.path }}
              </div>
            </div>
            <div
              v-if="securityLogsTotalPages > 1"
              class="mt-2 flex items-center justify-end gap-2 text-xs text-(--md-sys-color-on-surface-variant)"
            >
              <AnzuButton
                variant="text"
                :disabled="securityLogsPage <= 1 || loadingSecurityLogs"
                @click="
                  () => handleSecurityLogsPageChange(securityLogsPage - 1)
                "
              >
                {{ t("common.actions.paginationPrevious") }}
              </AnzuButton>
              <span>{{ securityLogsPage }} / {{ securityLogsTotalPages }}</span>
              <AnzuButton
                variant="text"
                :disabled="
                  securityLogsPage >= securityLogsTotalPages ||
                  loadingSecurityLogs
                "
                @click="
                  () => handleSecurityLogsPageChange(securityLogsPage + 1)
                "
              >
                {{ t("common.actions.paginationNext") }}
              </AnzuButton>
            </div>
          </div>
        </div>
      </template>
    </AnzuTabs>
  </div>

  <!-- Create Token Dialog -->
  <AnzuDialog
    v-model:visible="showCreateTokenDialog"
    :title="t('settings.apiTokens.createNew')"
    :actions="[
      {
        text: t('common.actions.cancel'),
        variant: 'text',
        handler: () => {
          showCreateTokenDialog = false;
        },
      },
      {
        text: t('settings.apiTokens.createNew'),
        primary: true,
        variant: 'filled',
        handler: handleCreateToken,
        loading: creatingToken,
      },
    ]"
  >
    <div class="flex flex-col gap-4">
      <div class="flex flex-col gap-2">
        <span class="text-sm font-medium text-(--md-sys-color-on-surface)">{{
          t("settings.apiTokens.tokenType")
        }}</span>
        <AnzuComboBox
          v-model="tokenForm.tokenType"
          :items="tokenTypeOptions"
          :aria-label="t('settings.apiTokens.tokenType')"
          :placeholder="t('settings.apiTokens.tokenTypePlaceholder')"
        />
      </div>
      <AnzuInput
        v-model="tokenForm.name"
        :label="t('settings.apiTokens.name')"
        :placeholder="t('settings.apiTokens.namePlaceholder')"
        name="token-name"
        autocomplete="off"
      />
      <AnzuTags
        v-model="tokenForm.ipAllowlist"
        :label="t('settings.apiTokens.ipAllowlist')"
        :max-tags="10"
        :hint="t('settings.apiTokens.ipAllowlistTip')"
      />
    </div>
  </AnzuDialog>

  <!-- Show Token Dialog -->
  <AnzuDialog
    v-model:visible="showTokenResultDialog"
    :title="t('settings.apiTokens.tokenCreatedTitle')"
    :actions="[
      {
        text: t('common.actions.close'),
        variant: 'filled',
        handler: () => {
          showTokenResultDialog = false;
        },
      },
    ]"
  >
    <div class="flex flex-col gap-4">
      <p class="text-(--md-sys-color-on-surface-variant)">
        {{ t("settings.apiTokens.tokenCreatedMessage") }}
      </p>
      <div class="relative">
        <AnzuInput
          :model-value="createdTokenRaw"
          readonly
          name="api-token"
          autocomplete="off"
        />
        <AnzuButton
          @click="copyToken"
          variant="tonal"
          class="absolute right-1 top-1 bottom-1"
        >
          {{ t("settings.apiTokens.copy") }}
        </AnzuButton>
      </div>
    </div>
  </AnzuDialog>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import Dashboard from "~/components/Dashboard.vue";
import AnzuButton from "~/components/AnzuButton.vue";
import AnzuComboBox from "~/components/AnzuComboBox.vue";
import AnzuInput from "~/components/AnzuInput.vue";
import AnzuTags from "~/components/AnzuTags.vue";
import AnzuDialog from "~/components/AnzuDialog.vue";
import AnzuTabs from "~/components/AnzuTabs.vue";
import AnzuProgressRing from "~/components/AnzuProgressRing.vue";
import { useAuth } from "~/composables/useAuth";
import { useNotification } from "~/composables/useNotification";
import { useDialog, isDialogDismissedError } from "~/composables/useDialog";
import { parseApiError } from "~/utils/api-error";
import { NotificationType } from "~/types/notification";
import { DialogVariant } from "~/types/dialog";
import { formatDate, formatRelativeTime } from "~/utils/format";
import { validatePassword } from "~/utils/password";
import { TrashIcon } from "@heroicons/vue/24/outline";
import type { APIToken, APITokenLog } from "~/types/api_token";
import type { PasskeyCredential } from "~/types/passkey";
import type { SecurityLog } from "~/types/security_log";

const { t, locale } = useI18n();
const {
  changePassword,
  logout,
  listPasskeys,
  deletePasskey,
  registerPasskey,
  checkPasskeyExists,
  createAPIToken,
  listAPITokens,
  deleteAPIToken,
  listAPITokenLogs,
  cleanupAPITokenLogs,
  listSecurityLogs,
  getLastApiErrorDisplay,
} = useAuth();
const { notify } = useNotification();
const { confirm } = useDialog();

const passwordForm = ref({
  currentPassword: "",
  newPassword: "",
  confirmPassword: "",
});
const changingPassword = ref(false);

// PassKey管理
const passkeys = ref<PasskeyCredential[]>([]);
const loadingPasskeys = ref(false);
const registeringPasskey = ref(false);
const deletingPasskeyId = ref<string | null>(null);

// API Tokens state
const apiTokens = ref<APIToken[]>([]);
const loadingTokens = ref(false);
const showCreateTokenDialog = ref(false);
const creatingToken = ref(false);
const tokenForm = ref({
  name: "",
  ipAllowlist: [] as string[],
  tokenType: "full",
});
const showTokenResultDialog = ref(false);
const createdTokenRaw = ref("");
const tokenLogs = ref<APITokenLog[]>([]);
const loadingTokenLogs = ref(false);
const tokenLogsPage = ref(1);
const tokenLogsPageSize = 20;
const tokenLogsTotal = ref(0);
const tokenLogsTotalPages = computed(() =>
  Math.max(1, Math.ceil(tokenLogsTotal.value / tokenLogsPageSize)),
);
const logRetentionDays = ref(30);
const cleaningTokenLogs = ref(false);
const securityLogs = ref<SecurityLog[]>([]);
const loadingSecurityLogs = ref(false);
const activeLogTab = ref<string | number>("token");
const logTabs = computed(() => [
  { label: t("settings.apiTokens.logsTitle"), value: "token" },
  { label: t("settings.securityLogs.title"), value: "security" },
]);
const securityLogsPage = ref(1);
const securityLogsPageSize = 20;
const securityLogsTotal = ref(0);
const securityLogsTotalPages = computed(() =>
  Math.max(1, Math.ceil(securityLogsTotal.value / securityLogsPageSize)),
);

const tokenTypeOptions = computed(() => [
  { value: "full", label: t("settings.apiTokens.tokenTypes.full") },
  { value: "upload", label: t("settings.apiTokens.tokenTypes.upload") },
  { value: "list", label: t("settings.apiTokens.tokenTypes.list") },
]);

const loadTokens = async () => {
  loadingTokens.value = true;
  apiTokens.value = await listAPITokens();
  loadingTokens.value = false;
};

const handleCreateToken = async () => {
  if (!tokenForm.value.name) return;
  creatingToken.value = true;
  const res = await createAPIToken(
    tokenForm.value.name,
    tokenForm.value.ipAllowlist,
    tokenForm.value.tokenType,
  );
  creatingToken.value = false;
  if (res) {
    createdTokenRaw.value = res.raw_token;
    showCreateTokenDialog.value = false;
    showTokenResultDialog.value = true;
    loadTokens();
    loadTokenLogs();
    tokenForm.value = { name: "", ipAllowlist: [], tokenType: "full" };
    notify({
      message: t("settings.apiTokens.createSuccess"),
      type: NotificationType.SUCCESS,
    });
  } else {
    notify({
      message: getLastApiErrorDisplay(t("settings.apiTokens.createFailed")),
      type: NotificationType.ERROR,
    });
  }
};

const handleDeleteToken = async (id: number) => {
  const result = await confirm(t("common.actions.deleteConfirm"), {
    title: t("common.actions.delete"),
    variant: DialogVariant.DESTRUCTIVE,
    actions: [
      { text: t("common.actions.cancel"), variant: "text" },
      { text: t("common.actions.delete"), primary: true, variant: "filled" },
    ],
  });
  if (!result) return;

  if (await deleteAPIToken(id)) {
    loadTokens();
    loadTokenLogs();
    notify({
      message: t("common.actions.deleteSuccess"),
      type: NotificationType.SUCCESS,
    });
  } else {
    notify({
      message: getLastApiErrorDisplay(t("common.actions.deleteFailed")),
      type: NotificationType.ERROR,
    });
  }
};

const copyToken = () => {
  navigator.clipboard.writeText(createdTokenRaw.value);
  notify({
    message: t("settings.apiTokens.copySuccess"),
    type: NotificationType.SUCCESS,
  });
};

const tokenLogFilters = ref({
  search: "",
  startDate: "",
  endDate: "",
  type: "",
});

const securityLogFilters = ref({
  search: "",
  startDate: "",
  endDate: "",
  type: "",
});

const tokenLogTypeOptions = computed(() => [
  { value: "", label: t("common.labels.all") },
  { value: "token_create", label: t("settings.apiTokens.logActions.tokenCreate") },
  { value: "token_delete", label: t("settings.apiTokens.logActions.tokenDelete") },
  { value: "image_upload", label: t("settings.apiTokens.logActions.imageUpload") },
  { value: "image_list", label: t("settings.apiTokens.logActions.imageList") },
]);

const securityLogTypeOptions = computed(() => [
  { value: "", label: t("common.labels.all") },
  { value: "login_success", label: t("settings.securityLogs.actions.loginSuccess") },
  { value: "login_failed", label: t("settings.securityLogs.actions.loginFailed") },
  { value: "logout", label: t("settings.securityLogs.actions.logout") },
  { value: "password_changed", label: t("settings.securityLogs.actions.passwordChanged") },
  // Add more as needed
]);

const loadTokenLogs = async (page = tokenLogsPage.value) => {
  loadingTokenLogs.value = true;
  const res = await listAPITokenLogs(
    page,
    tokenLogsPageSize,
    tokenLogFilters.value.search,
    tokenLogFilters.value.startDate,
    tokenLogFilters.value.endDate,
    tokenLogFilters.value.type
  );
  tokenLogs.value = res.data;
  tokenLogsPage.value = res.page;
  tokenLogsTotal.value = res.total;
  loadingTokenLogs.value = false;
};


const handleTokenLogsPageChange = async (page: number) => {
  if (
    page < 1 ||
    page > tokenLogsTotalPages.value ||
    page === tokenLogsPage.value
  )
    return;
  await loadTokenLogs(page);
};

const handleCleanupTokenLogs = async () => {
  if (!logRetentionDays.value || logRetentionDays.value <= 0) {
    notify({
      message: t("settings.apiTokens.cleanupInvalidDays"),
      type: NotificationType.WARNING,
    });
    return;
  }

  const result = await confirm(
    t("settings.apiTokens.cleanupConfirm", { days: logRetentionDays.value }),
    {
      title: t("settings.apiTokens.cleanupTitle"),
      variant: DialogVariant.DESTRUCTIVE,
      actions: [
        { text: t("common.actions.cancel"), variant: "text" },
        { text: t("common.actions.confirm"), primary: true, variant: "filled" },
      ],
    },
  );
  if (!result) return;

  cleaningTokenLogs.value = true;
  const res = await cleanupAPITokenLogs(logRetentionDays.value);
  cleaningTokenLogs.value = false;

  if (res) {
    notify({
      message: t("settings.apiTokens.cleanupSuccess", { count: res.deleted }),
      type: NotificationType.SUCCESS,
    });
    loadTokenLogs();
  } else {
    notify({
      message: getLastApiErrorDisplay(t("settings.apiTokens.cleanupFailed")),
      type: NotificationType.ERROR,
    });
  }
};

const getTokenTypeLabel = (type: string) => {
  switch (type) {
    case "upload":
      return t("settings.apiTokens.tokenTypes.upload");
    case "list":
      return t("settings.apiTokens.tokenTypes.list");
    default:
      return t("settings.apiTokens.tokenTypes.full");
  }
};

const getTokenLogActionLabel = (action: string) => {
  switch (action) {
    case "token_create":
      return t("settings.apiTokens.logActions.tokenCreate");
    case "token_delete":
      return t("settings.apiTokens.logActions.tokenDelete");
    case "image_upload":
      return t("settings.apiTokens.logActions.imageUpload");
    case "image_list":
      return t("settings.apiTokens.logActions.imageList");
    default:
      return action;
  }
};

const loadSecurityLogs = async (page = securityLogsPage.value) => {
  loadingSecurityLogs.value = true;
  const res = await listSecurityLogs(
    page,
    securityLogsPageSize,
    false,
    securityLogFilters.value.search,
    securityLogFilters.value.startDate,
    securityLogFilters.value.endDate,
    securityLogFilters.value.type
  );
  securityLogs.value = res.data;
  securityLogsPage.value = res.page;
  securityLogsTotal.value = res.total;
  loadingSecurityLogs.value = false;
};

const handleSecurityLogsPageChange = async (page: number) => {
  if (
    page < 1 ||
    page > securityLogsTotalPages.value ||
    page === securityLogsPage.value
  )
    return;
  await loadSecurityLogs(page);
};

const getSecurityLogActionLabel = (action: string) => {
  switch (action) {
    case "login_failed":
      return t("settings.securityLogs.actions.loginFailed");
    case "login_success":
      return t("settings.securityLogs.actions.loginSuccess");
    case "login_rate_limited":
      return t("settings.securityLogs.actions.loginRateLimited");
    case "login_bruteforce_alert":
      return t("settings.securityLogs.actions.loginBruteforceAlert");
    case "logout":
      return t("settings.securityLogs.actions.logout");
    case "password_changed":
      return t("settings.securityLogs.actions.passwordChanged");
    case "password_change_failed":
      return t("settings.securityLogs.actions.passwordChangeFailed");
    case "passkey_register_success":
      return t("settings.securityLogs.actions.passkeyRegisterSuccess");
    case "passkey_register_failed":
      return t("settings.securityLogs.actions.passkeyRegisterFailed");
    case "passkey_login_success":
      return t("settings.securityLogs.actions.passkeyLoginSuccess");
    case "passkey_login_failed":
      return t("settings.securityLogs.actions.passkeyLoginFailed");
    case "passkey_delete_success":
      return t("settings.securityLogs.actions.passkeyDeleteSuccess");
    case "passkey_delete_failed":
      return t("settings.securityLogs.actions.passkeyDeleteFailed");
    case "token_create_success":
      return t("settings.securityLogs.actions.tokenCreateSuccess");
    case "token_create_failed":
      return t("settings.securityLogs.actions.tokenCreateFailed");
    case "token_delete_success":
      return t("settings.securityLogs.actions.tokenDeleteSuccess");
    case "token_delete_failed":
      return t("settings.securityLogs.actions.tokenDeleteFailed");
    case "token_logs_cleanup":
      return t("settings.securityLogs.actions.tokenLogsCleanup");
    case "token_logs_cleanup_failed":
      return t("settings.securityLogs.actions.tokenLogsCleanupFailed");
    default:
      return action;
  }
};

const getSecurityLogLevelLabel = (level: string) => {
  switch (level) {
    case "warning":
      return t("settings.securityLogs.levels.warning");
    case "error":
      return t("settings.securityLogs.levels.error");
    default:
      return t("settings.securityLogs.levels.info");
  }
};

// 修改密码
const handleChangePassword = async () => {
  if (
    !passwordForm.value.currentPassword ||
    !passwordForm.value.newPassword ||
    !passwordForm.value.confirmPassword
  ) {
    notify({
      message: t("settings.changePassword.fillAllFields"),
      type: NotificationType.WARNING,
    });
    return;
  }

  const validation = validatePassword(passwordForm.value.newPassword, t);
  if (!validation.valid) {
    notify({
      message: validation.error!,
      type: NotificationType.WARNING,
    });
    return;
  }

  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    notify({
      message: t("settings.changePassword.passwordMatchError"),
      type: NotificationType.WARNING,
    });
    return;
  }

  changingPassword.value = true;

  const success = await changePassword(
    passwordForm.value.currentPassword,
    passwordForm.value.newPassword,
  );
  if (success) {
    notify({
      message: t("settings.changePassword.success"),
      type: NotificationType.SUCCESS,
    });
    // 修改密码成功后自动登出
    setTimeout(() => {
      logout();
    }, 1500);
  } else {
    notify({
      message: getLastApiErrorDisplay(t("settings.changePassword.failed")),
      type: NotificationType.ERROR,
    });
  }

  changingPassword.value = false;
};

// 注册新PassKey
const handleRegisterPasskey = async () => {
  registeringPasskey.value = true;

  const success = await registerPasskey();
  if (success) {
    notify({
      message: t("settings.passkeyManagement.registerSuccess"),
      type: NotificationType.SUCCESS,
    });
    loadPasskeys();
  } else {
    notify({
      message: getLastApiErrorDisplay(
        t("settings.passkeyManagement.registerFailed"),
      ),
      type: NotificationType.ERROR,
    });
  }

  registeringPasskey.value = false;
};

// 删除PassKey
const handleDeletePasskey = async (credentialId: string) => {
  try {
    const result = await confirm(t("common.actions.deleteConfirm"), {
      title: t("common.actions.delete"),
      variant: DialogVariant.DESTRUCTIVE,
      actions: [
        { text: t("common.actions.cancel"), variant: "text" },
        {
          text: t("common.actions.delete"),
          primary: true,
          variant: "filled",
          loading: false,
        },
      ],
    });

    if (!result) return;

    deletingPasskeyId.value = credentialId;

    const success = await deletePasskey(credentialId);
    if (success) {
      notify({
        message: t("common.actions.deleteSuccess"),
        type: NotificationType.SUCCESS,
      });
      loadPasskeys();
    } else {
      notify({
        message: getLastApiErrorDisplay(t("common.actions.deleteFailed")),
        type: NotificationType.ERROR,
      });
    }
  } catch (e: any) {
    if (isDialogDismissedError(e)) {
      return;
    }
    const parsed = parseApiError(e, t("common.actions.deleteFailed"));
    notify({
      message: parsed.displayMessage,
      type: NotificationType.ERROR,
    });
  } finally {
    deletingPasskeyId.value = null;
  }
};

// 加载PassKey列表
const loadPasskeys = async () => {
  loadingPasskeys.value = true;
  passkeys.value = await listPasskeys();
  loadingPasskeys.value = false;
};

const getDeviceIcon = (ua: string = "") => {
  ua = ua.toLowerCase();
  if (
    ua.includes("mobile") ||
    ua.includes("android") ||
    ua.includes("iphone") ||
    ua.includes("ipad")
  ) {
    return "M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3";
  }
  return "M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25";
};

onMounted(() => {
  loadPasskeys();
  loadTokens();
  loadTokenLogs();
  loadSecurityLogs();
});
</script>
