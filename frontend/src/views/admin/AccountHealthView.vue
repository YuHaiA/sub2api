<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="space-y-6">
          <section class="relative overflow-hidden rounded-[28px] border border-slate-200/80 bg-white shadow-[0_20px_70px_-45px_rgba(15,23,42,0.45)] dark:border-slate-700 dark:bg-slate-900">
            <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(16,185,129,0.16),_transparent_36%),radial-gradient(circle_at_bottom_right,_rgba(59,130,246,0.14),_transparent_34%)]"></div>
            <div class="relative grid gap-6 p-6 xl:grid-cols-[1.25fr,0.95fr] xl:p-8">
              <div class="space-y-6">
                <div class="flex flex-wrap items-start justify-between gap-4">
                  <div>
                    <p class="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-500">
                      {{ t('admin.accountHealth.title') }}
                    </p>
                    <h2 class="mt-3 text-3xl font-semibold tracking-tight text-slate-900 dark:text-white">
                      {{ t('admin.accountHealth.description') }}
                    </h2>
                    <p class="mt-3 max-w-2xl text-sm leading-6 text-slate-500 dark:text-slate-400">
                      {{ autoConfig.enabled ? t('admin.accounts.autoCheckEnabled') : t('admin.accounts.healthSummary.neverChecked') }}
                      <span v-if="autoConfig.last_run_at">
                        · {{ t('admin.accounts.healthSummary.lastChecked', { time: autoLastRunText }) }}
                      </span>
                    </p>
                  </div>

                  <div class="flex flex-wrap items-center gap-2">
                    <button class="btn btn-danger" :disabled="deletingUnhealthy || healthChecking" @click="deleteUnhealthyAccountsInScope">
                      {{ deletingUnhealthy ? t('admin.accounts.deleteUnhealthyRunning') : t('admin.accounts.deleteUnhealthy') }}
                    </button>
                    <button class="btn btn-secondary" :disabled="healthChecking" @click="runGlobalHealthCheck">
                      {{ healthChecking ? t('admin.accounts.healthCheckRunning') : t('admin.accounts.healthCheckAll') }}
                    </button>
                  </div>
                </div>

                <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                  <div class="rounded-2xl border border-slate-200 bg-white/80 p-4 backdrop-blur dark:border-slate-700 dark:bg-slate-800/70">
                    <p class="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">{{ t('admin.accounts.healthSummary.total') }}</p>
                    <p class="mt-3 text-3xl font-semibold text-slate-900 dark:text-white">{{ healthSummary.total_accounts }}</p>
                  </div>
                  <div class="rounded-2xl border border-emerald-200 bg-emerald-50/90 p-4 dark:border-emerald-900/40 dark:bg-emerald-900/10">
                    <p class="text-xs font-medium uppercase tracking-wide text-emerald-700 dark:text-emerald-300">{{ t('admin.accounts.healthSummary.healthy') }}</p>
                    <p class="mt-3 text-3xl font-semibold text-emerald-700 dark:text-emerald-200">{{ healthSummary.healthy_accounts }}</p>
                    <p class="mt-2 text-xs text-emerald-600/80 dark:text-emerald-300/80">{{ t('admin.accounts.healthSummary.healthyHint') }}</p>
                  </div>
                  <div class="rounded-2xl border border-amber-200 bg-amber-50/90 p-4 dark:border-amber-900/40 dark:bg-amber-900/10">
                    <p class="text-xs font-medium uppercase tracking-wide text-amber-700 dark:text-amber-300">{{ t('admin.accounts.healthSummary.bannedOrExhausted') }}</p>
                    <p class="mt-3 text-3xl font-semibold text-amber-700 dark:text-amber-200">{{ healthSummary.banned_or_exhausted_accounts }}</p>
                  </div>
                  <div class="rounded-2xl border border-rose-200 bg-rose-50/90 p-4 dark:border-rose-900/40 dark:bg-rose-900/10">
                    <p class="text-xs font-medium uppercase tracking-wide text-rose-700 dark:text-rose-300">{{ t('admin.accounts.healthSummary.unavailable') }}</p>
                    <p class="mt-3 text-3xl font-semibold text-rose-700 dark:text-rose-200">{{ healthSummary.unavailable_accounts }}</p>
                    <p class="mt-2 text-xs text-rose-600/80 dark:text-rose-300/80">
                      {{ t('admin.accounts.healthSummary.unchecked', { count: healthSummary.unchecked_accounts }) }}
                    </p>
                  </div>
                </div>
              </div>

              <div class="space-y-4 rounded-[24px] border border-slate-200 bg-white/85 p-5 backdrop-blur dark:border-slate-700 dark:bg-slate-800/75">
                <div>
                  <h3 class="text-base font-semibold text-slate-900 dark:text-white">{{ t('admin.accounts.autoCheck') }}</h3>
                  <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('admin.accounts.autoCheckIntervalHint') }}</p>
                </div>

                <label class="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-700 dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-200">
                  <input
                    v-model="autoConfig.enabled"
                    type="checkbox"
                    class="h-4 w-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                  />
                  {{ t('admin.accounts.autoCheckEnabled') }}
                </label>

                <div class="grid gap-4 sm:grid-cols-2">
                  <div>
                    <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
                      {{ t('admin.accounts.healthCheckModelPlaceholder') }}
                    </label>
                    <Input
                      v-model="manualModelId"
                      :disabled="healthChecking"
                      :placeholder="t('admin.accounts.healthCheckModelPlaceholder')"
                    />
                  </div>
                  <div>
                    <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
                      {{ t('admin.accounts.autoCheckInterval') }}
                    </label>
                    <Input
                      v-model="autoIntervalInput"
                      type="number"
                      :placeholder="'60'"
                    />
                  </div>
                </div>

                <div>
                  <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
                    {{ t('admin.accounts.autoCheckModel') }}
                  </label>
                  <Input
                    v-model="autoConfig.model_id"
                    :placeholder="t('admin.accounts.healthCheckModelPlaceholder')"
                  />
                </div>

                <div class="rounded-2xl bg-slate-50 px-4 py-3 text-sm text-slate-600 dark:bg-slate-900/70 dark:text-slate-300">
                  <div class="flex flex-wrap items-center justify-between gap-2">
                    <span>{{ t('admin.accounts.healthSummary.lastChecked', { time: autoLastRunText }) }}</span>
                    <span class="badge text-xs" :class="autoConfig.enabled ? 'badge-success' : 'badge-gray'">
                      {{ autoConfig.enabled ? t('common.enabled') : t('common.disabled') }}
                    </span>
                  </div>
                </div>

                <div class="flex justify-end">
                  <button class="btn btn-primary" :disabled="savingAutoConfig" @click="saveAutoConfig">
                    {{ savingAutoConfig ? t('common.saving') : t('admin.accounts.autoCheckSave') }}
                  </button>
                </div>
              </div>
            </div>
          </section>
        </div>
      </template>

      <template #table>
        <div class="rounded-[28px] border border-dashed border-slate-300 bg-white/70 px-6 py-10 text-center text-sm text-slate-500 shadow-[0_20px_70px_-45px_rgba(15,23,42,0.45)] dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-400">
          {{ t('admin.accountHealth.description') }}
        </div>
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Input from '@/components/common/Input.vue'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { formatRelativeTime } from '@/utils/format'
import type { AccountHealthAutoCheckConfig, AccountHealthSummary } from '@/api/admin/accounts'

