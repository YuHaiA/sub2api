<template>
  <div class="grid gap-6 xl:grid-cols-[1.25fr,0.95fr]">
    <div class="space-y-6">
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
          <p class="text-xs font-medium uppercase tracking-wide text-amber-700 dark:text-amber-300">{{ t('admin.accounts.healthSummary.constrained') }}</p>
          <p class="mt-3 text-3xl font-semibold text-amber-700 dark:text-amber-200">{{ healthSummary.constrained_accounts }}</p>
          <p class="mt-2 text-xs text-amber-600/80 dark:text-amber-300/80">{{ t('admin.accounts.healthSummary.constrainedHint') }}</p>
        </div>
        <div class="rounded-2xl border border-rose-200 bg-rose-50/90 p-4 dark:border-rose-900/40 dark:bg-rose-900/10">
          <p class="text-xs font-medium uppercase tracking-wide text-rose-700 dark:text-rose-300">{{ t('admin.accounts.healthSummary.unavailable') }}</p>
          <p class="mt-3 text-3xl font-semibold text-rose-700 dark:text-rose-200">{{ healthSummary.unavailable_accounts }}</p>
          <p class="mt-2 text-xs text-rose-600/80 dark:text-rose-300/80">{{ t('admin.accounts.healthSummary.unavailableHint') }}</p>
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
          <span>{{ t('admin.accounts.healthSummary.lastChecked', { time: autoLastRunText }) }}</span>
          <span class="badge text-xs" :class="autoConfig.enabled ? 'badge-success' : 'badge-gray'">
            {{ autoConfig.enabled ? t('common.enabled') : t('common.disabled') }}
          </span>
        </div>
      </div>

      <div class="flex flex-wrap justify-end gap-2">
        <button class="btn btn-danger" :disabled="deletingUnhealthy || healthChecking" @click="$emit('deleteUnhealthy')">
          {{ deletingUnhealthy ? t('admin.accounts.deleteUnhealthyRunning') : t('admin.accounts.deleteUnhealthy') }}
        </button>
        <button class="btn btn-secondary" :disabled="healthChecking" @click="$emit('runHealthCheck')">
          {{ healthChecking ? t('admin.accounts.healthCheckRunning') : t('admin.accounts.healthCheckAll') }}
        </button>
        <button class="btn btn-primary" :disabled="savingAutoConfig" @click="$emit('saveConfig')">
          {{ savingAutoConfig ? t('common.saving') : t('admin.accounts.autoCheckSave') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Input from '@/components/common/Input.vue'
import type { AccountHealthAutoCheckConfig, AccountHealthSummary } from '@/api/admin/accounts'

defineProps<{
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
</script>
