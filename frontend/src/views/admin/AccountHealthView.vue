<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="relative overflow-hidden rounded-[28px] border border-slate-200/80 bg-white shadow-[0_20px_70px_-45px_rgba(15,23,42,0.45)] dark:border-slate-700 dark:bg-slate-900">
        <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(16,185,129,0.16),_transparent_36%),radial-gradient(circle_at_bottom_right,_rgba(59,130,246,0.14),_transparent_34%)]"></div>
        <div class="relative space-y-7 p-6 xl:p-8">
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div>
              <p class="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">{{ t('admin.accountHealth.title') }}</p>
              <h2 class="mt-3 text-3xl font-semibold tracking-tight text-slate-900 dark:text-white">{{ t('admin.accountHealth.description') }}</h2>
              <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-500 dark:text-slate-400">
                {{ activeTab === 'health' ? healthStatusText : tokenStatusText }}
              </p>
            </div>

            <div v-if="activeTab === 'health'" class="flex flex-wrap items-center gap-2">
              <button class="btn btn-danger" :disabled="deletingUnhealthy || healthChecking" @click="deleteUnhealthyAccountsInScope">
                {{ deletingUnhealthy ? t('admin.accounts.deleteUnhealthyRunning') : t('admin.accounts.deleteUnhealthy') }}
              </button>
              <button class="btn btn-secondary" :disabled="healthChecking" @click="runGlobalHealthCheck">
                {{ healthChecking ? t('admin.accounts.healthCheckRunning') : t('admin.accounts.healthCheckAll') }}
              </button>
            </div>
          </div>

          <div class="inline-flex w-fit rounded-2xl border border-slate-200 bg-slate-50/90 p-1 shadow-sm dark:border-slate-700 dark:bg-slate-800/80">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              type="button"
              class="rounded-xl px-4 py-2 text-sm font-medium transition"
              :class="activeTab === tab.key ? 'bg-white text-slate-900 shadow-sm dark:bg-slate-900 dark:text-white' : 'text-slate-500 hover:text-slate-800 dark:text-slate-300 dark:hover:text-white'"
              @click="activeTab = tab.key"
            >
              {{ tab.label }}
            </button>
          </div>

          <AccountHealthAutoCheckPanel
            v-if="activeTab === 'health'"
            :health-summary="healthSummary"
            :auto-config="autoConfig"
            :manual-model-id="manualModelId"
            :auto-interval-input="autoIntervalInput"
            :auto-last-run-text="autoLastRunText"
            :health-checking="healthChecking"
            :saving-auto-config="savingAutoConfig"
            :deleting-unhealthy="deletingUnhealthy"
            @update:manual-model-id="manualModelId = $event"
            @update:auto-interval-input="autoIntervalInput = $event"
            @run-health-check="runGlobalHealthCheck"
            @save-config="saveAutoConfig"
            @delete-unhealthy="deleteUnhealthyAccountsInScope"
          />

          <AccountTokenAutoRefreshPanel
            v-else
            :token-config="tokenConfig"
            :token-interval-value-input="tokenIntervalValueInput"
            :token-batch-size-input="tokenBatchSizeInput"
            :token-last-run-text="tokenLastRunText"
            :saving-token-config="savingTokenConfig"
            :running-token-refresh="runningTokenRefresh"
            :groups="groups"
            @update:token-interval-value-input="tokenIntervalValueInput = $event"
            @update:token-batch-size-input="tokenBatchSizeInput = $event"
            @save-config="saveTokenConfig"
            @run-now="runTokenRefreshNow"
          />
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AccountHealthAutoCheckConfig, AccountHealthSummary, AccountTokenAutoRefreshConfig } from '@/api/admin/accounts'
import type { AdminGroup } from '@/types'
import { adminAPI } from '@/api/admin'
import AccountHealthAutoCheckPanel from '@/components/admin/account-health/AccountHealthAutoCheckPanel.vue'
import AccountTokenAutoRefreshPanel from '@/components/admin/account-health/AccountTokenAutoRefreshPanel.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'
import { formatRelativeTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const AUTO_POLL_MS = 15000

const activeTab = ref<'health' | 'token'>('health')
const healthChecking = ref(false)
const savingAutoConfig = ref(false)
const savingTokenConfig = ref(false)
const runningTokenRefresh = ref(false)
const deletingUnhealthy = ref(false)
const polling = ref(false)
const groups = ref<AdminGroup[]>([])
const lastObservedAutoRunAt = ref<number | null>(null)
const lastObservedTokenRunAt = ref<number | null>(null)
const manualModelId = ref('')
const autoIntervalInput = ref('60')
const tokenIntervalValueInput = ref('1')
const tokenBatchSizeInput = ref('10')

const autoConfig = reactive<AccountHealthAutoCheckConfig>({
  enabled: false,
  interval_minutes: 60,
  model_id: '',
  running: false,
  current_total: 0,
  current_success: 0,
  current_failed: 0,
  last_run_at: null
})

const tokenConfig = reactive<AccountTokenAutoRefreshConfig>({
  enabled: false,
  interval_value: 1,
  interval_unit: 'day',
  batch_size: 10,
  scope: 'all',
  group_id: 0,
  running: false,
  current_total: 0,
  current_success: 0,
  current_failed: 0,
  last_run_at: null,
  last_run_total: 0,
  last_run_success: 0,
  last_run_failed: 0
})

const healthSummary = ref<AccountHealthSummary>({
  total_accounts: 0,
  healthy_accounts: 0,
  constrained_accounts: 0,
  unavailable_accounts: 0,
  unchecked_accounts: 0,
  last_checked_at: ''
})

const tabs = computed(() => [
  { key: 'health' as const, label: t('admin.accounts.autoCheck') },
  { key: 'token' as const, label: t('admin.accounts.tokenRefresh.tab') }
])

const autoLastRunText = computed(() => formatLastRun(autoConfig.last_run_at))
const tokenLastRunText = computed(() => formatLastRun(tokenConfig.last_run_at))
const tokenGroupName = computed(() => {
  const groupID = tokenConfig.group_id ?? 0
  return groups.value.find((group) => group.id === groupID)?.name ?? ''
})

const healthStatusText = computed(() => {
  if (autoConfig.running) {
    return t('admin.accounts.healthCheckProgress', {
      current: autoConfig.current_success ?? 0,
      total: autoConfig.current_total ?? 0,
      failed: autoConfig.current_failed ?? 0
    })
  }
  const base = autoConfig.enabled ? t('admin.accounts.autoCheckEnabled') : t('admin.accounts.healthSummary.neverChecked')
  return autoConfig.last_run_at
    ? `${base} · ${t('admin.accounts.healthSummary.lastChecked', { time: autoLastRunText.value })}`
    : base
})

const tokenStatusText = computed(() => {
  const base = tokenConfig.enabled ? t('admin.accounts.tokenRefresh.enabled') : t('admin.accounts.tokenRefresh.disabledHint')
  const scope = tokenConfig.scope === 'group'
    ? `${t('admin.accounts.tokenRefresh.scopeGroup')} · ${tokenGroupName.value || t('admin.accounts.tokenRefresh.groupPlaceholder')}`
    : t('admin.accounts.tokenRefresh.scopeAll')
  const status = `${base} · ${scope}`
  return tokenConfig.last_run_at
    ? `${status} · ${t('admin.accounts.tokenRefresh.lastRunAt', { time: tokenLastRunText.value })}`
    : status
})

function formatLastRun(timestamp?: number | null) {
  if (!timestamp) {
    return t('admin.accounts.healthSummary.neverChecked')
  }
  return formatRelativeTime(new Date(timestamp * 1000).toISOString())
}

function applyAutoConfig(cfg: AccountHealthAutoCheckConfig) {
  autoConfig.enabled = cfg.enabled
  autoConfig.interval_minutes = cfg.interval_minutes || 60
  autoConfig.model_id = cfg.model_id || ''
  autoConfig.running = cfg.running ?? false
  autoConfig.current_total = cfg.current_total ?? 0
  autoConfig.current_success = cfg.current_success ?? 0
  autoConfig.current_failed = cfg.current_failed ?? 0
  autoConfig.last_run_at = cfg.last_run_at ?? null
  autoIntervalInput.value = String(autoConfig.interval_minutes)
  lastObservedAutoRunAt.value = autoConfig.last_run_at ?? null
}

function applyTokenConfig(cfg: AccountTokenAutoRefreshConfig) {
  tokenConfig.enabled = cfg.enabled
  tokenConfig.interval_value = cfg.interval_value || 1
  tokenConfig.interval_unit = cfg.interval_unit || 'day'
  tokenConfig.batch_size = cfg.batch_size || 10
  tokenConfig.scope = cfg.scope === 'group' ? 'group' : 'all'
  tokenConfig.group_id = tokenConfig.scope === 'group' ? (cfg.group_id ?? 0) : 0
  tokenConfig.running = cfg.running ?? false
  tokenConfig.current_total = cfg.current_total ?? 0
  tokenConfig.current_success = cfg.current_success ?? 0
  tokenConfig.current_failed = cfg.current_failed ?? 0
  tokenConfig.last_run_at = cfg.last_run_at ?? null
  tokenConfig.last_run_total = cfg.last_run_total ?? 0
  tokenConfig.last_run_success = cfg.last_run_success ?? 0
  tokenConfig.last_run_failed = cfg.last_run_failed ?? 0
  tokenIntervalValueInput.value = String(tokenConfig.interval_value)
  tokenBatchSizeInput.value = String(tokenConfig.batch_size)
  lastObservedTokenRunAt.value = tokenConfig.last_run_at ?? null
}

async function loadHealthSummary() {
  healthSummary.value = await adminAPI.accounts.getHealthSummary()
}

async function loadAutoConfig() {
  applyAutoConfig(await adminAPI.accounts.getAccountHealthAutoCheckConfig())
}

async function loadTokenConfig() {
  applyTokenConfig(await adminAPI.accounts.getAccountTokenAutoRefreshConfig())
}

async function loadGroups() {
  groups.value = await adminAPI.groups.getAll()
}

async function reloadPage() {
  await Promise.all([loadHealthSummary(), loadAutoConfig(), loadTokenConfig(), loadGroups()])
}

async function runGlobalHealthCheck() {
  if (healthChecking.value) return
  healthChecking.value = true
  try {
    const modelID = manualModelId.value.trim() || autoConfig.model_id.trim()
    const result = await adminAPI.accounts.runHealthCheck({ model_id: modelID || undefined })
    autoConfig.running = true
    autoConfig.current_total = result.total
    autoConfig.current_success = 0
    autoConfig.current_failed = 0
    appStore.showSuccess(
      result.started
        ? t('admin.accounts.healthCheckRunStarted')
        : t('admin.accounts.healthCheckAlreadyRunning')
    )
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.accounts.healthCheckFailed'))
  } finally {
    healthChecking.value = false
  }
}

async function saveAutoConfig() {
  if (savingAutoConfig.value) return

  const interval = Number(autoIntervalInput.value)
  if (!Number.isFinite(interval) || interval < 1) {
    appStore.showError(t('admin.accounts.autoCheckIntervalHint'))
    return
  }

  savingAutoConfig.value = true
  try {
    const updated = await adminAPI.accounts.updateAccountHealthAutoCheckConfig({
      enabled: autoConfig.enabled,
      interval_minutes: interval,
      model_id: autoConfig.model_id.trim()
    })
    applyAutoConfig(updated)
    appStore.showSuccess(t('admin.accounts.autoCheckSaved'))
  } catch (error: any) {
    appStore.showError(error?.message || t('common.error'))
  } finally {
    savingAutoConfig.value = false
  }
}

async function saveTokenConfig() {
  if (savingTokenConfig.value) return

  const intervalValue = Number(tokenIntervalValueInput.value)
  const batchSize = Number(tokenBatchSizeInput.value)
  if (!Number.isFinite(intervalValue) || intervalValue < 1) {
    appStore.showError(t('admin.accounts.tokenRefresh.intervalHint'))
    return
  }
  if (!Number.isFinite(batchSize) || batchSize < 1 || batchSize > 50) {
    appStore.showError(t('admin.accounts.tokenRefresh.batchHint'))
    return
  }
  if (tokenConfig.scope === 'group' && !(tokenConfig.group_id && tokenConfig.group_id > 0)) {
    appStore.showError(t('admin.accounts.tokenRefresh.groupRequired'))
    return
  }

  savingTokenConfig.value = true
  try {
    const updated = await adminAPI.accounts.updateAccountTokenAutoRefreshConfig({
      enabled: tokenConfig.enabled,
      interval_value: intervalValue,
      interval_unit: tokenConfig.interval_unit,
      batch_size: batchSize,
      scope: tokenConfig.scope,
      group_id: tokenConfig.scope === 'group' ? tokenConfig.group_id : 0
    })
    applyTokenConfig(updated)
    appStore.showSuccess(t('admin.accounts.tokenRefresh.saved'))
  } catch (error: any) {
    appStore.showError(error?.message || t('common.error'))
  } finally {
    savingTokenConfig.value = false
  }
}

async function runTokenRefreshNow() {
  if (runningTokenRefresh.value) return
  runningTokenRefresh.value = true
  try {
    const result = await adminAPI.accounts.runAccountTokenAutoRefreshNow()
    tokenConfig.running = true
    tokenConfig.current_total = 0
    tokenConfig.current_success = 0
    tokenConfig.current_failed = 0
    tokenConfig.batch_size = result.batch_size
    tokenBatchSizeInput.value = String(result.batch_size)
    appStore.showSuccess(
      result.started
        ? t('admin.accounts.tokenRefresh.runStarted')
        : t('admin.accounts.tokenRefresh.runAlreadyRunning')
    )
  } catch (error: any) {
    appStore.showError(error?.message || t('common.error'))
  } finally {
    runningTokenRefresh.value = false
  }
}

async function deleteUnhealthyAccountsInScope() {
  if (deletingUnhealthy.value) return
  if (!confirm(t('admin.accounts.deleteUnhealthyConfirm'))) return

  deletingUnhealthy.value = true
  try {
    const result = await adminAPI.accounts.deleteUnhealthyAccounts()
    await reloadPage()
    appStore.showSuccess(t('admin.accounts.deleteUnhealthyDone', { count: result.deleted_count }))
  } catch (error: any) {
    appStore.showError(error?.message || t('common.error'))
  } finally {
    deletingUnhealthy.value = false
  }
}

async function pollUpdates() {
  if (polling.value || healthChecking.value || deletingUnhealthy.value) return
  if (typeof document !== 'undefined' && document.hidden) return

  polling.value = true
  try {
    const [healthCfg, refreshCfg] = await Promise.all([
      adminAPI.accounts.getAccountHealthAutoCheckConfig(),
      adminAPI.accounts.getAccountTokenAutoRefreshConfig()
    ])

    const nextHealthRunAt = healthCfg.last_run_at ?? null
    const hasNewHealthRun = nextHealthRunAt !== null && lastObservedAutoRunAt.value !== null && nextHealthRunAt !== lastObservedAutoRunAt.value

    applyAutoConfig(healthCfg)
    applyTokenConfig(refreshCfg)

    if (hasNewHealthRun || (!tokenConfig.running && !autoConfig.running)) {
      await loadHealthSummary()
    }
  } catch (error) {
    console.error('Failed to poll account health page state:', error)
  } finally {
    polling.value = false
  }
}

let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  try {
    await reloadPage()
    pollTimer = setInterval(() => {
      void pollUpdates()
    }, AUTO_POLL_MS)
  } catch (error) {
    console.error('Failed to initialize account health page:', error)
  }
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})
</script>