const { t } = useI18n()
const appStore = useAppStore()

const healthChecking = ref(false)
const savingAutoConfig = ref(false)
const deletingUnhealthy = ref(false)
const autoHealthPolling = ref(false)

const AUTO_HEALTH_POLL_MS = 15000
const lastObservedAutoRunAt = ref<number | null>(null)

const manualModelId = ref('')
const autoIntervalInput = ref('60')

const autoConfig = reactive<AccountHealthAutoCheckConfig>({
  enabled: false,
  interval_minutes: 60,
  model_id: '',
  last_run_at: null
})

const healthSummary = ref<AccountHealthSummary>({
  total_accounts: 0,
  healthy_accounts: 0,
  banned_or_exhausted_accounts: 0,
  unavailable_accounts: 0,
  unchecked_accounts: 0,
  last_checked_at: ''
})

const autoLastRunText = computed(() => {
  if (!autoConfig.last_run_at) return t('admin.accounts.healthSummary.neverChecked')
  return formatRelativeTime(new Date(autoConfig.last_run_at * 1000).toISOString())
})

const loadHealthSummary = async () => {
  healthSummary.value = await adminAPI.accounts.getHealthSummary()
}

const loadAutoConfig = async () => {
  const cfg = await adminAPI.accounts.getAccountHealthAutoCheckConfig()
  autoConfig.enabled = cfg.enabled
  autoConfig.interval_minutes = cfg.interval_minutes || 60
  autoConfig.model_id = cfg.model_id || ''
  autoConfig.last_run_at = cfg.last_run_at ?? null
  autoIntervalInput.value = String(autoConfig.interval_minutes)
  lastObservedAutoRunAt.value = cfg.last_run_at ?? null
}

