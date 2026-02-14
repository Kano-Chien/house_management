<template>
  <div class="step-editor" @click="focusEditor">
    <editor-content :editor="editor" />
  </div>
</template>

<script setup>
import { watch, onBeforeUnmount } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'

const props = defineProps({
  modelValue: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue', 'blur'])

const editor = useEditor({
  content: props.modelValue || '',
  extensions: [
    StarterKit.configure({
      heading: { levels: [1, 2, 3] },
    }),
    Placeholder.configure({
      placeholder: 'Type your recipe steps here... Use markdown shortcuts like "1. " or "- "',
    }),
  ],
  editorProps: {
    attributes: {
      class: 'prose-editor focus:outline-none min-h-[60px] p-3 text-sm text-gray-700',
    },
  },
  onUpdate({ editor }) {
    emit('update:modelValue', editor.getHTML())
  },
  onBlur() {
    emit('blur')
  },
})

watch(() => props.modelValue, (val) => {
  if (!editor.value) return
  const current = editor.value.getHTML()
  if (current !== val) {
    editor.value.commands.setContent(val || '', false)
  }
})

const focusEditor = () => {
  if (editor.value) editor.value.commands.focus()
}

onBeforeUnmount(() => {
  if (editor.value) editor.value.destroy()
})
</script>

<style scoped>
.step-editor {
  cursor: text;
  background: white;
  border: 1px solid #bfdbfe;
  border-radius: 0.5rem;
  transition: all 0.2s;
}
.step-editor:focus-within {
  border-color: #93c5fd;
  box-shadow: 0 0 0 2px rgba(147, 197, 253, 0.5);
}

/* Editor content styles */
.step-editor :deep(.ProseMirror) {
  min-height: 60px;
  outline: none;
}
.step-editor :deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: #d1d5db;
  pointer-events: none;
  height: 0;
  font-style: italic;
  font-size: 0.8rem;
}
.step-editor :deep(.ProseMirror p) {
  margin: 0.2rem 0;
}
.step-editor :deep(.ProseMirror ol) {
  list-style: decimal;
  padding-left: 1.25rem;
  margin: 0.25rem 0;
}
.step-editor :deep(.ProseMirror ul) {
  list-style: disc;
  padding-left: 1.25rem;
  margin: 0.25rem 0;
}
.step-editor :deep(.ProseMirror li) {
  margin: 0.1rem 0;
}
.step-editor :deep(.ProseMirror li p) {
  margin: 0;
}
.step-editor :deep(.ProseMirror strong) {
  font-weight: 600;
}
.step-editor :deep(.ProseMirror h1) {
  font-size: 1.2rem;
  font-weight: 700;
  margin: 0.5rem 0 0.25rem;
}
.step-editor :deep(.ProseMirror h2) {
  font-size: 1.05rem;
  font-weight: 700;
  margin: 0.4rem 0 0.2rem;
}
.step-editor :deep(.ProseMirror h3) {
  font-size: 0.95rem;
  font-weight: 600;
  margin: 0.3rem 0 0.15rem;
}
.step-editor :deep(.ProseMirror blockquote) {
  border-left: 3px solid #93c5fd;
  padding-left: 0.75rem;
  margin: 0.25rem 0;
  color: #6b7280;
}
.step-editor :deep(.ProseMirror hr) {
  border: none;
  border-top: 1px solid #e5e7eb;
  margin: 0.5rem 0;
}
.step-editor :deep(.ProseMirror code) {
  background: #f3f4f6;
  border-radius: 0.25rem;
  padding: 0.1rem 0.3rem;
  font-size: 0.85em;
}
</style>
