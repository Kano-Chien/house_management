<template>
  <div class="step-editor select-none flex flex-col">
    <!-- Toolbar -->
    <div class="flex items-center gap-2 p-2 border-b border-blue-100 bg-white/50 rounded-t-lg">
      <button @click="triggerImageUpload" title="Insert Image"
        class="p-1.5 text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors flex items-center justify-center">
        <span class="text-lg leading-none">🖼️</span>
      </button>
      <input type="file" ref="fileInput" accept="image/*" class="hidden" @change="handleFileUpload" />
      <span v-if="uploading" class="text-xs text-blue-500 font-medium ml-2 animate-pulse">Uploading...</span>
    </div>
    
    <!-- Editor Area -->
    <div @click="focusEditor" class="flex-1 cursor-text" @drop.prevent="handleDrop" @dragover.prevent>
      <editor-content :editor="editor" />
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onBeforeUnmount } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import Image from '@tiptap/extension-image'

const props = defineProps({
  modelValue: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue', 'blur'])

const fileInput = ref(null)
const uploading = ref(false)

const editor = useEditor({
  content: props.modelValue || '',
  extensions: [
    StarterKit.configure({
      heading: { levels: [1, 2, 3] },
    }),
    Placeholder.configure({
      placeholder: 'Type your recipe steps here... Use markdown shortcuts like "1. " or "- ". Drag and drop images.',
    }),
    Image.configure({
      HTMLAttributes: {
        class: 'rounded-lg max-w-full h-auto shadow-sm my-4 border border-gray-100',
      },
    }),
  ],
  editorProps: {
    attributes: {
      class: 'prose-editor focus:outline-none min-h-[100px] p-4 text-sm text-gray-700',
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

// ---- Image Upload Handling ----
const triggerImageUpload = () => {
  if (fileInput.value) fileInput.value.click()
}

const uploadToServer = async (file) => {
  if (!file || !file.type.startsWith('image/')) return null
  
  const formData = new FormData()
  formData.append('image', file)
  
  try {
    uploading.value = true
    const res = await fetch('/api/upload', {
      method: 'POST',
      body: formData
    })
    
    if (res.ok) {
      const data = await res.json()
      return data.url
    }
    return null
  } catch (e) {
    console.error('Image upload failed', e)
    return null
  } finally {
    uploading.value = false
  }
}

const handleFileUpload = async (e) => {
  const file = e.target.files[0]
  if (!file) return
  
  const url = await uploadToServer(file)
  if (url && editor.value) {
    editor.value.chain().focus().setImage({ src: url }).run()
  }
  
  // Reset input
  if (fileInput.value) fileInput.value.value = ''
}

const handleDrop = async (e) => {
  e.preventDefault()
  if (!e.dataTransfer || !e.dataTransfer.files.length) return
  
  const file = e.dataTransfer.files[0]
  const url = await uploadToServer(file)
  if (url && editor.value) {
    editor.value.chain().focus().setImage({ src: url }).run()
  }
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
  min-height: 100px;
  outline: none;
}
.step-editor :deep(.ProseMirror img) {
  display: block;
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
  margin: 1rem auto;
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
