<script setup>
defineProps({
  variant: {
    type: String,
    default: 'primary',
    validator: (v) => ['primary', 'secondary', 'danger', 'icon'].includes(v),
  },
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['sm', 'md', 'lg'].includes(v),
  },
  disabled: Boolean,
  loading:  Boolean,
})
</script>

<template>
  <button
    :disabled="disabled || loading"
    :aria-disabled="disabled || loading"
    :class="[
      'inline-flex items-center justify-center font-medium rounded-xl transition-colors',
      'focus:outline-none focus:ring-2 focus:ring-offset-2',
      variant === 'primary'   && 'bg-primary-500 hover:bg-primary-600 text-white focus:ring-primary-500',
      variant === 'secondary' && 'bg-gray-100 hover:bg-gray-200 text-gray-700 focus:ring-gray-400',
      variant === 'danger'    && 'bg-red-500 hover:bg-red-600 text-white focus:ring-red-500',
      variant === 'icon'      && 'bg-gray-100 hover:bg-gray-200 text-gray-600 focus:ring-gray-400',
      size === 'sm' && 'px-3 py-1.5 text-xs',
      size === 'md' && variant !== 'icon' && 'px-4 py-2 text-sm',
      size === 'md' && variant === 'icon' && 'w-10 h-10 text-base',
      size === 'lg' && 'px-6 py-3 text-base',
      (disabled || loading) && 'opacity-50 cursor-not-allowed',
    ]"
  >
    <slot />
  </button>
</template>