const reloadPage = async () => {
  await Promise.all([
    loadHealthSummary(),
    loadAutoConfig()
  ])
}

const runGlobalHealthCheck = async () => {
  if (healthChecking.value) return
  healthChecking.value = true
  try {
    const modelID = manualModelId.value.trim() || autoConfig.model_id.trim()
    await adminAPI.accounts.runHealthCheck({
      model_id: modelID || undefined
    })
    await reloadPage()
    appStore.showSuccess(t('admin.accounts.healthCheckCompleted', { count: healthSummary.value.total_accounts }))
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.accounts.healthCheckFailed'))
  } finally {
    healthChecking.value = false
  }
}

const saveAutoConfig = async () => {
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
    autoConfig.enabled = updated.enabled
    autoConfig.interval_minutes = updated.interval_minutes
    autoConfig.model_id = updated.model_id
    autoConfig.last_run_at = updated.last_run_at ?? null
    autoIntervalInput.value = String(updated.interval_minutes)
    lastObservedAutoRunAt.value = updated.last_run_at ?? null
    appStore.showSuccess(t('admin.accounts.autoCheckSaved'))
  } catch (error: any) {
    appStore.showError(error?.message || t('common.error'))
  } finally {
    savingAutoConfig.value = false
  }
}

const deleteUnhealthyAccountsInScope = async () => {
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

const pollAutoHealthUpdates = async () => {
  if (autoHealthPolling.value || healthChecking.value || deletingUnhealthy.value) return
  if (typeof document !== 'undefined' && document.hidden) return
  autoHealthPolling.value = true
  try {
    const cfg = await adminAPI.accounts.getAccountHealthAutoCheckConfig()
    const nextLastRunAt = cfg.last_run_at ?? null
    const hasNewAutoRun =
      nextLastRunAt !== null &&
      lastObservedAutoRunAt.value !== null &&
      nextLastRunAt !== lastObservedAutoRunAt.value

    autoConfig.enabled = cfg.enabled
    autoConfig.interval_minutes = cfg.interval_minutes || 60
    autoConfig.model_id = cfg.model_id || ''
    autoConfig.last_run_at = nextLastRunAt
    autoIntervalInput.value = String(autoConfig.interval_minutes)

    if (hasNewAutoRun) {
      await loadHealthSummary()
    }

    lastObservedAutoRunAt.value = nextLastRunAt
  } catch (error) {
    console.error('Failed to poll account health auto-check state:', error)
  } finally {
    autoHealthPolling.value = false
  }
}

let autoHealthPollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  try {
    await reloadPage()
    autoHealthPollTimer = setInterval(() => {
      void pollAutoHealthUpdates()
    }, AUTO_HEALTH_POLL_MS)
  } catch (error) {
    console.error('Failed to initialize account health page:', error)
  }
})

onUnmounted(() => {
  if (autoHealthPollTimer) {
    clearInterval(autoHealthPollTimer)
    autoHealthPollTimer = null
  }
})
</script>
