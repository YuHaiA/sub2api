<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="space-y-4">
          <div class="grid gap-4 xl:grid-cols-[2fr,1fr]">
            <div class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-800">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.accountHealth.title') }}</h2>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.accountHealth.description') }}</p>
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <button class="btn btn-danger" :disabled="healthChecking || deletingUnhealthy" @click="deleteUnhealthyAccountsInScope">
                    {{ deletingUnhealthy ? t('admin.accounts.deleteUnhealthyRunning') : t('admin.accounts.deleteUnhealthy') }}
                  </button>
                  <button class="btn btn-secondary" :disabled="healthChecking" @click="runFilteredHealthCheck">
                    {{ healthChecking ? t('admin.accounts.healthCheckRunning') : t('admin.accounts.healthCheckAll') }}
                  </button>
                </div>
              </div>
              <div class="mt-4 flex flex-wrap items-start gap-3">
                <div class="w-full md:w-72">
                  <Input
                    v-model="manualModelId"
                    :disabled="healthChecking"
                    :label="t('admin.accounts.healthCheckModelPlaceholder')"
                    :placeholder="t('admin.accounts.healthCheckModelPlaceholder')"
                  />
                </div>
                <div class="rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm dark:border-gray-600 dark:bg-gray-900/40">
                  <label class="flex items-center gap-2 font-medium text-gray-700 dark:text-gray-200">
                    <input v-model="autoConfig.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                    {{ t('admin.accounts.autoCheckEnabled') }}
                  </label>
                  <div class="mt-3 flex flex-wrap items-end gap-3">
                    <div class="w-36">
                      <Input
                        v-model="autoIntervalInput"
                        type="number"
                        :label="t('admin.accounts.autoCheckInterval')"
                        :placeholder="'60'"
                      />
                    </div>
                    <div class="w-full md:w-72">
                      <Input
                        v-model="autoConfig.model_id"
                        :label="t('admin.accounts.autoCheckModel')"
                        :placeholder="t('admin.accounts.healthCheckModelPlaceholder')"
                      />
                    </div>
                    <button class="btn btn-primary" :disabled="savingAutoConfig" @click="saveAutoConfig">
                      {{ t('admin.accounts.autoCheckSave') }}
                    </button>
                  </div>
                  <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.accounts.autoCheckIntervalHint') }}
                    <span v-if="autoConfig.last_run_at"> · {{ t('admin.accounts.healthSummary.lastChecked', { time: autoLastRunText }) }}</span>
                  </p>
                </div>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3 xl:grid-cols-2">
              <div class="card p-4">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.healthSummary.total') }}</p>
                <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ healthSummary.total_accounts }}</p>
              </div>
              <div class="card p-4">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.healthSummary.healthy') }}</p>
                <p class="mt-2 text-2xl font-bold text-emerald-600 dark:text-emerald-400">{{ healthSummary.healthy_accounts }}</p>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.accounts.healthSummary.healthyHint') }}</p>
              </div>
              <div class="card p-4">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.healthSummary.bannedOrExhausted') }}</p>
                <p class="mt-2 text-2xl font-bold text-amber-600 dark:text-amber-400">{{ healthSummary.banned_or_exhausted_accounts }}</p>
              </div>
              <div class="card p-4">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.healthSummary.unavailable') }}</p>
                <p class="mt-2 text-2xl font-bold text-rose-600 dark:text-rose-400">{{ healthSummary.unavailable_accounts }}</p>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.accounts.healthSummary.unchecked', { count: healthSummary.unchecked_accounts }) }}</p>
              </div>
            </div>
          </div>

          <AccountTableFilters
            v-model:searchQuery="params.search"
            :filters="params"
            :groups="groups"
            @update:filters="(newFilters) => Object.assign(params, newFilters)"
            @change="debouncedReload"
            @update:searchQuery="debouncedReload"
          />
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="accounts"
          :loading="loading"
          row-key="id"
          :server-side-sort="false"
        >
          <template #cell-name="{ row }">
            <div class="flex flex-col">
              <span class="font-medium text-gray-900 dark:text-white">{{ row.name }}</span>
              <span v-if="row.health_message" class="mt-1 max-w-[320px] truncate text-xs text-gray-500 dark:text-gray-400" :title="row.health_message">
                {{ row.health_message }}
              </span>
            </div>
          </template>
          <template #cell-platform_type="{ row }">
            <PlatformTypeBadge :platform="row.platform" :type="row.type" />
          </template>
          <template #cell-status="{ row }">
            <AccountStatusIndicator :account="row" />
          </template>
          <template #cell-health_status="{ value }">
            <span :class="['badge text-xs', getHealthStatusBadgeClass(value)]">
              {{ getHealthStatusLabel(value) }}
            </span>
          </template>
          <template #cell-health_last_checked_at="{ value }">
            <span class="text-sm text-gray-600 dark:text-gray-300">
              {{ value ? formatRelativeTime(value) : t('admin.accounts.healthSummary.neverChecked') }}
            </span>
          </template>
          <template #cell-health_latency_ms="{ value }">
            <span class="text-sm text-gray-600 dark:text-gray-300">{{ value ? `${value} ms` : '-' }}</span>
          </template>
          <template #cell-actions="{ row }">
            <button class="btn btn-secondary btn-sm" :disabled="healthChecking" @click="runSingleHealthCheck(row.id)">
              {{ t('admin.accounts.healthCheckSelected') }}
            </button>
          </template>
        </DataTable>
        <Pagination
          class="mt-4"
          :page="pagination.page"
          :page-size="pagination.page_size"
          :total="pagination.total"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Input from '@/components/common/Input.vue'
