<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { Cpu, MemoryStick, HardDrive, Thermometer, Wifi, X, Sparkles } from '@lucide/vue'

const props = defineProps({
  bootData: { type: Object, default: null },
})
const emit = defineEmits(['close', 'disable'])

const progress = ref(0)
const autoDismissed = ref(false)
let timer = null
let progressTimer = null

const stats = computed(() => props.bootData?.systemStats || {})

onMounted(() => {
  progressTimer = setInterval(() => {
    progress.value = Math.min(100, progress.value + 100 / 40)
  }, 100)
  timer = setTimeout(() => {
    if (!autoDismissed.value) dismiss()
  }, 4000)
})
onBeforeUnmount(() => {
  clearTimeout(timer)
  clearInterval(progressTimer)
})

function dismiss() {
  autoDismissed.value = true
  emit('close')
}
function disable() {
  emit('disable', true)
}
</script>

<template>
  <div class="welcome" @click="dismiss">
    <div class="welcome__frost glass-strong" @click.stop>
      <div class="welcome__decor">
        <div class="orb orb--1"></div>
        <div class="orb orb--2"></div>
        <div class="orb orb--3"></div>
      </div>

      <button class="welcome__close" title="Close (click anywhere)" @click="dismiss">
        <X :size="18" />
      </button>

      <div class="welcome__content">
        <div class="welcome__brand">
          <div class="welcome__logo">
            <Sparkles :size="28" />
          </div>
          <div>
            <h1>ihu</h1>
            <p class="welcome__sub">Your WSL2 environment, beautifully tamed</p>
          </div>
        </div>

        <div class="welcome__distro" v-if="stats.distro">
          <span class="dot"></span>
          {{ stats.distro }}
          <span class="muted">· {{ stats.kernel }}</span>
        </div>

        <div class="welcome__grid">
          <div class="stat">
            <Cpu :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">CPU</div>
              <div class="stat__bar"><div class="stat__fill" :style="{ width: (stats.cpuUsage || 0) + '%' }"></div></div>
              <div class="stat__val">{{ stats.cpuUsage || 0 }}%</div>
            </div>
          </div>
          <div class="stat">
            <MemoryStick :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">Memory</div>
              <div class="stat__bar"><div class="stat__fill stat__fill--teal" :style="{ width: (stats.memoryUsage || 0) + '%' }"></div></div>
              <div class="stat__val">{{ stats.memoryUsage || 0 }}%</div>
            </div>
          </div>
          <div class="stat">
            <HardDrive :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">Disk</div>
              <div class="stat__bar"><div class="stat__fill stat__fill--amber" :style="{ width: (stats.diskUsage || 0) + '%' }"></div></div>
              <div class="stat__val">{{ stats.diskUsage || 0 }}%</div>
            </div>
          </div>
          <div class="stat">
            <Thermometer :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">Temperature</div>
              <div class="stat__val stat__val--big">{{ stats.temperature || 0 }}°C</div>
            </div>
          </div>
          <div class="stat">
            <Wifi :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">Network</div>
              <div class="stat__val" :class="stats.networkStatus === 'active' ? 'ok' : 'warn'">
                {{ stats.networkStatus || 'inactive' }}
              </div>
            </div>
          </div>
          <div class="stat">
            <div class="stat__body">
              <div class="stat__label">Architecture</div>
              <div class="stat__val stat__val--mono">{{ stats.arch || '—' }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="welcome__footer">
        <button class="link" @click="disable">Don't show at startup</button>
        <div class="welcome__progress"><div class="welcome__progress-fill" :style="{ width: progress + '%' }"></div></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.welcome {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: grid;
  place-items: center;
  animation: fadeIn var(--t-med) var(--ease);
  background: rgba(5, 6, 8, 0.35);
}
.welcome__frost {
  position: relative;
  width: min(560px, 92vw);
  border-radius: var(--r-xl);
  padding: 36px 32px 22px;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-slow) var(--ease-out);
}
.welcome__decor {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
  z-index: 0;
}
.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(50px);
  opacity: 0.55;
}
.orb--1 { width: 220px; height: 220px; background: #1e6feb; top: -70px; left: -60px; }
.orb--2 { width: 180px; height: 180px; background: #2dd4bf; bottom: -60px; right: -40px; opacity: 0.4; }
.orb--3 { width: 140px; height: 140px; background: #fbbf24; top: 40%; right: 30%; opacity: 0.22; }

.welcome__close {
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
.welcome__close:hover { background: rgba(255, 255, 255, 0.16); color: var(--text-primary); }

.welcome__content { position: relative; z-index: 2; }

.welcome__brand { display: flex; align-items: center; gap: 14px; margin-bottom: 20px; }
.welcome__logo {
  width: 52px; height: 52px;
  border-radius: 14px;
  background: linear-gradient(135deg, #3b82f6, #2dd4bf);
  display: grid; place-items: center;
  color: #fff;
  box-shadow: 0 8px 22px rgba(59, 130, 246, 0.45);
}
.welcome__brand h1 { margin: 0; font-size: 28px; font-weight: 700; letter-spacing: -0.02em; }
.welcome__sub { margin: 2px 0 0; color: var(--text-secondary); font-size: 13px; }

.welcome__distro {
  display: inline-flex; align-items: center; gap: 8px;
  padding: 6px 14px;
  border-radius: var(--r-full);
  background: var(--info-50);
  border: 1px solid var(--info-100);
  font-size: 12px;
  color: var(--text-primary);
  margin-bottom: 24px;
}
.welcome__distro .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--success); box-shadow: 0 0 8px var(--success); }
.welcome__distro .muted { color: var(--text-muted); }

.welcome__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.stat {
  display: flex; gap: 12px; align-items: flex-start;
  padding: 14px;
  border-radius: var(--r-md);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
}
.stat__icon { color: var(--accent-hover); margin-top: 2px; flex-shrink: 0; }
.stat__body { flex: 1; min-width: 0; }
.stat__label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.06em; margin-bottom: 6px; }
.stat__val { font-size: 16px; font-weight: 600; }
.stat__val--big { font-size: 18px; }
.stat__val--mono { font-family: var(--font-mono); font-size: 13px; }
.stat__val.ok { color: var(--success); }
.stat__val.warn { color: var(--warning); }
.stat__bar {
  height: 5px;
  border-radius: var(--r-full);
  background: rgba(255, 255, 255, 0.08);
  overflow: hidden;
  margin-bottom: 6px;
}
.stat__fill { height: 100%; background: linear-gradient(90deg, #3b82f6, #60a5fa); transition: width 0.8s var(--ease); }
.stat__fill--teal { background: linear-gradient(90deg, #2dd4bf, #34d399); }
.stat__fill--amber { background: linear-gradient(90deg, #fbbf24, #f59e0b); }

.welcome__footer {
  position: relative; z-index: 2;
  display: flex; align-items: center; justify-content: space-between;
  margin-top: 24px;
}
.link {
  background: none; border: none; color: var(--text-muted);
  font-size: 12px; text-decoration: underline; padding: 4px 0;
}
.link:hover { color: var(--text-secondary); }
.welcome__progress {
  width: 120px; height: 3px;
  border-radius: var(--r-full);
  background: rgba(255, 255, 255, 0.08);
  overflow: hidden;
}
.welcome__progress-fill { height: 100%; background: var(--accent); transition: width 0.1s linear; }

@media (max-width: 520px) {
  .welcome__grid { grid-template-columns: 1fr; }
}
</style>
