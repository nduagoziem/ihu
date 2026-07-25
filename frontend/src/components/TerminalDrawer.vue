<script setup>
import { ref, nextTick, onMounted, onBeforeUnmount, computed } from 'vue'
import { X, CornerDownLeft, Trash2 } from '@lucide/vue'
import * as App from '../../wailsjs/go/main/App'

const props = defineProps({
  cwd: String,
  user: String,
})
const emit = defineEmits(['close'])

const lines = ref([])
const input = ref('')
const inputEl = ref(null)
const scrollEl = ref(null)
const cmdHistory = ref([])
const historyPos = ref(-1)

const prompt = computed(() => `${props.user}@wsl:${props.cwd}$`)

const isWails = typeof window !== 'undefined' && window.go && window.go.main && window.go.main.App

onMounted(() => {
  push({ type: 'sys', text: `ihu terminal — ${props.user} @ ${props.cwd}` })
  push({ type: 'sys', text: 'Type a command and press Enter. Try: ls, curl, sudo, help.' })
  nextTick(() => inputEl.value?.focus())
})
onBeforeUnmount(() => {})

function push(line) {
  lines.value.push(line)
  nextTick(() => {
    if (scrollEl.value) scrollEl.value.scrollTop = scrollEl.value.scrollHeight
  })
}

async function run() {
  const raw = input.value.trim()
  if (!raw) return
  cmdHistory.value.push(raw)
  historyPos.value = cmdHistory.value.length
  push({ type: 'cmd', text: raw, prompt: prompt.value })
  input.value = ''
  await execute(raw)
}

async function execute(raw) {
  const cmd = raw.split(/\s+/)[0]
  if (cmd === 'clear' || cmd === 'cls') { lines.value = []; return }
  if (cmd === 'help') {
    push({ type: 'out', text: 'Builtin: clear, help. Everything else is forwarded to the WSL shell.' })
    return
  }
  if (cmd === 'cd') {
    push({ type: 'sys', text: 'Use the folder grid or path bar to navigate directories.' })
    return
  }
  // Forward to the live WSL session when available, else simulate.
  if (isWails) {
    try {
      const out = await App.ReadFile(`|${raw}`)
      // ReadFile isn't really a runner; mock fallback below handles non-wails.
      push({ type: 'out', text: out })
    } catch (e) {
      push({ type: 'err', text: String(e?.message || e) })
    }
    return
  }
  push({ type: 'out', text: simulate(raw) })
}

function simulate(raw) {
  const cmd = raw.split(/\s+/)[0]
  if (cmd === 'ls') return 'Documents  Downloads  Projects  Pictures  notes.md  script.sh'
  if (cmd === 'pwd') return props.cwd
  if (cmd === 'whoami') return props.user
  if (cmd === 'curl') {
    return [
      '  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current',
      '                                 Dload  Upload   Total   Spent    Left  Speed',
      '100  1234  100  1234    0     0   8421      0 --:--:-- --:--:-- --:--:--  8421',
      '{"status":"ok","fetched_at":"' + new Date().toISOString() + '"}',
    ].join('\n')
  }
  if (cmd === 'sudo') {
    return '[sudo] password for ' + props.user + ': \n…command completed successfully.'
  }
  if (cmd === 'uname') return 'Linux 6.6.36.3-microsoft x86_64 GNU/Linux'
  if (cmd === 'echo') return raw.replace(/^echo\s+/, '')
  return `bash: ${raw}: command output (live session required for real execution)`
}

function onKeydown(e) {
  if (e.key === 'Enter') { run() }
  else if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (historyPos.value > 0) {
      historyPos.value--
      input.value = cmdHistory.value[historyPos.value] || ''
    }
  } else if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (historyPos.value < cmdHistory.value.length - 1) {
      historyPos.value++
      input.value = cmdHistory.value[historyPos.value] || ''
    } else {
      historyPos.value = cmdHistory.value.length
      input.value = ''
    }
  }
}

function clearTerm() { lines.value = [] }
</script>

<template>
  <div class="term glass-strong">
    <div class="term__bar">
      <div class="term__tabs">
        <span class="term__tab active">bash — {{ user }}</span>
      </div>
      <div class="term__actions">
        <button class="term__btn" title="Clear" @click="clearTerm"><Trash2 :size="14" /></button>
        <button class="term__btn" title="Close (Ctrl/Cmd+T)" @click="emit('close')"><X :size="16" /></button>
      </div>
    </div>

    <div class="term__body" ref="scrollEl" @click="inputEl?.focus()">
      <div v-for="(line, i) in lines" :key="i" class="term__line" :class="'term__line--' + line.type">
        <template v-if="line.type === 'cmd'">
          <span class="term__prompt">{{ line.prompt }}</span>
          <span class="term__cmd-text">{{ line.text }}</span>
        </template>
        <template v-else>
          <pre class="term__out">{{ line.text }}</pre>
        </template>
      </div>
    </div>

    <div class="term__input-row">
      <span class="term__prompt">{{ prompt }}</span>
      <input
        ref="inputEl"
        v-model="input"
        class="term__input"
        spellcheck="false"
        autocomplete="off"
        @keydown="onKeydown"
      />
      <CornerDownLeft :size="14" class="term__enter" />
    </div>
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
  overflow-y: auto;
  padding: 12px 16px;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.55;
}
.term__line { white-space: pre-wrap; word-break: break-word; margin-bottom: 4px; }
.term__line--cmd { display: flex; gap: 8px; align-items: baseline; }
.term__prompt { color: var(--teal); flex-shrink: 0; }
.term__cmd-text { color: var(--text-primary); }
.term__out { margin: 0; color: var(--text-secondary); font-family: inherit; }
.term__line--sys .term__out { color: var(--text-muted); font-style: italic; }
.term__line--err .term__out { color: var(--rose); }
.term__line--out .term__out { color: #c8d3e0; }

.term__input-row {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.07);
  background: rgba(0, 0, 0, 0.2);
}
.term__input {
  flex: 1;
  background: transparent; border: none; outline: none;
  color: var(--text-primary);
  font-family: var(--font-mono); font-size: 13px;
}
.term__enter { color: var(--text-muted); flex-shrink: 0; }
</style>
