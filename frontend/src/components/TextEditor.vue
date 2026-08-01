<script setup>
import { ref, watch, onMounted, computed } from 'vue'
import { X, Save, FileText } from '@lucide/vue'
import * as App from '../../wailsjs/go/main/App'
import hljs from 'highlight.js'

const props = defineProps({
  file: Object,
  currentUser: String,
  currentDistro: String,
  superUser: Boolean,
})
const emit = defineEmits(['close'])

const content = ref('')
const loading = ref(true)
const error = ref('')
const dirty = ref(false)
const edited = ref('')
const saving = ref(false)
const highlightEl = ref(null)

const lang = computed(() => {
  const n = (props.file?.name || '').toLowerCase()
  if (n.endsWith('.sh') || n.endsWith('.bashrc') || n.endsWith('.profile')) return 'bash'
  if (n.endsWith('.go')) return 'go'
  if (n.endsWith('.js') || n.endsWith('.ts')) return 'javascript'
  if (n.endsWith('.md')) return 'markdown'
  if (n.endsWith('.json')) return 'json'
  return 'bash'
})

const highlighted = computed(() => {
  if (dirty.value) return escapeHtml(edited.value)
  try {
    if (hljs.getLanguage(lang.value)) {
      return hljs.highlight(content.value, { language: lang.value, ignoreIllegals: true }).value
    }
    return hljs.highlightAuto(content.value).value
  } catch {
    return escapeHtml(content.value)
  }
})

function escapeHtml(s) {
  return (s || '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

onMounted(load)
watch(() => [props.file?.path, props.currentUser, props.currentDistro, props.superUser], load)

async function load() {
  if (!props.file) return
  loading.value = true
  error.value = ''
  try {
    content.value = await App.ReadFileAs(props.file.path, props.currentDistro || '', props.currentUser || '', isElevated())
    edited.value = content.value
    dirty.value = false
  } catch (e) {
    error.value = String(e?.message || e)
    content.value = `// Unable to read ${props.file.name}\n// ${error.value}`
  } finally {
    loading.value = false
  }
}

function onInput(e) {
  edited.value = e.target.value
  dirty.value = edited.value !== content.value
}
function onScroll(e) {
  if (!highlightEl.value) return
  highlightEl.value.scrollTop = e.target.scrollTop
  highlightEl.value.scrollLeft = e.target.scrollLeft
}
async function save() {
  if (!props.file || saving.value) return
  saving.value = true
  error.value = ''
  try {
    await App.WriteFileAs(props.file.path, edited.value, props.currentDistro || '', props.currentUser || '', isElevated())
    content.value = edited.value
    dirty.value = false
  } catch (e) {
    error.value = String(e?.message || e || 'Failed to save file')
  } finally {
    saving.value = false
  }
}
function isElevated() {
  return props.superUser || props.currentUser === 'root'
}
</script>

<template>
  <div class="editor-overlay" @click.self="emit('close')">
    <div class="editor">
      <div class="editor__bar">
        <div class="editor__file">
          <FileText :size="15" />
          <span>{{ file?.name || 'Untitled' }}</span>
          <span v-if="dirty" class="editor__dot"></span>
        </div>
        <div class="editor__actions">
          <button class="editor__btn" :disabled="!dirty || saving" @click="save"><Save :size="14" /> {{ saving ? 'Saving' : 'Save' }}</button>
          <button class="editor__btn" @click="emit('close')"><X :size="16" /></button>
        </div>
      </div>
      <div class="editor__body">
        <pre ref="highlightEl" class="editor__highlight" v-html="highlighted"></pre>
        <textarea
          class="editor__textarea"
          :value="dirty ? edited : content"
          spellcheck="false"
          @input="onInput"
          @scroll="onScroll"
          @keydown.tab.prevent="$event.target.setRangeText('  ', $event.target.selectionStart, $event.target.selectionEnd, 'end')"
        ></textarea>
      </div>
      <div class="editor__status">
        <span>{{ lang }}</span>
        <span>{{ (dirty ? edited : content).split('\n').length }} lines</span>
        <span v-if="dirty">Unsaved changes</span>
        <span v-if="error" class="editor__error">{{ error }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.editor-overlay {
  position: fixed; inset: 0; z-index: 75;
  display: grid; place-items: center;
  background: rgba(5, 6, 8, 0.5);
  animation: fadeIn var(--t-med) var(--ease);
}
.editor {
  width: min(900px, 94vw);
  height: min(640px, 86vh);
  border-radius: var(--r-xl);
  display: flex; flex-direction: column;
  overflow: hidden;
  background: rgba(24, 28, 36, 0.86);
  backdrop-filter: blur(36px) saturate(170%);
  -webkit-backdrop-filter: blur(36px) saturate(170%);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-slow) var(--ease-out);
}
.editor__bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 14px; height: 44px;
  background: rgba(12, 15, 20, 0.54);
  border-bottom: 1px solid #181a1f;
}
.editor__file { display: flex; align-items: center; gap: 8px; font-size: 13px; color: #abb2bf; }
.editor__dot { width: 7px; height: 7px; border-radius: 50%; background: #e5c07b; }
.editor__actions { display: flex; gap: 4px; align-items: center; }
.editor__btn {
  display: flex; align-items: center; gap: 6px;
  height: 30px; padding: 0 12px;
  border-radius: var(--r-sm); border: none;
  background: rgba(255,255,255,0.06); color: #abb2bf;
  font-size: 12px;
  transition: all var(--t-fast) var(--ease);
}
.editor__btn:hover:not(:disabled) { background: rgba(255,255,255,0.12); color: #fff; }
.editor__btn:disabled { opacity: 0.4; cursor: not-allowed; }

.editor__body { flex: 1; position: relative; overflow: hidden; }
.editor__highlight,
.editor__textarea {
  position: absolute; inset: 0;
  margin: 0; padding: 18px 22px;
  font-family: var(--font-mono); font-size: 13.5px; line-height: 1.6;
  white-space: pre-wrap; word-break: break-word;
  overflow: auto;
  tab-size: 2;
}
.editor__highlight { color: #abb2bf; pointer-events: none; }
.editor__highlight::-webkit-scrollbar { width: 0; height: 0; }
.editor__highlight :deep(.hljs-keyword) { color: #c678dd; }
.editor__highlight :deep(.hljs-string) { color: #98c379; }
.editor__highlight :deep(.hljs-comment) { color: #5c6370; font-style: italic; }
.editor__highlight :deep(.hljs-number) { color: #d19a66; }
.editor__highlight :deep(.hljs-function) { color: #61afef; }
.editor__highlight :deep(.hljs-title) { color: #61afef; }
.editor__highlight :deep(.hljs-built_in) { color: #e5c07b; }
.editor__highlight :deep(.hljs-attr) { color: #d19a66; }
.editor__highlight :deep(.hljs-variable) { color: #e06c75; }

.editor__textarea {
  background: transparent; color: transparent;
  caret-color: #528bff; border: none; outline: none; resize: none;
  -webkit-text-fill-color: transparent;
}

.editor__status {
  display: flex; gap: 16px;
  padding: 8px 22px;
  background: rgba(12, 15, 20, 0.54);
  border-top: 1px solid #181a1f;
  font-size: 11px; color: #5c6370;
}
.editor__error { color: #e06c75; }
</style>
