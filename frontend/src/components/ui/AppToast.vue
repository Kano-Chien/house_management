<script setup>
import { useToast } from '../../composables/useToast.js'
const { toasts, dismiss } = useToast()
</script>

<template>
  <div class="fixed top-4 right-4 z-[60] flex flex-col gap-2 pointer-events-none">
    <TransitionGroup name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-xl shadow-lg text-sm font-medium min-w-[240px] max-w-sm cursor-pointer"
        :class="{
          'bg-green-500 text-white': t.type === 'success',
          'bg-red-500 text-white':   t.type === 'error',
          'bg-amber-500 text-white': t.type === 'warning',
        }"
        @click="dismiss(t.id)"
      >
        <span class="flex-1">{{ t.message }}</span>
        <span class="opacity-70 text-xs">✕</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active { transition: all 0.2s ease; }
.toast-enter-from   { opacity: 0; transform: translateX(0.75rem); }
.toast-leave-to     { opacity: 0; transform: translateX(0.75rem); }
</style>
