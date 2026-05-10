<template>
  <div class="account-toolbar">
    <slot name="before"></slot>
    <button @click="$emit('refresh')" :disabled="loading" class="account-toolbar-btn account-toolbar-icon-btn">
      <Icon name="refresh" size="sm" :class="[loading ? 'animate-spin' : '']" />
    </button>
    <slot name="after"></slot>
    <button @click="$emit('sync')" class="account-toolbar-btn">{{ t('admin.accounts.syncFromCrs') }}</button>
    <slot name="beforeCreate"></slot>
    <button @click="$emit('create')" class="account-toolbar-btn account-toolbar-primary">{{ t('admin.accounts.createAccount') }}</button>
    <slot name="afterCreate"></slot>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps(['loading'])
defineEmits(['refresh', 'sync', 'create'])

const { t } = useI18n()
</script>

<style scoped>
.account-toolbar {
  @apply flex min-w-0 flex-wrap items-center gap-2;
}

.account-toolbar-btn {
  @apply inline-flex h-9 shrink-0 items-center justify-center gap-1.5 whitespace-nowrap rounded-xl border border-slate-200 bg-white px-3 text-sm font-medium text-slate-700 shadow-sm transition;
  @apply hover:border-slate-300 hover:bg-slate-50 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60;
  @apply dark:border-dark-600 dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700;
}

.account-toolbar-icon-btn {
  @apply w-9 px-0;
}

.account-toolbar-primary {
  @apply border-primary-500 bg-primary-500 text-white shadow-primary-500/20 hover:border-primary-600 hover:bg-primary-600;
}
</style>