import AccountTableFilters from '@/components/admin/account/AccountTableFilters.vue'
import AccountStatusIndicator from '@/components/account/AccountStatusIndicator.vue'
import PlatformTypeBadge from '@/components/common/PlatformTypeBadge.vue'
import { useTableLoader } from '@/composables/useTableLoader'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { formatRelativeTime } from '@/utils/format'
import type { Account, AdminGroup } from '@/types'
import type { AccountHealthAutoCheckConfig, AccountHealthSummary } from '@/api/admin/accounts'

const { t } = useI18n()
const appStore = useAppStore()
const groups = ref<AdminGroup[]>([])
const healthChecking = ref(false)
const savingAutoConfig = ref(false)
const deletingUnhealthy = ref(false)
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

const {
  items: accounts,
  loading,
  params,
  pagination,
  load: baseLoad,
  reload: baseReload,
  debouncedReload: baseDebouncedReload,
  handlePageChange: baseHandlePageChange,
  handlePageSizeChange: baseHandlePageSizeChange
} = useTableLoader<Account, any>({
  fetchFn: adminAPI.accounts.list,
  initialParams: {
    platform: '',
    type: '',
    status: '',
    health_status: '',
    privacy_mode: '',
    group: '',
    search: '',
    sort_by: 'name',
    sort_order: 'asc'
  }
})

const columns = computed(() => ([
  { key: 'name', label: t('admin.accounts.columns.name'), sortable: false },
  { key: 'platform_type', label: t('admin.accounts.columns.platformType'), sortable: false },
  { key: 'status', label: t('admin.accounts.columns.status'), sortable: false },
  { key: 'health_status', label: t('admin.accounts.autoCheck'), sortable: false },
  { key: 'health_last_checked_at', label: t('admin.accounts.healthSummary.lastChecked', { time: '' }), sortable: false },
  { key: 'health_latency_ms', label: 'Latency', sortable: false },
  { key: 'actions', label: t('admin.accounts.columns.actions'), sortable: false }
]))

const autoLastRunText = computed(() => {
  if (!autoConfig.last_run_at) return t('admin.accounts.healthSummary.neverChecked')
  return formatRelativeTime(new Date(autoConfig.last_run_at * 1000).toISOString())
})

