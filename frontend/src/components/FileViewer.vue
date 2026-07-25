<script setup>
import { ref, watch, onMounted, computed } from 'vue'
import { X, FileText, Loader2 } from '@lucide/vue'
import * as App from '../../wailsjs/go/main/App'

const props = defineProps({ file: Object })
const emit = defineEmits(['close'])

const loading = ref(true)
const error = ref('')
const kind = ref('text')
const imageUrl = ref('')
const pdfPages = ref([])
const docHtml = ref('')
const textContent = ref('')

const ext = computed(() => {
  const n = (props.file?.name || '').toLowerCase()
  const m = n.match(/\.([a-z0-9]+)$/)
  return m ? m[1] : ''
})

onMounted(load)
watch(() => props.file?.path, load)

async function load() {
  if (!props.file) return
  loading.value = true
  error.value = ''
  kind.value = 'text'
  imageUrl.value = ''
  pdfPages.value = []
  docHtml.value = ''
  textContent.value = ''

  const e = ext.value
  try {
    if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp'].includes(e)) {
      kind.value = 'image'
      const txt = await App.ReadFile(props.file.path).catch(() => '')
      if (txt && /^data:/.test(txt)) {
        imageUrl.value = txt
      } else {
        imageUrl.value = ''
        error.value = 'Preview not available in this build. Open in an external viewer.'
      }
    } else if (e === 'pdf') {
      kind.value = 'pdf'
      await loadPdf()
    } else if (e === 'docx') {
      kind.value = 'docx'
      await loadDocx()
    } else {
      kind.value = 'text'
      textContent.value = await App.ReadFile(props.file.path)
    }
  } catch (err) {
    error.value = String(err?.message || err || 'Failed to open file')
  } finally {
    loading.value = false
  }
}

async function loadPdf() {
  const txt = await App.ReadFile(props.file.path).catch(() => '')
  if (!txt) { error.value = 'Could not read PDF.'; return }
  const { default: pdfjsLib } = await import('pdfjs-dist')
  let data
  try {
    data = Uint8Array.from(atob(txt), (c) => c.charCodeAt(0))
  } catch {
    textContent.value = txt
    kind.value = 'text'
    return
  }
  pdfjsLib.GlobalWorkerOptions.workerSrc = new URL('pdfjs-dist/build/pdf.worker.min.mjs', import.meta.url).toString()
  const pdf = await pdfjsLib.getDocument({ data }).promise
  const pages = []
  for (let i = 1; i <= pdf.numPages; i++) {
    const page = await pdf.getPage(i)
    const vc = await page.getViewport({ scale: 1.3 })
    const canvas = document.createElement('canvas')
    canvas.width = vc.width
    canvas.height = vc.height
    await page.render({ canvasContext: canvas.getContext('2d'), viewport: vc }).promise
    pages.push(canvas.toDataURL('image/png'))
  }
  pdfPages.value = pages
}

async function loadDocx() {
  const [{ default: mammoth }, txt] = await Promise.all([
    import('mammoth'),
    App.ReadFile(props.file.path).catch(() => ''),
  ])
  if (!txt) { error.value = 'Could not read document.'; return }
  let arrayBuffer
  try {
    const bytes = Uint8Array.from(atob(txt), (c) => c.charCodeAt(0))
    arrayBuffer = bytes.buffer
  } catch {
    textContent.value = txt
    kind.value = 'text'
    return
  }
  const result = await mammoth.convertToHtml({ arrayBuffer })
  docHtml.value = result.value
}
</script>

<template>
  <div class="viewer-overlay" @click.self="emit('close')">
    <div class="viewer">
      <div class="viewer__bar">
        <div class="viewer__file">
          <FileText :size="15" />
          <span>{{ file?.name || 'Preview' }}</span>
        </div>
        <button class="viewer__btn" @click="emit('close')"><X :size="18" /></button>
      </div>
      <div class="viewer__body">
        <div v-if="loading" class="viewer__loading">
          <Loader2 :size="26" class="spin" />
          <p>Loading…</p>
        </div>
        <div v-else-if="error" class="viewer__error">
          <p>{{ error }}</p>
        </div>
        <template v-else>
          <div v-if="kind === 'image'" class="viewer__image">
            <img v-if="imageUrl" :src="imageUrl" :alt="file?.name" />
            <p v-else class="viewer__placeholder">Image preview unavailable</p>
          </div>
          <div v-else-if="kind === 'pdf'" class="viewer__pdf">
            <img v-for="(p, i) in pdfPages" :key="i" :src="p" :alt="`Page ${i + 1}`" />
          </div>
          <div v-else-if="kind === 'docx'" class="viewer__docx" v-html="docHtml"></div>
          <pre v-else class="viewer__text">{{ textContent }}</pre>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.viewer-overlay {
  position: fixed; inset: 0; z-index: 78;
  display: grid; place-items: center;
  background: rgba(5, 6, 8, 0.55);
  animation: fadeIn var(--t-med) var(--ease);
}
.viewer {
  width: min(840px, 94vw);
  height: min(680px, 88vh);
  border-radius: var(--r-xl);
  display: flex; flex-direction: column;
  overflow: hidden;
  background: var(--surface-2);
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-slow) var(--ease-out);
}
.viewer__bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 14px; height: 44px;
  background: var(--surface-1);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}
.viewer__file { display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--text-secondary); }
.viewer__btn {
  width: 32px; height: 32px; border-radius: var(--r-sm); border: none;
  background: transparent; color: var(--text-muted);
  display: grid; place-items: center;
}
.viewer__btn:hover { background: rgba(255,255,255,0.08); color: var(--text-primary); }

.viewer__body { flex: 1; overflow: auto; display: grid; place-items: center; }
.viewer__loading, .viewer__error { display: grid; place-items: center; gap: 12px; color: var(--text-muted); }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.viewer__image { padding: 24px; max-width: 100%; max-height: 100%; display: grid; place-items: center; }
.viewer__image img { max-width: 100%; max-height: 70vh; border-radius: var(--r-md); box-shadow: var(--shadow-md); }
.viewer__placeholder { color: var(--text-muted); }

.viewer__pdf { padding: 20px; display: flex; flex-direction: column; align-items: center; gap: 16px; }
.viewer__pdf img { max-width: 100%; border-radius: var(--r-sm); box-shadow: var(--shadow-md); }

.viewer__docx { padding: 32px 40px; max-width: 720px; color: var(--text-primary); line-height: 1.7; }
.viewer__docx :deep(h1), .viewer__docx :deep(h2), .viewer__docx :deep(h3) { margin: 1.2em 0 0.5em; color: var(--text-primary); }
.viewer__docx :deep(p) { margin: 0.6em 0; }
.viewer__docx :deep(strong) { color: var(--text-primary); }
.viewer__docx :deep(ul), .viewer__docx :deep(ol) { padding-left: 24px; }

.viewer__text {
  margin: 0; padding: 24px;
  font-family: var(--font-mono); font-size: 13px; line-height: 1.6;
  color: var(--text-secondary);
  white-space: pre-wrap; word-break: break-word;
  text-align: left; width: 100%;
}
</style>
