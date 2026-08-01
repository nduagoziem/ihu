<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import { Folder, Pin, PinOff, Home, Star, Clock } from '@lucide/vue'
import * as App from '../../wailsjs/go/main/App'
import { useToast } from '../composables/useToast'

const props = defineProps({
  config: Object,
  cwd: String,
  currentUser: String,
})
const emit = defineEmits(['navigate', 'toggle-pin'])

const homePath = ref('/')
const rootPath = ref('/')
const openFolders = ref([])
const { notify } = useToast()

onMounted(async () => {
  try {
    homePath.value = await App.HomePath(props.currentUser)
  } catch (e) {
    notify(errStr(e))
  }
})
watch(() => props.currentUser, async (u) => {
  try {
    homePath.value = await App.HomePath(u)
  } catch (e) {
    notify(errStr(e))
  }
})

watch(() => props.cwd, (c, old) => {
  if (c && c !== old && !openFolders.value.includes(c)) {
    openFolders.value = [c, ...openFolders.value.filter((p) => p !== c)].slice(0, 6)
  }
}, { immediate: true })

const pinned = computed(() => props.config.pinnedFolders || [])
function isPinned(path) { return pinned.value.includes(path) }

function shortName(path) {
  if (!path || path === '/') return 'root'
  const parts = path.split('/').filter(Boolean)
  return parts[parts.length - 1] || path
}

function nav(path) { emit('navigate', path) }
function togglePin(path) { emit('toggle-pin', path) }

function errStr(e) {
  return String(e?.message || e || 'unknown error')
}

const sections = computed(() => [
  { id: 'root', label: 'Root', icon: Home, items: [{ path: rootPath.value, name: 'root' }, { path: homePath.value, name: 'home' }] },
  { id: 'pinned', label: 'Pinned', icon: Pin, items: pinned.value.map((p) => ({ path: p, name: shortName(p) })) },
  { id: 'open', label: 'Open', icon: Clock, items: openFolders.value.map((p) => ({ path: p, name: shortName(p) })) },
])
</script>

<template>
  <div class="dock glass">
    <div v-for="section in sections" :key="section.id" class="dock__section">
      <div class="dock__label">
        <component :is="section.icon" :size="11" />
        <span>{{ section.label }}</span>
      </div>
      <div class="dock__items">
        <template v-if="section.items.length">
          <button
            v-for="item in section.items"
            :key="item.path"
            class="dock-item"
            :class="{ active: cwd === item.path }"
            :title="item.path"
            @click="nav(item.path)"
            @contextmenu.prevent="togglePin(item.path)"
          >
            <Folder :size="18" class="dock-item__icon" />
            <span class="dock-item__name">{{ item.name }}</span>
            <button
              v-if="section.id === 'open' || section.id === 'root'"
              class="dock-item__pin"
              :class="{ on: isPinned(item.path) }"
              :title="isPinned(item.path) ? 'Unpin' : 'Pin'"
              @click.stop="togglePin(item.path)"
            >
              <component :is="isPinned(item.path) ? Pin : PinOff" :size="11" />
            </button>
          </button>
        </template>
        <span v-else class="dock__empty">Nothing pinned yet</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dock {
  position: fixed;
  left: 12px; right: 12px; bottom: 12px;
  height: 88px;
  border-radius: var(--r-lg);
  display: flex;
  align-items: stretch;
  gap: 14px;
  padding: 10px 14px;
  z-index: 25;
  box-shadow: var(--shadow-md);
  overflow-x: auto;
}
.dock__section { display: flex; flex-direction: column; gap: 6px; min-width: 0; flex: 1; }
.dock__section + .dock__section { border-left: 1px solid rgba(255, 255, 255, 0.07); padding-left: 14px; }
.dock__label {
  display: flex; align-items: center; gap: 5px;
  font-size: 10px; text-transform: uppercase; letter-spacing: 0.08em;
  color: var(--text-muted);
  padding-left: 2px;
}
.dock__items { display: flex; gap: 8px; align-items: center; overflow-x: auto; }
.dock__empty { font-size: 11px; color: var(--text-muted); padding: 8px; }

.dock-item {
  position: relative;
  display: flex; flex-direction: column; align-items: center; gap: 4px;
  width: 64px;
  padding: 6px 4px 8px;
  border-radius: var(--r-sm);
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-secondary);
  transition: all var(--t-fast) var(--ease);
  flex-shrink: 0;
}
.dock-item:hover { background: rgba(255, 255, 255, 0.06); color: var(--text-primary); }
.dock-item.active { background: var(--info-100); border-color: var(--info-300); color: var(--accent-hover); }
.dock-item__icon { flex-shrink: 0; }
.dock-item__name {
  font-size: 10.5px;
  max-width: 56px;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.dock-item__pin {
  position: absolute;
  top: 2px; right: 2px;
  width: 16px; height: 16px;
  border-radius: var(--r-full);
  border: none;
  background: rgba(0, 0, 0, 0.4);
  color: var(--text-muted);
  display: grid; place-items: center;
  opacity: 0;
  transition: all var(--t-fast) var(--ease);
}
.dock-item:hover .dock-item__pin { opacity: 1; }
.dock-item__pin:hover { background: var(--accent); color: #fff; }
.dock-item__pin.on { opacity: 1; color: var(--amber); }
</style>
