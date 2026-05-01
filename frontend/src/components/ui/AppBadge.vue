<script setup>
defineProps({
  variant: {
    type: String,
    default: 'default',
    validator: (v) => ['ok', 'low', 'out', 'auto', 'custom', 'default'].includes(v),
  },
  // ok | low | out | auto | custom | default
})
</script>

<template>
  <span
    :class="[
      'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold',
      variant === 'ok'      && 'bg-green-100 text-green-700',
      variant === 'low'     && 'bg-amber-100 text-amber-700',
      variant === 'out'     && 'bg-red-100 text-red-700',
      variant === 'auto'    && 'bg-primary-100 text-primary-700',
      variant === 'custom'  && 'bg-gray-100 text-gray-600',
      variant === 'default' && 'bg-gray-100 text-gray-600',
    ]"
  >
    <slot />
  </span>
</template>
