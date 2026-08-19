<script setup>
import { ref, computed, nextTick, onBeforeUnmount } from 'vue'
import { HardDriveDownload, TriangleAlert, LoaderCircle, CircleCheck, X, ShieldAlert, RotateCw, Power } from '@lucide/vue'
import * as App from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const props = defineProps({
  distro: String,
})
// `busy` tells the parent to lock the rest of the app while reclaiming runs.
const emit = defineEmits(['close', 'busy'])

// phase: 'confirm' | 'running' | 'done' | 'error'
const phase = ref('confirm')
const log = ref([])
const result = ref(null)
const errorMsg = ref('')
const logEl = ref(null)

const isRunning = computed(() => phase.value === 'running')
const distroLabel = computed(() => props.distro || 'default')

const PROGRESS_EVENT = 'janitor:progress'

async function pushLog(line) {
  log.value.push(line)
  await nextTick()
  if (logEl.value) logEl.value.scrollTop = logEl.value.scrollHeight
}

async function start() {
  phase.value = 'running'
  log.value = []
  result.value = null
  errorMsg.value = ''
  emit('busy', true)

  EventsOn(PROGRESS_EVENT, (msg) => { pushLog(msg) })

  try {
    const res = await App.ReclaimSpace(props.distro || '')
    result.value = res
    phase.value = 'done'
  } catch (e) {
    errorMsg.value = String(e?.message || e || 'unknown error')
    phase.value = 'error'
  } finally {
    EventsOff(PROGRESS_EVENT)
    emit('busy', false)
  }
}

function requestClose() {
  // Never allow dismissal mid-operation.
  if (isRunning.value) return
  emit('close')
}

function retry() {
  phase.value = 'confirm'
}

onBeforeUnmount(() => {
  EventsOff(PROGRESS_EVENT)
})

function humanBytes(n) {
  if (!n || n <= 0) return '0 B'
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
  let i = 0
  let v = n
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++ }
  return `${i === 0 ? v : v.toFixed(2)} ${units[i]}`
}
</script>

<template>
  <div class="reclaim-modal" @click.self="requestClose">
    <div class="reclaim-card glass-strong" @click.stop>
      <div class="reclaim-card__decor"></div>

      <button v-if="!isRunning" class="reclaim-card__close" title="Close" @click="requestClose">
        <X :size="18" />
      </button>

      <div class="reclaim-card__content">
        <div class="reclaim-card__brand">
          <div class="reclaim-card__logo">
            <HardDriveDownload :size="26" />
          </div>
          <div>
            <h1>Reclaim Disk Space</h1>
            <p class="reclaim-card__sub">Shrink the WSL virtual disk and return unused space to Windows</p>
          </div>
        </div>

        <!-- CONFIRM -->
        <template v-if="phase === 'confirm'">
          <p class="reclaim-card__lead">
            This trims unused blocks inside <strong>{{ distroLabel }}</strong> and compacts its
            <code>ext4.vhdx</code> so Windows gets the freed space back. Before you continue, note:
          </p>

          <ul class="reclaim-notes">
            <li>
              <RotateCw :size="16" class="reclaim-notes__icon" />
              <span>Your background <strong>WSL session will be shut down and restarted</strong>. Any unsaved work in the terminal will be lost.</span>
            </li>
            <li>
              <ShieldAlert :size="16" class="reclaim-notes__icon reclaim-notes__icon--warn" />
              <span>A <strong>Windows admin (UAC) prompt</strong> will appear — compaction needs elevated rights. Approve it to proceed.</span>
            </li>
            <li>
              <Power :size="16" class="reclaim-notes__icon" />
              <span>All running distributions are stopped via <code>wsl --shutdown</code>.</span>
            </li>
            <li>
              <TriangleAlert :size="16" class="reclaim-notes__icon reclaim-notes__icon--warn" />
              <span><strong>Don't close the app</strong> until it finishes. The rest of the app is locked while it runs.</span>
            </li>
          </ul>

          <div class="reclaim-card__actions">
            <button class="btn btn--ghost" @click="requestClose">Cancel</button>
            <button class="btn btn--primary" @click="start">
              <HardDriveDownload :size="16" />
              Reclaim space
            </button>
          </div>
        </template>

        <!-- RUNNING -->
        <template v-else-if="phase === 'running'">
          <div class="reclaim-status">
            <LoaderCircle :size="18" class="spin" />
            <span>Working… please wait. Do not close the app.</span>
          </div>
          <div ref="logEl" class="reclaim-log">
            <div v-for="(line, i) in log" :key="i" class="reclaim-log__line">{{ line }}</div>
            <div v-if="!log.length" class="reclaim-log__line reclaim-log__line--muted">Starting…</div>
          </div>
          <p class="reclaim-card__hint">Approve the Windows admin prompt if it appears.</p>
        </template>

        <!-- DONE -->
        <template v-else-if="phase === 'done'">
          <div class="reclaim-result">
            <CircleCheck :size="40" class="reclaim-result__icon" />
            <div class="reclaim-result__headline">Reclaimed {{ humanBytes(result?.savedBytes) }}</div>
          </div>
          <div class="reclaim-figures">
            <div class="figure">
              <div class="figure__label">Before</div>
              <div class="figure__val">{{ humanBytes(result?.beforeBytes) }}</div>
            </div>
            <div class="figure">
              <div class="figure__label">After</div>
              <div class="figure__val">{{ humanBytes(result?.afterBytes) }}</div>
            </div>
            <div class="figure figure--accent">
              <div class="figure__label">Saved</div>
              <div class="figure__val">{{ humanBytes(result?.savedBytes) }}</div>
            </div>
          </div>
          <p v-if="result?.vhdxPath" class="reclaim-card__path" :title="result.vhdxPath">{{ result.vhdxPath }}</p>
          <div class="reclaim-card__actions">
            <button class="btn btn--primary" @click="requestClose">Done</button>
          </div>
        </template>

        <!-- ERROR -->
        <template v-else>
          <div class="reclaim-status reclaim-status--error">
            <TriangleAlert :size="18" />
            <span>Reclaim failed</span>
          </div>
          <div class="reclaim-error">{{ errorMsg }}</div>
          <div class="reclaim-card__actions">
            <button class="btn btn--ghost" @click="requestClose">Close</button>
            <button class="btn btn--primary" @click="retry">
              <RotateCw :size="16" />
              Try again
            </button>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reclaim-modal {
  position: fixed;
  inset: 0;
  z-index: 110;
  display: grid;
  place-items: center;
  animation: fadeIn var(--t-med) var(--ease);
  background: rgba(5, 6, 8, 0.45);
}
.reclaim-card {
  position: relative;
  width: min(560px, 92vw);
  border-radius: var(--r-xl);
  padding: 32px 30px 26px;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-slow) var(--ease-out);
}
.reclaim-card__decor {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background:
    radial-gradient(circle at 0% 0%, rgba(30, 111, 235, 0.4), transparent 34%),
    radial-gradient(circle at 100% 100%, rgba(45, 212, 191, 0.26), transparent 32%);
}
.reclaim-card__close {
  position: absolute;
  top: 14px;
  right: 14px;
  width: 30px;
  height: 30px;
  border-radius: var(--r-full);
  border: none;
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-secondary);
  display: grid;
  place-items: center;
  z-index: 3;
  transition: all var(--t-fast) var(--ease);
}
.reclaim-card__close:hover { background: rgba(255, 255, 255, 0.16); color: var(--text-primary); }
.reclaim-card__content { position: relative; z-index: 2; }

