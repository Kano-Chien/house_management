<script setup>
import AppModal from './AppModal.vue'
import AppButton from './AppButton.vue'

defineProps({
  title:          { type: String, default: 'Are you sure?' },
  message:        { type: String, required: true },
  confirmLabel:   { type: String, default: 'Delete' },
  confirmVariant: { type: String, default: 'danger' },
})
const emit = defineEmits(['confirm', 'cancel'])
</script>

<template>
  <AppModal :title="title" @close="$emit('cancel')">
    <div class="px-5 py-4">
      <p class="text-sm text-gray-600 mb-6">{{ message }}</p>
      <div class="flex gap-3 justify-end">
        <AppButton variant="secondary" @click="$emit('cancel')">Cancel</AppButton>
        <AppButton :variant="confirmVariant" @click="$emit('confirm')">{{ confirmLabel }}</AppButton>
      </div>
    </div>
  </AppModal>
</template>
