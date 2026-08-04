<script setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { X, Trash2, RotateCw } from '@lucide/vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import * as App from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const props = defineProps({
  cwd: String,
  user: String,
  distro: String,
  superUser: Boolean,
})
const emit = defineEmits(['close'])

const termEl = ref(null)
const exited = ref(false)

let term = null
let fit = null
let resizeObserver = null
let disposed = false

const effectiveUser = computed(() => (props.superUser || props.user === 'root') ? 'root' : props.user)

// xterm theme aligned with the app's design tokens (see style.css).
const theme = {
  background: 'rgba(0, 0, 0, 0)',
  foreground: '#c8d3e0',
  cursor: '#2dd4bf',
  cursorAccent: '#0b0f17',
  selectionBackground: 'rgba(45, 212, 191, 0.28)',
  black: '#1b2130', red: '#fb7185', green: '#2dd4bf', yellow: '#e5c07b',
  blue: '#61afef', magenta: '#c678dd', cyan: '#56b6c2', white: '#c8d3e0',
  brightBlack: '#5c6672', brightRed: '#fb7185', brightGreen: '#2dd4bf',
  brightYellow: '#e5c07b', brightBlue: '#61afef', brightMagenta: '#c678dd',
  brightCyan: '#56b6c2', brightWhite: '#f2f5fa',
}

onMounted(async () => {
  term = new Terminal({
    fontFamily: '"SF Mono", "JetBrains Mono", "Fira Code", ui-monospace, "Cascadia Code", Menlo, monospace',
    fontSize: 13,
    lineHeight: 1.2,
    cursorBlink: true,
    allowProposedApi: true,
    theme,
  })
  fit = new FitAddon()
  term.loadAddon(fit)
  term.open(termEl.value)
  fit.fit()

  // Keystrokes (including Ctrl-C as 0x03) flow straight to the pseudoconsole.
  term.onData((d) => { App.TerminalWrite(d).catch(() => {}) })

  // Pseudoconsole output — raw ANSI/VT bytes — written verbatim to xterm.
  EventsOn('terminal:data', (chunk) => { term?.write(chunk) })
  EventsOn('terminal:exit', () => {
    exited.value = true
    term?.write('\r\n\x1b[2m[process exited]\x1b[0m\r\n')
  })

  await startSession()

  resizeObserver = new ResizeObserver(() => doFit())
  resizeObserver.observe(termEl.value)

  term.focus()
})

onBeforeUnmount(() => {
  disposed = true
  EventsOff('terminal:data')
  EventsOff('terminal:exit')
  resizeObserver?.disconnect()
  App.TerminalStop().catch(() => {})
  term?.dispose()
  term = null
})

async function startSession() {
  exited.value = false
  const { cols, rows } = term
  try {
    await App.TerminalStart(
      props.distro || '',
      props.user || '',
      props.cwd || '',
      props.superUser || props.user === 'root',
      cols,
      rows,
    )
  } catch (e) {
    term?.write(`\r\n\x1b[31m${String(e?.message || e)}\x1b[0m\r\n`)
    exited.value = true
  }
}

function doFit() {
  if (disposed || !fit || !term) return
  try {
    fit.fit()
    App.TerminalResize(term.cols, term.rows).catch(() => {})
  } catch { /* element not measurable yet */ }
}

async function restart() {
  await App.TerminalStop().catch(() => {})
  term?.reset()
  await startSession()
  term?.focus()
}

function clearTerm() {
  term?.clear()
  term?.focus()
}
</script>

<template>
  <div class="term glass-strong">
    <div class="term__bar">
      <div class="term__tabs">
        <span class="term__tab active">bash - {{ effectiveUser }}</span>
      </div>
      <div class="term__actions">
        <button v-if="exited" class="term__btn" title="Restart session" @click="restart"><RotateCw :size="14" /></button>
        <button class="term__btn" title="Clear" @click="clearTerm"><Trash2 :size="14" /></button>
        <button class="term__btn" title="Close (Ctrl/Cmd+T)" @click="emit('close')"><X :size="16" /></button>
      </div>
    </div>

    <div class="term__body" ref="termEl"></div>
  </div>
</template>

<style scoped>
.term {
  position: fixed;
  left: 12px; right: 12px; bottom: 12px;
  height: 300px;
  border-radius: var(--r-lg);
  display: flex; flex-direction: column;
  z-index: 60;
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}
.term__bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 12px;
  height: 36px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  flex-shrink: 0;
}
.term__tabs { display: flex; gap: 6px; }
.term__tab {
  font-size: 12px;
  color: var(--text-secondary);
  padding: 4px 10px;
  border-radius: var(--r-sm);
  background: rgba(255, 255, 255, 0.06);
}
.term__tab.active { color: var(--text-primary); background: var(--info-100); }
.term__actions { display: flex; gap: 4px; }
.term__btn {
  width: 28px; height: 28px; border-radius: var(--r-sm); border: none;
  background: transparent; color: var(--text-muted);
  display: grid; place-items: center;
  transition: all var(--t-fast) var(--ease);
}
.term__btn:hover { background: rgba(255, 255, 255, 0.08); color: var(--text-primary); }

.term__body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  padding: 8px 10px 4px;
}
/* Let the xterm viewport blend with the glass drawer instead of its own bg. */
.term__body :deep(.xterm),
.term__body :deep(.xterm-viewport) {
  background: transparent !important;
}
.term__body :deep(.xterm-viewport) {
  scrollbar-width: thin;
}
</style>