.reclaim-card__brand {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
}
.reclaim-card__logo {
  width: 50px;
  height: 50px;
  border-radius: 14px;
  background: linear-gradient(135deg, #3b82f6, #2dd4bf);
  display: grid;
  place-items: center;
  color: #fff;
  box-shadow: 0 8px 22px rgba(59, 130, 246, 0.45);
  flex-shrink: 0;
}
.reclaim-card__brand h1 { margin: 0; font-size: 22px; font-weight: 700; }
.reclaim-card__sub { margin: 3px 0 0; color: var(--text-secondary); font-size: 12.5px; }

.reclaim-card__lead { font-size: 13.5px; color: var(--text-secondary); line-height: 1.5; margin: 0 0 16px; }
.reclaim-card__lead code,
.reclaim-card__path code { font-family: var(--font-mono); font-size: 12px; color: var(--text-primary); }

.reclaim-notes { list-style: none; margin: 0 0 22px; padding: 0; display: flex; flex-direction: column; gap: 12px; }
.reclaim-notes li {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.45;
}
.reclaim-notes li code { font-family: var(--font-mono); font-size: 11.5px; color: var(--text-primary); }
.reclaim-notes__icon { color: var(--accent-hover); margin-top: 1px; flex-shrink: 0; }
.reclaim-notes__icon--warn { color: var(--warning); }

.reclaim-card__actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 4px; }
.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 9px 20px;
  border-radius: var(--r-md);
  border: none;
  font-weight: 600;
  font-size: 13px;
  transition: all var(--t-fast) var(--ease);
}
.btn--primary { background: var(--accent); color: #fff; }
.btn--primary:hover { background: var(--accent-hover); transform: translateY(-1px); }
.btn--ghost { background: rgba(255, 255, 255, 0.06); color: var(--text-secondary); }
.btn--ghost:hover { background: rgba(255, 255, 255, 0.12); color: var(--text-primary); }

.reclaim-status {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13.5px;
  color: var(--text-primary);
  margin-bottom: 14px;
}
.reclaim-status--error { color: var(--error); }
.reclaim-status--error :deep(svg) { color: var(--error); }

.reclaim-log {
  height: 190px;
  overflow-y: auto;
  border-radius: var(--r-md);
  background: rgba(0, 0, 0, 0.32);
  border: 1px solid rgba(255, 255, 255, 0.06);
  padding: 12px 14px;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.55;
  color: var(--text-secondary);
}
.reclaim-log__line { white-space: pre-wrap; word-break: break-word; }
.reclaim-log__line--muted { color: var(--text-muted); }
.reclaim-card__hint { font-size: 12px; color: var(--text-muted); margin: 12px 0 0; }

.reclaim-result { display: flex; flex-direction: column; align-items: center; gap: 8px; margin: 8px 0 22px; }
.reclaim-result__icon { color: var(--success); }
.reclaim-result__headline { font-size: 22px; font-weight: 700; }
.reclaim-figures { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin-bottom: 16px; }
.figure {
  padding: 12px;
  border-radius: var(--r-md);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  text-align: center;
}
.figure--accent { border-color: var(--info-300); background: var(--info-50); }
.figure__label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.06em; margin-bottom: 6px; }
.figure__val { font-size: 15px; font-weight: 600; }
.reclaim-card__path {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-muted);
  margin: 0 0 16px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reclaim-error {
  border-radius: var(--r-md);
  background: rgba(251, 113, 133, 0.08);
  border: 1px solid rgba(251, 113, 133, 0.3);
  padding: 12px 14px;
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-primary);
  max-height: 180px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
  margin-bottom: 18px;
}

.spin { animation: spin 0.9s linear infinite; color: var(--accent-hover); }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
