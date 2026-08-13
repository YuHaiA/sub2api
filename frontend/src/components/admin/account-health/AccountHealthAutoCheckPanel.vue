<template>
  <div class="grid gap-5 xl:grid-cols-[minmax(0,1.15fr)_minmax(360px,440px)] xl:items-start">
    <div>
      <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
        <div class="health-stat-card border-slate-200 bg-white/90 dark:border-slate-700 dark:bg-slate-800/70">
          <p class="health-stat-label text-slate-500 dark:text-slate-400">
            {{ t('admin.accounts.healthSummary.total') }}
          </p>
          <p class="health-stat-value text-slate-900 dark:text-white">
            {{ displayTotal }}
          </p>
        </div>
        <div class="health-stat-card border-emerald-200 bg-emerald-50/80 dark:border-emerald-900/40 dark:bg-emerald-900/10">
          <p class="health-stat-label text-emerald-700 dark:text-emerald-300">
            {{ t('admin.accounts.healthSummary.healthy') }}
          </p>
          <p class="health-stat-value text-emerald-700 dark:text-emerald-200">
            {{ healthSummary.healthy_accounts }}
          </p>
          <p class="health-stat-hint text-emerald-600/80 dark:text-emerald-300/80">
            {{ t('admin.accounts.healthSummary.healthyHint') }}
          </p>
        </div>
        <div class="health-stat-card border-amber-200 bg-amber-50/75 dark:border-amber-900/40 dark:bg-amber-900/10">
          <p class="health-stat-label text-amber-700 dark:text-amber-300">
            {{ t('admin.accounts.healthSummary.constrained') }}
          </p>
          <p class="health-stat-value text-amber-700 dark:text-amber-200">
            {{ healthSummary.constrained_accounts }}
          </p>
          <p class="health-stat-hint text-amber-600/80 dark:text-amber-300/80">
            {{ t('admin.accounts.healthSummary.constrainedHint') }}
          </p>
        </div>
        <div class="health-stat-card border-rose-200 bg-rose-50/75 dark:border-rose-900/40 dark:bg-rose-900/10">
          <p class="health-stat-label text-rose-700 dark:text-rose-300">
            {{ t('admin.accounts.healthSummary.unavailable') }}
          </p>
          <p class="health-stat-value text-rose-700 dark:text-rose-200">
            {{ healthSummary.unavailable_accounts }}
          </p>
          <div class="health-stat-hint space-y-1 text-rose-600/80 dark:text-rose-300/80">
            <p>{{ t('admin.accounts.healthSummary.unavailableHint') }}</p>
            <p>
              {{
                t('admin.accounts.healthSummary.unchecked', {
                  count: healthSummary.unchecked_accounts,
                })
              }}
            </p>
          </div>
        </div>

        <div class="health-stat-card border-violet-200 bg-violet-50/70 dark:border-violet-900/40 dark:bg-violet-900/10">
          <p class="health-stat-label text-violet-700 dark:text-violet-300">
            {{ t('admin.accounts.healthSummary.completedCount') }}
          </p>
          <p class="health-stat-value text-violet-700 dark:text-violet-200">
            {{ completedCount }}
          </p>
          <div class="health-stat-hint text-violet-600/80 dark:text-violet-300/80">
            {{ progressHint }}
          </div>
        </div>

        <div class="health-stat-card border-sky-200 bg-sky-50/75 dark:border-sky-900/40 dark:bg-sky-900/10">
          <p class="health-stat-label text-sky-700 dark:text-sky-300">
            {{ t('admin.accounts.healthSummary.pendingCount') }}
          </p>
          <p class="health-stat-value text-sky-700 dark:text-sky-200">
            {{ pendingCount }}
          </p>
          <div class="health-stat-hint text-sky-600/80 dark:text-sky-300/80">
            {{ progressHint }}
          </div>
        </div>
      </div>

      <div class="mt-4 space-y-3 rounded-2xl border border-rose-200 bg-rose-50/80 p-4 dark:border-rose-900/40 dark:bg-rose-950/20">
        <div>
          <p class="text-sm font-semibold text-rose-800 dark:text-rose-200">
            {{ t('admin.accounts.deleteStatusTitle') }}
          </p>
          <p class="mt-1 text-xs leading-5 text-rose-700/80 dark:text-rose-300/80">
            {{ t('admin.accounts.deleteStatusHint') }}
          </p>
        </div>

        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          <label v-for="option in accountDeleteOptions" :key="option.value" class="flex items-center gap-2 text-sm text-rose-800 dark:text-rose-100">
            <input
              type="checkbox"
              class="h-4 w-4 rounded border-rose-300 text-rose-600 focus:ring-rose-500"
              :checked="deleteAccountStatuses.includes(option.value)"
              @change="toggleDeleteAccountStatus(option.value, ($event.target as HTMLInputElement).checked)"
            />
            {{ option.label }}
          </label>
        </div>

        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          <label v-for="option in healthDeleteOptions" :key="option.value" class="flex items-center gap-2 text-sm text-rose-800 dark:text-rose-100">
            <input
              type="checkbox"
              class="h-4 w-4 rounded border-rose-300 text-rose-600 focus:ring-rose-500"
              :checked="deleteHealthStatuses.includes(option.value)"
              @change="toggleDeleteHealthStatus(option.value, ($event.target as HTMLInputElement).checked)"
            />
            {{ option.label }}
          </label>
        </div>

        <div class="flex justify-end border-t border-rose-200/70 pt-3 dark:border-rose-900/50">
          <button class="btn btn-danger w-full sm:w-auto" :disabled="deleteDisabled || healthChecking" @click="$emit('deleteUnhealthy')">
            {{ deletingUnhealthy ? t('admin.accounts.deleteUnhealthyRunning') : t('admin.accounts.deleteUnhealthy') }}
          </button>
        </div>
      </div>
    </div>

    <div class="space-y-4 rounded-[24px] border border-slate-200 bg-white/90 p-6 backdrop-blur xl:flex xl:h-full xl:flex-col dark:border-slate-700 dark:bg-slate-800/75">
      <div>
        <h3 class="text-base font-semibold text-slate-900 dark:text-white">
          {{ t('admin.accounts.autoCheck') }}
        </h3>
        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
          {{ t('admin.accounts.autoCheckIntervalHint') }}
        </p>
      </div>

      <label class="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3.5 text-sm font-medium text-slate-700 dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-200">
        <input
          :checked="autoConfig.enabled"
          type="checkbox"
          class="h-4 w-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
          @change="
            updateAutoConfig({
              enabled: ($event.target as HTMLInputElement).checked,
            })
          "
        />
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
            {{ t('admin.accounts.healthCheckGroup') }}
          </label>
          <select
            :value="manualGroup"
            class="input h-10"
            :disabled="healthChecking"
            @change="$emit('update:manualGroup', ($event.target as HTMLSelectElement).value)"
          >
            <option value="">{{ t('admin.accounts.allGroups') }}</option>
            <option value="ungrouped">{{ t('admin.accounts.ungroupedGroup') }}</option>
            <option v-for="group in groups" :key="group.id" :value="String(group.id)">
              {{ group.name }}
            </option>
          </select>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ t('admin.accounts.healthCheckStatus') }}
          </label>
          <Select
            :model-value="manualStatus"
            :options="manualStatusOptions"
            size="sm"
            :disabled="healthChecking"
            @update:model-value="$emit('update:manualStatus', String($event ?? ''))"
          />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ t('admin.accounts.autoCheckInterval') }}
          </label>
          <div class="grid grid-cols-[minmax(0,1fr)_112px] gap-2">
            <Input
              :model-value="autoIntervalInput"
              type="number"
              min="1"
              :placeholder="autoIntervalUnit === 'hour' ? '1' : '60'"
              @update:model-value="$emit('update:autoIntervalInput', $event)"
            />
            <select
              :value="autoIntervalUnit"
              class="input h-10"
              @change="$emit('update:autoIntervalUnit', ($event.target as HTMLSelectElement).value as 'minute' | 'hour')"
            >
              <option value="minute">{{ t('admin.accounts.intervalUnit.minute') }}</option>
              <option value="hour">{{ t('admin.accounts.intervalUnit.hour') }}</option>
            </select>
          </div>
        </div>
      </div>

      <div>
        <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
          {{ t('admin.accounts.autoCheckModel') }}
        </label>
        <Input :model-value="autoConfig.model_id" :placeholder="t('admin.accounts.healthCheckModelPlaceholder')" @update:model-value="updateAutoConfig({ model_id: $event })" />
      </div>

      <div class="rounded-2xl bg-slate-50 px-4 py-3 text-sm text-slate-600 dark:bg-slate-900/70 dark:text-slate-300">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span>{{ statusText }}</span>
          <span class="badge text-xs" :class="autoConfig.enabled ? 'badge-success' : 'badge-gray'">
            {{ autoConfig.running ? t('admin.accounts.healthCheckRunning') : autoConfig.enabled ? t('common.enabled') : t('common.disabled') }}
          </span>
        </div>
        <div class="mt-2 space-y-1 text-xs text-slate-500 dark:text-slate-400">
          <div>
            {{ t('admin.accounts.queueRunning', { task: queueRunningText }) }}
          </div>
          <div>
            {{ t('admin.accounts.queuePending', { task: queuePendingText }) }}
          </div>
        </div>
      </div>

      <div class="grid gap-3 border-t border-slate-200 pt-3 sm:grid-cols-2 xl:mt-auto dark:border-slate-700">
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
import Select from '@/components/common/Select.vue'
import type { AccountHealthAutoCheckConfig, AccountHealthSummary, DeleteAccountStatus, DeleteHealthStatus } from '@/api/admin/accounts'
import type { AdminGroup } from '@/types'

