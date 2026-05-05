<template>
  <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(380px,460px)] xl:items-stretch">
    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 xl:h-full xl:auto-rows-fr">
      <div class="flex min-h-[168px] flex-col rounded-2xl border border-slate-200 bg-white/80 p-5 backdrop-blur xl:h-full dark:border-slate-700 dark:bg-slate-800/70">
        <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">{{ t('admin.accounts.tokenRefresh.lastRunTotal') }}</p>
        <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-slate-900 dark:text-white">{{ tokenConfig.running ? (tokenConfig.current_total ?? 0) : (tokenConfig.last_run_total ?? 0) }}</p>
        <div class="mt-auto pt-6"></div>
      </div>

      <div class="flex min-h-[168px] flex-col rounded-2xl border border-emerald-200 bg-emerald-50/90 p-5 xl:h-full dark:border-emerald-900/40 dark:bg-emerald-900/10">
        <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-emerald-700 dark:text-emerald-300">{{ t('admin.accounts.tokenRefresh.lastRunSuccess') }}</p>
        <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-emerald-700 dark:text-emerald-200">{{ tokenConfig.running ? (tokenConfig.current_success ?? 0) : (tokenConfig.last_run_success ?? 0) }}</p>
        <div class="mt-auto pt-6"></div>
      </div>

      <div class="flex min-h-[168px] flex-col rounded-2xl border border-rose-200 bg-rose-50/90 p-5 xl:h-full dark:border-rose-900/40 dark:bg-rose-900/10">
        <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-rose-700 dark:text-rose-300">{{ t('admin.accounts.tokenRefresh.lastRunFailed') }}</p>
        <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-rose-700 dark:text-rose-200">{{ tokenConfig.running ? (tokenConfig.current_failed ?? 0) : (tokenConfig.last_run_failed ?? 0) }}</p>
        <div class="mt-auto pt-6"></div>
      </div>

      <div class="flex min-h-[168px] flex-col rounded-2xl border border-violet-200 bg-violet-50/90 p-5 xl:h-full dark:border-violet-900/40 dark:bg-violet-900/10">
        <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-violet-700 dark:text-violet-300">{{ t('admin.accounts.tokenRefresh.completedCount') }}</p>
        <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-violet-700 dark:text-violet-200">{{ completedCount }}</p>
        <div class="mt-auto pt-6"></div>
      </div>

      <div class="flex min-h-[168px] flex-col rounded-2xl border border-amber-200 bg-amber-50/90 p-5 xl:h-full dark:border-amber-900/40 dark:bg-amber-900/10">
        <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-amber-700 dark:text-amber-300">{{ t('admin.accounts.tokenRefresh.pendingCount') }}</p>
        <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-amber-700 dark:text-amber-200">{{ pendingCount }}</p>
        <div class="mt-auto pt-6"></div>
      </div>

      <div class="flex min-h-[168px] flex-col rounded-2xl border border-sky-200 bg-sky-50/90 p-5 xl:h-full dark:border-sky-900/40 dark:bg-sky-900/10">
        <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-sky-700 dark:text-sky-300">{{ t('admin.accounts.tokenRefresh.currentBatch') }}</p>
        <p class="mt-5 text-[42px] font-semibold leading-none tracking-tight text-sky-700 dark:text-sky-200">{{ tokenConfig.batch_size }}</p>
        <p class="mt-auto pt-6 text-xs leading-5 text-sky-600/80 dark:text-sky-300/80">{{ scopeDescription }}</p>
      </div>
    </div>

    <div class="space-y-4 rounded-[24px] border border-slate-200 bg-white/90 p-6 backdrop-blur xl:flex xl:h-full xl:flex-col dark:border-slate-700 dark:bg-slate-800/75">
      <div>
        <h3 class="text-base font-semibold text-slate-900 dark:text-white">{{ t('admin.accounts.tokenRefresh.title') }}</h3>
        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('admin.accounts.tokenRefresh.hint') }}</p>
      </div>

      <label class="flex items-center gap-3 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3.5 text-sm font-medium text-slate-700 dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-200">
        <input v-model="tokenConfig.enabled" type="checkbox" class="h-4 w-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500" />
        {{ t('admin.accounts.tokenRefresh.enabled') }}
      </label>

      <div class="grid gap-4 sm:grid-cols-2">
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

      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ t('admin.accounts.tokenRefresh.batchSize') }}
          </label>
          <Input :model-value="tokenBatchSizeInput" type="number" :placeholder="'10'" @update:model-value="$emit('update:tokenBatchSizeInput', $event)" />
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
            {{ t('admin.accounts.tokenRefresh.scope') }}
          </label>
          <select v-model="tokenConfig.scope" class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-900 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-200 dark:border-slate-600 dark:bg-slate-900 dark:text-white">
            <option value="all">{{ t('admin.accounts.tokenRefresh.scopeAll') }}</option>
            <option value="group">{{ t('admin.accounts.tokenRefresh.scopeGroup') }}</option>
          </select>
        </div>
      </div>

      <div v-if="tokenConfig.scope === 'group'">
        <label class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-300">
          {{ t('admin.accounts.tokenRefresh.group') }}
        </label>
        <select v-model.number="tokenConfig.group_id" class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-900 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-200 dark:border-slate-600 dark:bg-slate-900 dark:text-white">
          <option :value="0">{{ t('admin.accounts.tokenRefresh.groupPlaceholder') }}</option>
          <option v-for="group in groups" :key="group.id" :value="group.id">
            {{ group.name }}
          </option>
        </select>
      </div>

      <div class="rounded-2xl bg-slate-50 px-4 py-3 text-sm text-slate-600 dark:bg-slate-900/70 dark:text-slate-300">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span>{{ tokenConfig.running ? t('admin.accounts.tokenRefresh.runningProgress') : t('admin.accounts.tokenRefresh.lastRunAt', { time: tokenLastRunText }) }}</span>
          <span class="badge text-xs" :class="tokenConfig.enabled ? 'badge-success' : 'badge-gray'">
            {{ tokenConfig.running ? t('admin.accounts.tokenRefresh.running') : (tokenConfig.enabled ? t('common.enabled') : t('common.disabled')) }}
          </span>
        </div>
      </div>

      <div class="flex flex-wrap justify-end gap-3 border-t border-slate-200 pt-3 xl:mt-auto dark:border-slate-700">
        <button class="btn btn-secondary" :disabled="runningTokenRefresh" @click="$emit('runNow')">
          {{ runningTokenRefresh ? t('admin.accounts.tokenRefresh.running') : t('admin.accounts.tokenRefresh.runNow') }}
        </button>
        <button class="btn btn-primary" :disabled="savingTokenConfig" @click="$emit('saveConfig')">
          {{ savingTokenConfig ? t('common.saving') : t('admin.accounts.tokenRefresh.save') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AccountTokenAutoRefreshConfig } from '@/api/admin/accounts'
import Input from '@/components/common/Input.vue'
import type { AdminGroup } from '@/types'

const props = defineProps<{
  tokenConfig: AccountTokenAutoRefreshConfig
  tokenIntervalValueInput: string
  tokenBatchSizeInput: string
  tokenLastRunText: string
  savingTokenConfig: boolean
  runningTokenRefresh: boolean
  groups: AdminGroup[]
}>()

defineEmits<{
  (e: 'update:tokenIntervalValueInput', value: string): void
  (e: 'update:tokenBatchSizeInput', value: string): void
  (e: 'saveConfig'): void
  (e: 'runNow'): void
}>()

const { t } = useI18n()

const selectedGroupName = computed(() => {
  const groupID = props.tokenConfig.group_id ?? 0
  return props.groups.find((group) => group.id === groupID)?.name ?? ''
})

const scopeDescription = computed(() => {
  if (props.tokenConfig.scope === 'group') {
    return selectedGroupName.value || t('admin.accounts.tokenRefresh.scopeGroup')
  }
  return t('admin.accounts.tokenRefresh.scopeAll')
})

const completedCount = computed(() => {
  if (props.tokenConfig.running) {
    return (props.tokenConfig.current_success ?? 0) + (props.tokenConfig.current_failed ?? 0)
  }
  return (props.tokenConfig.last_run_success ?? 0) + (props.tokenConfig.last_run_failed ?? 0)
})

const pendingCount = computed(() => {
  if (!props.tokenConfig.running) {
    return 0
  }
  return Math.max((props.tokenConfig.current_total ?? 0) - completedCount.value, 0)
})
</script>
