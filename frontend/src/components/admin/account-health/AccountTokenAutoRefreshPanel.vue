<template>
  <div class="grid gap-6 xl:grid-cols-[1.15fr,1fr]">
    <div class="grid gap-4 sm:grid-cols-3">
      <div class="rounded-2xl border border-slate-200 bg-white/80 p-4 backdrop-blur dark:border-slate-700 dark:bg-slate-800/70">
        <p class="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">{{ t('admin.accounts.tokenRefresh.lastRunTotal') }}</p>
        <p class="mt-3 text-3xl font-semibold text-slate-900 dark:text-white">{{ tokenConfig.last_run_total ?? 0 }}</p>
      </div>
      <div class="rounded-2xl border border-emerald-200 bg-emerald-50/90 p-4 dark:border-emerald-900/40 dark:bg-emerald-900/10">
        <p class="text-xs font-medium uppercase tracking-wide text-emerald-700 dark:text-emerald-300">{{ t('admin.accounts.tokenRefresh.lastRunSuccess') }}</p>
        <p class="mt-3 text-3xl font-semibold text-emerald-700 dark:text-emerald-200">{{ tokenConfig.last_run_success ?? 0 }}</p>
      </div>
      <div class="rounded-2xl border border-rose-200 bg-rose-50/90 p-4 dark:border-rose-900/40 dark:bg-rose-900/10">
        <p class="text-xs font-medium uppercase tracking-wide text-rose-700 dark:text-rose-300">{{ t('admin.accounts.tokenRefresh.lastRunFailed') }}</p>
        <p class="mt-3 text-3xl font-semibold text-rose-700 dark:text-rose-200">{{ tokenConfig.last_run_failed ?? 0 }}</p>
      </div>
    </div>

    <div class="space-y-4 rounded-[24px] border border-slate-200 bg-white/85 p-5 backdrop-blur dark:border-slate-700 dark:bg-slate-800/75">
      <div>
        <h3 class="text-base font-semibold text-slate-900 dark:text-white">{{ t('admin.accounts.tokenRefresh.title') }}</h3>
        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('admin.accounts.tokenRefresh.hint') }}</p>
      </div>

      <label class="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-700 dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-200">
        <input v-model="tokenConfig.enabled" type="checkbox" class="h-4 w-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500" />
        {{ t('admin.accounts.tokenRefresh.enabled') }}
      </label>

      <div class="grid gap-4 sm:grid-cols-[1fr,180px]">
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ t('admin.accounts.tokenRefresh.interval') }}
          </label>
          <Input :model-value="tokenIntervalValueInput" type="number" :placeholder="'1'" @update:model-value="$emit('update:tokenIntervalValueInput', $event)" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ t('admin.accounts.tokenRefresh.unit') }}
          </label>
          <select v-model="tokenConfig.interval_unit" class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-900 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-200 dark:border-slate-600 dark:bg-slate-900 dark:text-white">
            <option value="hour">{{ t('admin.accounts.tokenRefresh.unitHour') }}</option>
            <option value="day">{{ t('admin.accounts.tokenRefresh.unitDay') }}</option>
          </select>
        </div>
      </div>

      <div>
        <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
          {{ t('admin.accounts.tokenRefresh.batchSize') }}
        </label>
        <Input :model-value="tokenBatchSizeInput" type="number" :placeholder="'10'" @update:model-value="$emit('update:tokenBatchSizeInput', $event)" />
      </div>

      <div class="rounded-2xl bg-slate-50 px-4 py-3 text-sm text-slate-600 dark:bg-slate-900/70 dark:text-slate-300">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span>{{ t('admin.accounts.tokenRefresh.lastRunAt', { time: tokenLastRunText }) }}</span>
          <span class="badge text-xs" :class="tokenConfig.enabled ? 'badge-success' : 'badge-gray'">
            {{ tokenConfig.enabled ? t('common.enabled') : t('common.disabled') }}
          </span>
        </div>
      </div>

      <div class="flex justify-end">
        <button class="btn btn-primary" :disabled="savingTokenConfig" @click="$emit('saveConfig')">
          {{ savingTokenConfig ? t('common.saving') : t('admin.accounts.tokenRefresh.save') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Input from '@/components/common/Input.vue'
import type { AccountTokenAutoRefreshConfig } from '@/api/admin/accounts'

defineProps<{
  tokenConfig: AccountTokenAutoRefreshConfig
  tokenIntervalValueInput: string
  tokenBatchSizeInput: string
  tokenLastRunText: string
  savingTokenConfig: boolean
}>()

defineEmits<{
  (e: 'update:tokenIntervalValueInput', value: string): void
  (e: 'update:tokenBatchSizeInput', value: string): void
  (e: 'saveConfig'): void
}>()

const { t } = useI18n()
</script>