const props = defineProps<{
  healthSummary: AccountHealthSummary
  autoConfig: AccountHealthAutoCheckConfig
  manualModelId: string
  manualGroup: string
  manualStatus: string
  autoIntervalInput: string
  autoIntervalUnit: 'minute' | 'hour'
  autoLastRunText: string
  healthChecking: boolean
  savingAutoConfig: boolean
  deletingUnhealthy: boolean
  groups: AdminGroup[]
  deleteAccountStatuses: DeleteAccountStatus[]
  deleteHealthStatuses: DeleteHealthStatus[]
}>()

const emit = defineEmits<{
  (e: 'update:autoConfig', value: AccountHealthAutoCheckConfig): void
  (e: 'update:manualModelId', value: string): void
  (e: 'update:manualGroup', value: string): void
  (e: 'update:manualStatus', value: string): void
  (e: 'update:autoIntervalInput', value: string): void
  (e: 'update:autoIntervalUnit', value: 'minute' | 'hour'): void
  (e: 'update:deleteAccountStatuses', value: DeleteAccountStatus[]): void
  (e: 'update:deleteHealthStatuses', value: DeleteHealthStatus[]): void
  (e: 'runHealthCheck'): void
  (e: 'saveConfig'): void
  (e: 'deleteUnhealthy'): void
}>()