const buildFilters = () => ({
  platform: params.platform || '',
  type: params.type || '',
  status: params.status || '',
  health_status: params.health_status || '',
  group: params.group || '',
  privacy_mode: params.privacy_mode || '',
  search: params.search || '',
  sort_by: 'name',
  sort_order: 'asc' as const
})

const refreshHealthSummary = async () => {
  healthSummary.value = await adminAPI.accounts.getHealthSummary(buildFilters())
}

const loadAutoConfig = async () => {
  const cfg = await adminAPI.accounts.getAccountHealthAutoCheckConfig()
  autoConfig.enabled = cfg.enabled
  autoConfig.interval_minutes = cfg.interval_minutes || 60
  autoConfig.model_id = cfg.model_id || ''
  autoConfig.last_run_at = cfg.last_run_at ?? null
  autoIntervalInput.value = String(autoConfig.interval_minutes)
}

const loadPage = async () => {
  await baseLoad()
  await refreshHealthSummary()
}

const reloadPage = async () => {
  await baseReload()
  await refreshHealthSummary()
}

const debouncedReload = () => {
  baseDebouncedReload()
}

const handlePageChange = (page: number) => {
  baseHandlePageChange(page)
}

const handlePageSizeChange = (size: number) => {
  baseHandlePageSizeChange(size)
}

watch(loading, (isLoading, wasLoading) => {
  if (wasLoading && !isLoading) {
    refreshHealthSummary().catch((error) => {
      console.error('Failed to refresh account health summary:', error)
    })
  }
})

const runFilteredHealthCheck = async () => {
  if (healthChecking.value) return
  healthChecking.value = true
  try {
    const modelID = manualModelId.value.trim() || autoConfig.model_id.trim()
    await adminAPI.accounts.runHealthCheck({
      filters: buildFilters(),
      model_id: modelID || undefined
    })
    await reloadPage()
    appStore.showSuccess(t('admin.accounts.healthCheckCompleted', { count: accounts.value.length }))
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.accounts.healthCheckFailed'))
  } finally {
    healthChecking.value = false
  }
}

const runSingleHealthCheck = async (accountID: number) => {
  if (healthChecking.value) return
  healthChecking.value = true
  try {
    const modelID = manualModelId.value.trim() || autoConfig.model_id.trim()
    await adminAPI.accounts.runHealthCheck({
      account_ids: [accountID],
      model_id: modelID || undefined
    })
    await reloadPage()
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
    const result = await adminAPI.accounts.deleteUnhealthyAccounts({
      filters: buildFilters()
    })
    await reloadPage()
    appStore.showSuccess(t('admin.accounts.deleteUnhealthyDone', { count: result.deleted_count }))
  } catch (error: any) {
    appStore.showError(error?.message || t('common.error'))
  } finally {
    deletingUnhealthy.value = false
  }
}

const getHealthStatusLabel = (status?: string) => {
  const value = status || 'unchecked'
  if (value === 'banned_or_exhausted') {
    return t('admin.accounts.healthStatus.bannedOrExhausted')
  }
  if (value === 'rate_limited') {
    return t('admin.accounts.healthStatus.rateLimited')
  }
  return t(`admin.accounts.healthStatus.${value}`)
}

const getHealthStatusBadgeClass = (status?: string) => {
  switch (status) {
    case 'healthy':
      return 'badge-success'
    case 'rate_limited':
      return 'badge-warning'
    case 'banned_or_exhausted':
      return 'badge-warning'
    case 'unavailable':
      return 'badge-danger'
    default:
      return 'badge-gray'
  }
}

onMounted(async () => {
  try {
    const [, groupItems] = await Promise.all([
      loadPage(),
      adminAPI.groups.getAll(),
      loadAutoConfig()
    ])
    groups.value = groupItems as AdminGroup[]
  } catch (error) {
    console.error('Failed to initialize account health page:', error)
  }
})
</script>
