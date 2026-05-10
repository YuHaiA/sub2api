<template>
  <div class="account-toolbar">
    <div class="account-toolbar-group">
      <slot name="before"></slot>
      <button @click="$emit('refresh')" :disabled="loading" class="account-toolbar-btn account-toolbar-icon-btn">
        <Icon name="refresh" size="sm" :class="[loading ? 'animate-spin' : '']" />
      </button>
      <slot name="after"></slot>
    </div>
    <div class="account-toolbar-tail">
      <button @click="$emit('create')" class="account-toolbar-btn account-toolbar-primary">{{ t('admin.accounts.createAccount') }}</button>
      <slot name="afterCreate"></slot>
    </div>
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
  @apply flex min-w-0 flex-wrap items-center justify-between gap-x-3 gap-y-2;
}

.account-toolbar-group {
  @apply flex min-w-0 flex-wrap items-center gap-1.5;
}

.account-toolbar-tail {
  @apply flex shrink-0 flex-wrap items-center gap-1.5;
}

.account-toolbar-btn {
  @apply inline-flex h-[30px] shrink-0 items-center justify-center gap-1 whitespace-nowrap rounded-lg border border-slate-200/90 bg-white/90 px-2.5 text-xs font-medium leading-none text-slate-700 shadow-[0_1px_2px_rgba(15,23,42,0.04)] transition;
  @apply hover:border-slate-300 hover:bg-white hover:text-slate-900 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60;
  @apply dark:border-dark-600 dark:bg-dark-800/90 dark:text-dark-200 dark:hover:bg-dark-700 dark:hover:text-white;
}

.account-toolbar-icon-btn {
  @apply w-[30px] px-0;
}

.account-toolbar-primary {
  @apply border-primary-500 bg-primary-500 px-3 text-white shadow-[0_6px_16px_rgba(20,184,166,0.24)] hover:border-primary-600 hover:bg-primary-600 hover:text-white;
}
</style>
