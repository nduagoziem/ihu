<script setup>
import { ref, onMounted, watch } from 'vue'
import { ChevronLeft, ChevronRight, ChevronDown, Check, Home, Eye, RefreshCw, Image as ImageIcon, Terminal, BookOpen, Shield, HardDriveDownload } from '@lucide/vue'
import * as App from '../../wailsjs/go/main/App'
import { useToast } from '../composables/useToast'

const props = defineProps({
  currentUser: String,
  currentDistro: String,
  config: Object,
  cwd: String,
  canGoBack: Boolean,
  canGoForward: Boolean,
  showTerminal: Boolean,
  showHelp: Boolean,
  superUser: Boolean,
})
const emit = defineEmits(['navigate', 'back', 'forward', 'home', 'refresh', 'update:user', 'update:distro', 'update:super-user', 'config-update', 'show-stats', 'show-background', 'show-reclaim', 'toggle-terminal', 'toggle-help'])

const distros = ref([])
const users = ref([])
const openMenu = ref(null)
const makeDefaultDistro = ref(false)
const pathInput = ref(props.cwd)
const { notify } = useToast()

onMounted(async () => {
  try {
    distros.value = await App.ListDistros()
  } catch (e) {
    notify('Could not list WSL distros: ' + errStr(e))
  }
  try {
    users.value = await App.ListUsers()
  } catch (e) {
    notify('Could not list users: ' + errStr(e))
  }
  if (!distros.value.length) distros.value = ['default']
  if (props.config.defaultLinuxDistro && !distros.value.includes(props.config.defaultLinuxDistro)) {
    distros.value = [props.config.defaultLinuxDistro, ...distros.value]
  }
  if (!props.currentDistro) {
    emit('update:distro', props.config.defaultLinuxDistro || distros.value[0])
  }
})

watch(() => props.cwd, (cwd) => {
  pathInput.value = cwd || '/'
})

function toggleMenu(name) {
  openMenu.value = openMenu.value === name ? null : name
}
function closeMenus() { openMenu.value = null }

async function chooseUser(u) {
  emit('update:user', u)
  emit('update:super-user', u === 'root')
  try {
    const home = await App.HomePath(u)
    emit('navigate', home)
    emit('refresh')
  } catch (e) {
    notify(errStr(e))
  }
  closeMenus()
}
async function chooseDistro(d) {
  emit('update:distro', d)
  if (makeDefaultDistro.value) {
    await setDefaultDistro(d)
    makeDefaultDistro.value = false
  }
  closeMenus()
}
async function setDefaultDistro(distro = props.currentDistro) {
  if (!distro) return
  try {
    const cfg = await App.SetDefaultLinuxDistro(distro)
    if (cfg) emit('config-update', { defaultLinuxDistro: cfg.defaultLinuxDistro })
  } catch (e) {
    notify('Could not set default distro: ' + errStr(e))
  }
}
async function onDefaultDistroToggle() {
  if (makeDefaultDistro.value) await setDefaultDistro()
}
function onPathKeydown(e) {
  if (e.key === 'Enter') {
    emit('navigate', pathInput.value.trim() || '/')
    closeMenus()
  }
}
function refresh() {
  pathInput.value = props.cwd
  emit('refresh')
}

function errStr(e) {
  return String(e?.message || e || 'unknown error')
}
</script>

