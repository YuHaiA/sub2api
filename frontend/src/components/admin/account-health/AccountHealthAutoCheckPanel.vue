<template>
  <div class="grid gap-6 xl:grid-cols-[minmax(0,1.2fr)_minmax(380px,460px)] xl:items-stretch">
    <div class="xl:h-full">
      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 xl:h-full xl:auto-rows-fr">
        <div class="flex min-h-[168px] flex-col rounded-2xl border border-slate-200 bg-white/80 p-5 backdrop-blur xl:h-full dark:border-slate-700 dark:bg-slate-800/70">
          <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">{{ t('admin.accounts.healthSummary.total') }}</p>
          <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-slate-900 dark:text-white">{{ displayTotal }}</p>
          <div class="mt-auto pt-6"></div>
        </div>
        <div class="flex min-h-[168px] flex-col rounded-2xl border border-emerald-200 bg-emerald-50/90 p-5 xl:h-full dark:border-emerald-900/40 dark:bg-emerald-900/10">
          <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-emerald-700 dark:text-emerald-300">{{ t('admin.accounts.healthSummary.healthy') }}</p>
          <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-emerald-700 dark:text-emerald-200">{{ healthSummary.healthy_accounts }}</p>
          <p class="mt-auto pt-6 text-xs leading-5 text-emerald-600/80 dark:text-emerald-300/80">{{ t('admin.accounts.healthSummary.healthyHint') }}</p>
        </div>
        <div class="flex min-h-[168px] flex-col rounded-2xl border border-amber-200 bg-amber-50/90 p-5 xl:h-full dark:border-amber-900/40 dark:bg-amber-900/10">
          <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-amber-700 dark:text-amber-300">{{ t('admin.accounts.healthSummary.constrained') }}</p>
          <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-amber-700 dark:text-amber-200">{{ healthSummary.constrained_accounts }}</p>
          <p class="mt-auto pt-6 text-xs leading-5 text-amber-600/80 dark:text-amber-300/80">{{ t('admin.accounts.healthSummary.constrainedHint') }}</p>
        </div>
        <div class="flex min-h-[168px] flex-col rounded-2xl border border-rose-200 bg-rose-50/90 p-5 xl:h-full dark:border-rose-900/40 dark:bg-rose-900/10">
          <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-rose-700 dark:text-rose-300">{{ t('admin.accounts.healthSummary.unavailable') }}</p>
          <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-rose-700 dark:text-rose-200">{{ healthSummary.unavailable_accounts }}</p>
          <div class="mt-auto space-y-2 pt-6 text-xs leading-5 text-rose-600/80 dark:text-rose-300/80">
            <p>{{ t('admin.accounts.healthSummary.unavailableHint') }}</p>
            <p>
            {{ t('admin.accounts.healthSummary.unchecked', { count: healthSummary.unchecked_accounts }) }}
            </p>
          </div>
        </div>

        <div class="flex min-h-[168px] flex-col rounded-2xl border border-violet-200 bg-violet-50/90 p-5 xl:h-full dark:border-violet-900/40 dark:bg-violet-900/10">
          <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-violet-700 dark:text-violet-300">{{ t('admin.accounts.healthSummary.completedCount') }}</p>
          <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-violet-700 dark:text-violet-200">{{ completedCount }}</p>
          <div class="mt-auto pt-6 text-xs leading-5 text-violet-600/80 dark:text-violet-300/80">{{ progressHint }}</div>
        </div>

        <div class="flex min-h-[168px] flex-col rounded-2xl border border-amber-200 bg-amber-50/90 p-5 xl:h-full dark:border-amber-900/40 dark:bg-amber-900/10">
          <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-amber-700 dark:text-amber-300">{{ t('admin.accounts.healthSummary.pendingCount') }}</p>
          <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-amber-700 dark:text-amber-200">{{ pendingCount }}</p>
          <div class="mt-auto pt-6 text-xs leading-5 text-amber-600/80 dark:text-amber-300/80">{{ progressHint }}</div>
        </div>
      </div>
    </div>

    <div class="space-y-4 rounded-[24px] border border-slate-200 bg-white/90 p-6 backdrop-blur xl:flex xl:h-full xl:flex-col dark:border-slate-700 dark:bg-slate-800/75">
      <div>
        <h3 class="text-base font-semibold text-slate-900 dark:text-white">{{ t('admin.accounts.autoCheck') }}</h3>
        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('admin.accounts.autoCheckIntervalHint') }}</p>
      </div>

      <label class="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3.5 text-sm font-medium text-slate-700 dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-200">
        <input v-model="autoConfig.enabled" type="checkbox" class="h-4 w-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500" />
        {{ t('admin.accounts.autoCheckEnabled') }}
      </label>

      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ t('admin.accounts.healthCheckModelPlaceholder') }}
          </label>
          <Input :model-value="manualModelId" :disabled="healthChecking" :placeholder="t('admin.accounts.healthCheckModelPlaceholder')" @update:model-value="$emit('update:manualModelId', $event)" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ t('admin.accounts.autoCheckInterval') }}
          </label>
          <Input :model-value="autoIntervalInput" type="number" :placeholder="'60'" @update:model-value="$emit('update:autoIntervalInput', $event)" />
        </div>
      </div>

      <div>
        <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
          {{ t('admin.accounts.autoCheckModel') }}
        </label>
        <Input v-model="autoConfig.model_id" :placeholder="t('admin.accounts.healthCheckModelPlaceholder')" />
      </div>

      <div class="rounded-2xl bg-slate-50 px-4 py-3 text-sm text-slate-600 dark:bg-slate-900/70 dark:text-slate-300">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span>{{ statusText }}</span>
          <span class="badge text-xs" :class="autoConfig.enabled ? 'badge-success' : 'badge-gray'">
            {{ autoConfig.running ? t('admin.accounts.healthCheckRunning') : (autoConfig.enabled ? t('common.enabled') : t('common.disabled')) }}
          </span>
        </div>
        <div class="mt-2 space-y-1 text-xs text-slate-500 dark:text-slate-400">
          <div>{{ t('admin.accounts.queueRunning', { task: queueRunningText }) }}</div>
          <div>{{ t('admin.accounts.queuePending', { task: queuePendingText }) }}</div>
        </div>
      </div>

      <div class="grid gap-3 border-t border-slate-200 pt-3 sm:grid-cols-3 xl:mt-auto dark:border-slate-700">
        <button class="btn btn-danger w-full" :disabled="deletingUnhealthy || healthChecking" @click="$emit('deleteUnhealthy')">
          {{ deletingUnhealthy ? t('admin.accounts.deleteUnhealthyRunning') : t('admin.accounts.deleteUnhealthy') }}
        </button>
        <button class="btn btn-secondary w-full" :disabled="healthChecking" @click="$emit('runHealthCheck')">
          {{ healthChecking ? t('admin.accounts.healthCheckRunning') : t('admin.accounts.healthCheckAll') }}
        </button>
        <button class="btn btn-primary w-full" :disabled="savingAutoConfig" @click="$emit('saveConfig')">
          {{ savingAutoConfig ? t('common.saving') : t('admin.accounts.autoCheckSave') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Input from '@/components/common/Input.vue'
import type { AccountHealthAutoCheckConfig, AccountHealthSummary } from '@/api/admin/accounts'

const props = defineProps<{
  healthSummary: AccountHealthSummary
  autoConfig: AccountHealthAutoCheckConfig
  manualModelId: string
  autoIntervalInput: string
  autoLastRunText: string
  healthChecking: boolean
  savingAutoConfig: boolean
  deletingUnhealthy: boolean
}>()

defineEmits<{
  (e: 'update:manualModelId', value: string): void
  (e: 'update:autoIntervalInput', value: string): void
  (e: 'runHealthCheck'): void
  (e: 'saveConfig'): void
  (e: 'deleteUnhealthy'): void
}>()

const { t } = useI18n()

const statusText = computed(() => {
  if (props.autoConfig.running) {
    return t('admin.accounts.healthCheckProgress', {
      current: props.autoConfig.current_success ?? 0,
      total: props.autoConfig.current_total ?? 0,
      failed: props.autoConfig.current_failed ?? 0
    })
  }
  return t('admin.accounts.healthSummary.lastChecked', { time: props.autoLastRunText })
})

const completedCount = computed(() => {
  if (!props.autoConfig.running) {
    return 0
  }
  return (props.autoConfig.current_success ?? 0) + (props.autoConfig.current_failed ?? 0)
})

const pendingCount = computed(() => {
  if (!props.autoConfig.running) {
    return 0
  }
  return Math.max((props.autoConfig.current_total ?? 0) - completedCount.value, 0)
})

const displayTotal = computed(() => {
  if (props.autoConfig.running) {
    return props.autoConfig.current_total ?? 0
  }
  return props.healthSummary.total_accounts
})

const progressHint = computed(() => {
  return props.autoConfig.running ? t('admin.accounts.healthSummary.progressLive') : t('admin.accounts.healthSummary.progressIdle')
})

const queueRunningText = computed(() => formatQueueTask(props.autoConfig.queue_running))
const queuePendingText = computed(() => formatQueueTask(props.autoConfig.queue_pending))

function formatQueueTask(task?: string) {
  switch (task) {
    case 'account_health_manual':
      return t('admin.accounts.queueTask.healthManual')
    case 'account_health_auto':
      return t('admin.accounts.queueTask.healthAuto')
    case 'token_refresh_manual':
      return t('admin.accounts.queueTask.refreshManual')
    case 'token_refresh_auto':
      return t('admin.accounts.queueTask.refreshAuto')
    default:
      return t('admin.accounts.queueTask.none')
  }
}
</script>