const { t } = useI18n()

function updateAutoConfig(patch: Partial<AccountHealthAutoCheckConfig>) {
  emit('update:autoConfig', { ...props.autoConfig, ...patch })
}

function toggleDeleteAccountStatus(status: DeleteAccountStatus, checked: boolean) {
  emit('update:deleteAccountStatuses', toggleValue(props.deleteAccountStatuses, status, checked))
}

function toggleDeleteHealthStatus(status: DeleteHealthStatus, checked: boolean) {
  emit('update:deleteHealthStatuses', toggleValue(props.deleteHealthStatuses, status, checked))
}

function toggleValue<T extends string>(values: T[], value: T, checked: boolean): T[] {
  if (checked) {
    return values.includes(value) ? values : [...values, value]
  }
  return values.filter((item) => item !== value)
}

const accountDeleteOptions = computed<Array<{ value: DeleteAccountStatus; label: string }>>(() => [
  { value: 'disabled', label: t('admin.accounts.deleteAccountStatus.disabled') },
  { value: 'error', label: t('admin.accounts.deleteAccountStatus.error') },
  { value: 'rate_limited', label: t('admin.accounts.deleteAccountStatus.rate_limited') },
  { value: 'temp_unschedulable', label: t('admin.accounts.deleteAccountStatus.temp_unschedulable') },
  { value: 'unschedulable', label: t('admin.accounts.deleteAccountStatus.unschedulable') },
])

const healthDeleteOptions = computed<Array<{ value: DeleteHealthStatus; label: string }>>(() => [
  { value: 'unavailable', label: t('admin.accounts.healthStatus.unavailable') },
  { value: 'constrained', label: t('admin.accounts.healthStatus.constrained') },
  { value: 'unchecked', label: t('admin.accounts.healthStatus.unchecked') },
])

const manualStatusOptions = computed(() => [
  { value: '', label: t('admin.accounts.healthStatus.all') },
  { value: 'healthy', label: t('admin.accounts.healthStatus.healthy') },
  { value: 'constrained', label: t('admin.accounts.healthStatus.constrained') },
  { value: 'unavailable', label: t('admin.accounts.healthStatus.unavailable') },
  { value: 'unchecked', label: t('admin.accounts.healthStatus.unchecked') },
])

const deleteDisabled = computed(() => props.deletingUnhealthy || (props.deleteAccountStatuses.length === 0 && props.deleteHealthStatuses.length === 0))

const statusText = computed(() => {
  if (props.autoConfig.running) {
    return t('admin.accounts.healthCheckProgress', {
      current: props.autoConfig.current_success ?? 0,
      total: props.autoConfig.current_total ?? 0,
      failed: props.autoConfig.current_failed ?? 0,
    })
  }
  return t('admin.accounts.healthSummary.lastChecked', {
    time: props.autoLastRunText,
  })
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

<style scoped>
.health-stat-card {
  @apply flex min-h-[128px] flex-col rounded-2xl border p-4 shadow-sm;
}

.health-stat-label {
  @apply text-[11px] font-semibold uppercase tracking-[0.12em];
}

.health-stat-value {
  @apply mt-3 text-[34px] font-semibold leading-none;
}

.health-stat-hint {
  @apply mt-auto pt-4 text-xs leading-5;
}
</style>
