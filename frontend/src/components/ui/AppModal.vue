<script setup>
import { onMounted, onUnmounted, ref } from 'vue'

const props = defineProps({
  title: String,
  maxWidth: { type: String, default: 'max-w-md' },
  zIndex: { type: String, default: 'z-50' },
})
const emit = defineEmits(['close'])

const panelEl = ref(null)

const modalStack = (() => {
  if (!window.__appModalStack) window.__appModalStack = []
  return window.__appModalStack
})()

function onKey(e) {
  if (e.key === 'Escape' && modalStack[modalStack.length - 1] === panelEl.value) {
    emit('close')
  }
}

onMounted(() => {
  modalStack.push(panelEl.value)
  document.addEventListener('keydown', onKey)
  document.body.style.overflow = 'hidden'
})

onUnmounted(() => {
  const idx = modalStack.indexOf(panelEl.value)
  if (idx > -1) modalStack.splice(idx, 1)
  document.removeEventListener('keydown', onKey)
  if (modalStack.length === 0) document.body.style.overflow = ''
})
</script>

<template>
  <div :class="['fixed inset-0 flex items-center justify-center p-4', zIndex]">
    <!-- Backdrop -->
    <div
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      @click="$emit('close')"
    />
    <!-- Panel -->
    <div
      ref="panelEl"
      role="dialog"
      aria-modal="true"
      :aria-label="title"
      :class="['relative bg-white rounded-2xl shadow-xl w-full flex flex-col max-h-[90vh]', maxWidth]"
    >
      <!-- Header -->
      <div v-if="title" class="flex items-center justify-between px-5 py-4 border-b border-gray-100 flex-shrink-0">
        <h2 class="text-base font-semibold text-gray-900">{{ title }}</h2>
        <button
          @click="$emit('close')"
          class="w-8 h-8 flex items-center justify-center rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
          aria-label="Close"
        >
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/>
          </svg>
        </button>
      </div>
      <!-- Body -->
      <div class="overflow-y-auto flex-1">
        <slot />
      </div>
      <!-- Footer slot (optional) -->
      <div v-if="$slots.footer" class="border-t border-gray-100 flex-shrink-0">
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>