<template>
  <div class="topbar glass" @click="closeMenus">
    <div class="topbar__nav">
      <button class="icon-btn" :disabled="!canGoBack" title="Back (Alt+Left)" @click.stop="emit('back')">
        <ChevronLeft :size="18" />
      </button>
      <button class="icon-btn" :disabled="!canGoForward" title="Forward (Alt+Right)" @click.stop="emit('forward')">
        <ChevronRight :size="18" />
      </button>
      <button class="icon-btn" title="Home" @click.stop="emit('home')">
        <Home :size="16" />
      </button>
    </div>

    <div class="topbar__path">
      <input
        v-model="pathInput"
        class="path-input"
        spellcheck="false"
        @keydown="onPathKeydown"
        @click.stop
      />
      <button class="icon-btn path-refresh" title="Refresh" @click.stop="refresh">
        <RefreshCw :size="14" />
      </button>
    </div>

    <div class="topbar__right">
      <div class="menu-host">
        <button class="chip" :class="{ open: openMenu === 'distro' }" @click.stop="toggleMenu('distro')">
          <span class="chip__label">Distro</span>
          <span class="chip__value">{{ currentDistro || config.defaultLinuxDistro || 'default' }}</span>
          <ChevronDown :size="14" />
        </button>
        <Transition name="fade">
          <div v-if="openMenu === 'distro'" class="menu glass-strong" @click.stop>
            <div class="menu__list">
              <button
                v-for="d in distros"
                :key="d"
                class="menu__item"
                :class="{ active: d === currentDistro }"
                @click="chooseDistro(d)"
              >
                <Check v-if="d === currentDistro" :size="14" class="menu__check" />
                <span>{{ d }}</span>
              </button>
            </div>
            <label class="menu__default">
              <input type="checkbox" v-model="makeDefaultDistro" @change="onDefaultDistroToggle" />
              Set current as default
            </label>
          </div>
        </Transition>
      </div>

      <div class="menu-host">
        <button class="chip" :class="{ open: openMenu === 'user' }" @click.stop="toggleMenu('user')">
          <span class="chip__label">User</span>
          <span class="chip__value">{{ currentUser }}</span>
          <ChevronDown :size="14" />
        </button>
        <Transition name="fade">
          <div v-if="openMenu === 'user'" class="menu glass-strong" @click.stop>
            <div class="menu__list">
              <button
                v-for="u in users"
                :key="u"
                class="menu__item"
                :class="{ active: u === currentUser }"
                @click="chooseUser(u)"
              >
                <Check v-if="u === currentUser" :size="14" class="menu__check" />
                <span>{{ u }}</span>
              </button>
            </div>
          </div>
        </Transition>
      </div>

      <button class="icon-btn" :class="{ active: superUser }" title="Super User" @click.stop="emit('update:super-user', !superUser)">
        <Shield :size="16" />
      </button>
      <button class="icon-btn" title="System Stats" @click.stop="emit('show-stats')">
        <Eye :size="16" />
      </button>
      <button class="icon-btn" title="Reclaim Disk Space" @click.stop="emit('show-reclaim')">
        <HardDriveDownload :size="16" />
      </button>
      <button class="icon-btn" :class="{ active: showTerminal }" title="Terminal (Ctrl+T)" @click.stop="emit('toggle-terminal')">
        <Terminal :size="16" />
      </button>
      <button class="icon-btn" :class="{ active: showHelp }" title="Command Help (Ctrl+H)" @click.stop="emit('toggle-help')">
        <BookOpen :size="16" />
      </button>
      <button class="icon-btn" title="Background (Shift+I)" @click.stop="emit('show-background')">
        <ImageIcon :size="16" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.topbar {
  position: fixed;
  top: 6px;
  left: 12px;
  right: 12px;
  height: 46px;
  border-radius: var(--r-lg);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 10px;
  z-index: 70;
  box-shadow: var(--shadow-md);
}
.topbar__nav { display: flex; gap: 2px; }
.topbar__path { flex: 1; display: flex; align-items: center; gap: 6px; min-width: 0; }
.path-input {
  flex: 1;
  min-width: 0;
  height: 30px;
  background: rgba(0, 0, 0, 0.22);
  border: 1px solid transparent;
  border-radius: var(--r-sm);
  padding: 0 10px;
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 12.5px;
  outline: none;
  transition: all var(--t-fast) var(--ease);
}
.path-input:focus { border-color: var(--accent); background: rgba(0, 0, 0, 0.35); }
.path-refresh { opacity: 0.6; }
.path-refresh:hover { opacity: 1; }

.topbar__right { display: flex; align-items: center; gap: 8px; }

.icon-btn {
  width: 32px; height: 32px;
  border-radius: var(--r-sm);
  border: none;
  background: transparent;
  color: var(--text-secondary);
  display: grid; place-items: center;
  transition: all var(--t-fast) var(--ease);
}
.icon-btn:hover:not(:disabled) { background: rgba(255, 255, 255, 0.08); color: var(--text-primary); }
.icon-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.icon-btn.active { background: var(--info-100); color: var(--accent-hover); box-shadow: inset 0 0 0 1px var(--info-300); }

.chip {
  display: flex; align-items: center; gap: 6px;
  height: 30px;
  padding: 0 10px;
  border-radius: var(--r-sm);
  border: 1px solid transparent;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
  font-size: 12px;
  transition: all var(--t-fast) var(--ease);
}
.chip:hover { background: rgba(255, 255, 255, 0.09); color: var(--text-primary); }
.chip.open { background: var(--info-100); color: var(--text-primary); border-color: var(--info-300); }
.chip__label { color: var(--text-muted); }
.chip__value { font-weight: 600; max-width: 110px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.menu-host { position: relative; }
.menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 220px;
  border-radius: var(--r-md);
  padding: 6px;
  box-shadow: var(--shadow-lg);
  z-index: 50;
}
.menu__list { display: flex; flex-direction: column; gap: 2px; max-height: 240px; overflow-y: auto; }
.menu__item {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 10px;
  border-radius: var(--r-sm);
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  text-align: left;
  transition: all var(--t-fast) var(--ease);
}
.menu__item:hover { background: rgba(255, 255, 255, 0.07); color: var(--text-primary); }
.menu__item.active { color: var(--accent-hover); }
.menu__check { color: var(--accent); flex-shrink: 0; }
.menu__default {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 10px;
  margin-top: 4px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  font-size: 12px;
  color: var(--text-muted);
  cursor: pointer;
}
.menu__default input { accent-color: var(--accent); }
</style>
