<script setup>
import { computed } from "vue";
import {
  Cpu,
  MemoryStick,
  HardDrive,
  X,
  Sparkles,
  Clock,
  CircuitBoard,
} from "@lucide/vue";

const props = defineProps({
  stats: { type: Object, default: null },
});
const emit = defineEmits(["close"]);

const systemStats = computed(() => props.stats || {});
const timestampText = computed(() => systemStats.value.timestamp || "-");

function dismiss() {
  emit("close");
}
</script>

<template>
  <div class="stats-modal" @click="dismiss">
    <div class="stats-card glass-strong" @click.stop>
      <div class="stats-card__decor"></div>

      <button class="stats-card__close" title="Close" @click="dismiss">
        <X :size="18" />
      </button>

      <div class="stats-card__content">
        <div class="stats-card__brand">
          <div class="stats-card__logo">
            <Sparkles :size="28" />
          </div>
          <div>
            <h1>System Stats</h1>
            <p class="stats-card__sub">Current WSL environment snapshot</p>
          </div>
        </div>

        <div class="stats-card__distro" v-if="systemStats.distro">
          <span class="dot"></span>
          {{ systemStats.distro }}
          <span class="muted">- {{ systemStats.kernel }}</span>
        </div>

        <div class="stats-card__grid">
          <div class="stat">
            <Cpu :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">CPU</div>
              <div class="stat__bar">
                <div
                  class="stat__fill"
                  :style="{ width: (systemStats.cpuUsage || 0) + '%' }"
                ></div>
              </div>
              <div class="stat__val">{{ systemStats.cpuUsage || 0 }}%</div>
            </div>
          </div>

          <div class="stat">
            <MemoryStick :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">Memory</div>
              <div class="stat__bar">
                <div
                  class="stat__fill stat__fill--teal"
                  :style="{ width: (systemStats.memoryUsage || 0) + '%' }"
                ></div>
              </div>
              <div class="stat__val">{{ systemStats.memoryUsage || 0 }}%</div>
            </div>
          </div>

          <div class="stat">
            <HardDrive :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">Disk</div>
              <div class="stat__bar">
                <div
                  class="stat__fill stat__fill--amber"
                  :style="{ width: (systemStats.diskUsage || 0) + '%' }"
                ></div>
              </div>
              <div class="stat__val">{{ systemStats.diskUsage || 0 }}%</div>
            </div>
          </div>

          <div class="stat">
            <CircuitBoard :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">Architecture</div>
              <div class="stat__val stat__val--mono">
                {{ systemStats.arch || "-" }}
              </div>
            </div>
          </div>

          <div class="stat">
            <Clock :size="18" class="stat__icon" />
            <div class="stat__body">
              <div class="stat__label">Timestamp</div>
              <div class="stat__val stat__val--time">{{ timestampText }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stats-modal {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: grid;
  place-items: center;
  animation: fadeIn var(--t-med) var(--ease);
  background: rgba(5, 6, 8, 0.35);
}
.stats-card {
  position: relative;
  width: min(560px, 92vw);
  border-radius: var(--r-xl);
  padding: 36px 32px 28px;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-slow) var(--ease-out);
}
.stats-card__decor {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background:
    radial-gradient(circle at 0% 0%, rgba(30, 111, 235, 0.42), transparent 34%),
    radial-gradient(
      circle at 100% 100%,
      rgba(45, 212, 191, 0.3),
      transparent 32%
    ),
    radial-gradient(
      circle at 65% 45%,
      rgba(251, 191, 36, 0.16),
      transparent 28%
    );
}
.stats-card__close {
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
.stats-card__close:hover {
  background: rgba(255, 255, 255, 0.16);
  color: var(--text-primary);
}
.stats-card__content {
  position: relative;
  z-index: 2;
}
.stats-card__brand {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
}
.stats-card__logo {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  background: linear-gradient(135deg, #3b82f6, #2dd4bf);
  display: grid;
  place-items: center;
  color: #fff;
  box-shadow: 0 8px 22px rgba(59, 130, 246, 0.45);
}
.stats-card__brand h1 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
}
.stats-card__sub {
  margin: 2px 0 0;
  color: var(--text-secondary);
  font-size: 13px;
}
.stats-card__distro {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: var(--r-full);
  background: var(--info-50);
  border: 1px solid var(--info-100);
  font-size: 12px;
  color: var(--text-primary);
  margin-bottom: 24px;
}
.stats-card__distro .dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--success);
  box-shadow: 0 0 8px var(--success);
}
.stats-card__distro .muted {
  color: var(--text-muted);
}
.stats-card__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.stat {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  padding: 14px;
  border-radius: var(--r-md);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
}
.stat__icon {
  color: var(--accent-hover);
  margin-top: 2px;
  flex-shrink: 0;
}
.stat__body {
  flex: 1;
  min-width: 0;
}
.stat__label {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 6px;
}
.stat__val {
  font-size: 16px;
  font-weight: 600;
}
.stat__val--big {
  font-size: 18px;
}
.stat__val--mono {
  font-family: var(--font-mono);
  font-size: 13px;
}
.stat__val--time {
  font-size: 13px;
  line-height: 1.35;
}
.stat__val.ok {
  color: var(--success);
}
.stat__val.warn {
  color: var(--warning);
}
.stat__bar {
  height: 5px;
  border-radius: var(--r-full);
  background: rgba(255, 255, 255, 0.08);
  overflow: hidden;
  margin-bottom: 6px;
}
.stat__fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
  transition: width 0.8s var(--ease);
}
.stat__fill--teal {
  background: linear-gradient(90deg, #2dd4bf, #34d399);
}
.stat__fill--amber {
  background: linear-gradient(90deg, #fbbf24, #f59e0b);
}

@media (max-width: 520px) {
  .stats-card__grid {
    grid-template-columns: 1fr;
  }
}
</style>
