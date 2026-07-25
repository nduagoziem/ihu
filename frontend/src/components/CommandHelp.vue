<script setup>
import { ref, computed } from 'vue'
import { X, Search, ChevronDown, ChevronRight } from '@lucide/vue'
import { commandGroups, searchCommands } from '../data/commands.js'

defineEmits(['close'])

const query = ref('')
const expanded = ref(new Set(['native']))
const totalCount = commandGroups.reduce((n, g) => n + g.commands.length, 0)

const results = computed(() => searchCommands(query.value))

function toggle(id) {
  const next = new Set(expanded.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expanded.value = next
}
</script>

<template>
  <div class="help-overlay" @click.self="emit('close')">
    <div class="help glass-strong">
      <div class="help__header">
        <div class="help__title">Command Help</div>
        <div class="help__search">
          <Search :size="15" class="help__search-icon" />
          <input v-model="query" placeholder="Search by command or type (e.g. sudo, PHP)…" spellcheck="false" />
        </div>
        <button class="help__close" @click="emit('close')"><X :size="18" /></button>
      </div>

      <div class="help__body">
        <div v-if="!results.length" class="help__empty">
          No commands match "{{ query }}".
        </div>
        <div v-for="group in results" :key="group.id" class="group">
          <button class="group__head" @click="toggle(group.id)">
            <component :is="expanded.has(group.id) ? ChevronDown : ChevronRight" :size="15" />
            <span class="group__dot" :style="{ background: group.color }"></span>
            <span class="group__name">{{ group.name }}</span>
            <span class="group__count">{{ group.commands.length }}</span>
          </button>
          <Transition name="fade">
            <div v-show="expanded.has(group.id) || query" class="group__list">
              <div v-for="(c, i) in group.commands" :key="i" class="cmd">
                <code class="cmd__name">{{ c.cmd }}</code>
                <span class="cmd__meaning">{{ c.meaning }}</span>
              </div>
            </div>
          </Transition>
        </div>
      </div>

      <div class="help__footer">
        <span>{{ totalCount }} commands across {{ commandGroups.length }} toolchains</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.help-overlay {
  position: fixed; inset: 0; z-index: 80;
  display: grid; place-items: center;
  background: rgba(5, 6, 8, 0.5);
  animation: fadeIn var(--t-med) var(--ease);
}
.help {
  width: min(720px, 92vw);
  max-height: 84vh;
  border-radius: var(--r-xl);
  display: flex; flex-direction: column;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-slow) var(--ease-out);
}
.help__header {
  display: flex; align-items: center; gap: 14px;
  padding: 18px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
}
.help__title { font-size: 16px; font-weight: 700; white-space: nowrap; }
.help__search {
  flex: 1; position: relative;
  display: flex; align-items: center;
}
.help__search-icon { position: absolute; left: 12px; color: var(--text-muted); pointer-events: none; }
.help__search input {
  width: 100%;
  height: 36px;
  padding: 0 12px 0 34px;
  border-radius: var(--r-md);
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(0, 0, 0, 0.25);
  color: var(--text-primary);
  font-size: 13px; outline: none;
  transition: border-color var(--t-fast) var(--ease);
}
.help__search input:focus { border-color: var(--accent); }
.help__close {
  width: 32px; height: 32px; border-radius: var(--r-sm);
  border: none; background: transparent; color: var(--text-muted);
  display: grid; place-items: center;
}
.help__close:hover { background: rgba(255,255,255,0.08); color: var(--text-primary); }

.help__body { flex: 1; overflow-y: auto; padding: 12px 20px; }
.help__empty { text-align: center; color: var(--text-muted); padding: 40px; }

.group { margin-bottom: 6px; }
.group__head {
  display: flex; align-items: center; gap: 8px;
  width: 100%;
  padding: 10px 8px;
  border: none; background: transparent;
  color: var(--text-secondary);
  border-radius: var(--r-sm);
  text-align: left;
  transition: background var(--t-fast) var(--ease);
}
.group__head:hover { background: rgba(255, 255, 255, 0.04); color: var(--text-primary); }
.group__dot { width: 9px; height: 9px; border-radius: 50%; }
.group__name { font-weight: 600; font-size: 14px; flex: 1; }
.group__count {
  font-size: 11px; color: var(--text-muted);
  padding: 2px 8px; border-radius: var(--r-full);
  background: rgba(255, 255, 255, 0.06);
}
.group__list { padding: 4px 0 8px 28px; display: flex; flex-direction: column; gap: 2px; }
.cmd {
  display: flex; align-items: baseline; gap: 14px;
  padding: 7px 10px;
  border-radius: var(--r-sm);
  transition: background var(--t-fast) var(--ease);
}
.cmd:hover { background: rgba(255, 255, 255, 0.04); }
.cmd__name {
  font-family: var(--font-mono); font-size: 12.5px;
  color: var(--teal);
  flex-shrink: 0; min-width: 200px;
}
.cmd__meaning { color: var(--text-secondary); font-size: 12.5px; }

.help__footer {
  padding: 12px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.07);
  font-size: 11px; color: var(--text-muted);
}
</style>
