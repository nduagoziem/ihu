<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import * as App from '../wailsjs/go/main/App'
import { useToast } from './composables/useToast'
import ToastStack from './components/ToastStack.vue'
import SystemStats from './components/SystemStats.vue'
import TopBar from './components/TopBar.vue'
import Desktop from './components/Desktop.vue'
import BottomDock from './components/BottomDock.vue'
import TerminalDrawer from './components/TerminalDrawer.vue'
import CommandHelp from './components/CommandHelp.vue'
import TextEditor from './components/TextEditor.vue'
import FileViewer from './components/FileViewer.vue'
import BackgroundPicker from './components/BackgroundPicker.vue'
import ReclaimSpace from './components/ReclaimSpace.vue'
import bgImage1 from './assets/images/app-bg-img-1.jpg'
import bgImage2 from './assets/images/app-bg-img-2.jpg'

const config = reactive({
  defaultLinuxDistro: '',
  pinnedFolders: [],
  backgroundImage: '',
  backgroundMode: 'gradient',
})

const ui = reactive({
  showStats: false,
  showTerminal: false,
  showHelp: false,
  showEditor: false,
  showViewer: false,
  showBackgroundPicker: false,
  showReclaim: false,
})

// reclaimBusy is true while a space-reclaim operation is mid-flight. It hard
// locks the rest of the app: no shortcuts, no dismissing, no other panels.
const reclaimBusy = ref(false)

const { notify } = useToast()
const editorFile = ref(null)
const viewerFile = ref(null)
const cwd = ref('')
const currentUser = ref('')
const currentDistro = ref('')
const superUser = ref(false)
const systemStats = ref(null)
const bgRefresh = ref(0)
const desktopRefresh = ref(0)
const modalActive = computed(() => (
  ui.showStats
  || ui.showTerminal
  || ui.showHelp
  || ui.showEditor
  || ui.showViewer
  || ui.showBackgroundPicker
  || ui.showReclaim
))

const desktopStyle = computed(() => {
  void bgRefresh.value
  const background = config.backgroundImage
  const mode = config.backgroundMode || 'gradient'
  const resolvedBackground = resolveBackground(background)
  if (resolvedBackground) {
    if (mode === 'gradient') {
      return { backgroundImage: resolvedBackground }
    }
    const backgroundImage = resolvedBackground === bgImage2
      ? `linear-gradient(135deg, rgba(5, 8, 12, 0.32), rgba(20, 32, 46, 0.2), rgba(5, 6, 8, 0.28)), url("${resolvedBackground}")`
      : `url("${resolvedBackground}")`

    return {
      backgroundImage,
      backgroundSize: mode === 'cover' ? 'cover' : 'contain',
      backgroundPosition: 'center',
      backgroundRepeat: 'no-repeat',
    }
  }
  return {}
})

onMounted(async () => {
  try {
    const cfg = await App.GetConfig()
    if (cfg) Object.assign(config, cfg)
  } catch (e) {
    notify('Could not load your saved settings: ' + errMsg(e))
  }
  currentDistro.value = config.defaultLinuxDistro
  await navigateDefaultHome()
  await refreshSystemStats()
  window.addEventListener('keydown', onKey)
  window.addEventListener('keydown', onNavKey)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('keydown', onNavKey)
})

function onKey(e) {
  // While a reclaim is running the whole app is locked — swallow shortcuts.
  if (reclaimBusy.value) return
  const mod = e.ctrlKey || e.metaKey
  if (mod && e.key.toLowerCase() === 't') {
    e.preventDefault()
    ui.showTerminal = !ui.showTerminal
  } else if (mod && e.key.toLowerCase() === 'h') {
    e.preventDefault()
    ui.showHelp = !ui.showHelp
  } else if (e.shiftKey && e.key.toLowerCase() === 'i') {
    e.preventDefault()
    ui.showBackgroundPicker = true
  } else if (e.key === 'Escape') {
    ui.showTerminal = false
    ui.showHelp = false
    ui.showEditor = false
    ui.showViewer = false
    ui.showStats = false
    ui.showBackgroundPicker = false
    ui.showReclaim = false
  }
}

function onNavKey(e) {
  if (reclaimBusy.value) return
  if (ui.showTerminal || ui.showHelp || ui.showEditor || ui.showViewer || ui.showBackgroundPicker || ui.showReclaim) return
  if (!e.altKey) return
  if (e.key === 'ArrowLeft') {
    e.preventDefault()
    navigateBack()
  } else if (e.key === 'ArrowRight') {
    e.preventDefault()
    navigateForward()
  }
}

const history = ref([])
const historyIndex = ref(-1)
function navigateTo(path) {
  if (path === cwd.value) return
  if (historyIndex.value < history.value.length - 1) {
    history.value = history.value.slice(0, historyIndex.value + 1)
  }
  history.value.push(cwd.value)
  historyIndex.value = history.value.length - 1
  cwd.value = path
}
function navigateBack() {
  if (historyIndex.value < 0) return
  cwd.value = history.value[historyIndex.value]
  historyIndex.value--
}
function navigateForward() {
  if (historyIndex.value >= history.value.length - 2) return
  historyIndex.value++
  cwd.value = history.value[historyIndex.value + 1]
}
const canGoBack = computed(() => historyIndex.value >= 0)
const canGoForward = computed(() => historyIndex.value < history.value.length - 2)

function openInEditor(file) {
  editorFile.value = file
  ui.showEditor = true
}
function openInViewer(file) {
  viewerFile.value = file
  ui.showViewer = true
}

async function onTogglePin(path) {
  try {
    const cfg = await App.TogglePinnedFolder(path)
    if (cfg) Object.assign(config, cfg)
  } catch (e) {
    notify('Could not update pinned folders: ' + errMsg(e))
  }
}
async function onSetBackground(image, mode) {
  try {
    const cfg = await App.SetBackground(image, mode)
    if (cfg) Object.assign(config, cfg)
    bgRefresh.value++
    ui.showBackgroundPicker = false
  } catch (e) {
    notify('Could not save background: ' + errMsg(e))
  }
}
async function onConfigUpdate(partial) {
  Object.assign(config, partial)
  if (partial.defaultLinuxDistro) currentDistro.value = partial.defaultLinuxDistro
}
async function navigateHome() {
  try {
    const home = await App.HomePath(currentUser.value)
    navigateTo(home || "/")
  } catch (e) {
    notify(errMsg(e))
  }
}
async function navigateDefaultHome() {
  try {
    const home = await App.DefaultHomePath()
    currentUser.value = home?.user
    superUser.value = currentUser.value === 'root'
    navigateTo(home?.home)
  } catch (e) {
    notify(errMsg(e))
  }
}
function refreshDesktop() {
  desktopRefresh.value++
}
async function openSystemStats() {
  await refreshSystemStats()
  ui.showStats = true
}

// openReclaim closes every other panel first so the reclaim modal owns the
// screen, then opens it. The actual app-wide lock kicks in once the operation
// starts (see onReclaimBusy).
function openReclaim() {
  ui.showStats = false
  ui.showTerminal = false
  ui.showHelp = false
  ui.showEditor = false
  ui.showViewer = false
  ui.showBackgroundPicker = false
  ui.showReclaim = true
}
function onReclaimBusy(busy) {
  reclaimBusy.value = busy
}
async function onReclaimClosed() {
  ui.showReclaim = false
  // The WSL session was torn down and restarted; refresh the environment view.
  await refreshSystemStats()
  refreshDesktop()
}
async function refreshSystemStats() {
  try {
    systemStats.value = await App.GetStats()
  } catch (e) {
    notify('Could not read system stats: ' + errMsg(e))
  }
}

function errMsg(e) {
  return String(e?.message || e || 'unknown error')
}
function resolveBackground(background) {
  if (background === '../assets/images/app-bg-img-1.jpg') return bgImage1
  if (background === '../assets/images/app-bg-img-2.jpg') return bgImage2
  return background
}
</script>

<template>
  <div class="app-shell" :style="desktopStyle">
    <div class="app-shell__overlay"></div>
    <Transition name="fade">
      <div v-if="modalActive" class="modal-frost"></div>
    </Transition>

    <TopBar
      :current-user="currentUser"
      :current-distro="currentDistro"
      :config="config"
      :cwd="cwd"
      :can-go-back="canGoBack"
      :can-go-forward="canGoForward"
      :show-terminal="ui.showTerminal"
      :show-help="ui.showHelp"
      :super-user="superUser"
      @navigate="navigateTo"
      @back="navigateBack"
      @forward="navigateForward"
      @home="navigateHome"
      @refresh="refreshDesktop"
      @update:user="currentUser = $event"
      @update:distro="currentDistro = $event"
      @update:super-user="superUser = $event"
      @config-update="onConfigUpdate"
      @show-stats="openSystemStats"
      @show-background="ui.showBackgroundPicker = true"
      @show-reclaim="openReclaim"
      @toggle-terminal="ui.showTerminal = !ui.showTerminal"
      @toggle-help="ui.showHelp = !ui.showHelp"
    />

    <Desktop
      :cwd="cwd"
      :config="config"
      :current-user="currentUser"
      :current-distro="currentDistro"
      :super-user="superUser"
      :refresh-key="desktopRefresh"
      @navigate="navigateTo"
      @open-editor="openInEditor"
      @open-viewer="openInViewer"
      @toggle-pin="onTogglePin"
    />

    <BottomDock
      :config="config"
      :cwd="cwd"
      :current-user="currentUser"
      @navigate="navigateTo"
      @toggle-pin="onTogglePin"
    />

    <Transition name="fade">
      <SystemStats
        v-if="ui.showStats"
        :stats="systemStats"
        @close="ui.showStats = false"
      />
    </Transition>

    <Transition name="terminal-slide">
      <TerminalDrawer v-if="ui.showTerminal" :cwd="cwd" :user="currentUser" :distro="currentDistro" :super-user="superUser" @close="ui.showTerminal = false" />
    </Transition>

    <Transition name="fade">
      <CommandHelp v-if="ui.showHelp" @close="ui.showHelp = false" />
    </Transition>

    <Transition name="scale">
      <TextEditor v-if="ui.showEditor" :file="editorFile" :current-user="currentUser" :current-distro="currentDistro" :super-user="superUser" @close="ui.showEditor = false" />
    </Transition>

    <Transition name="scale">
      <FileViewer v-if="ui.showViewer" :file="viewerFile" :current-user="currentUser" :current-distro="currentDistro" :super-user="superUser" @close="ui.showViewer = false" />
    </Transition>

    <Transition name="scale">
      <BackgroundPicker v-if="ui.showBackgroundPicker" :config="config" @apply="onSetBackground" @close="ui.showBackgroundPicker = false" />
    </Transition>

    <Transition name="scale">
      <ReclaimSpace
        v-if="ui.showReclaim"
        :distro="currentDistro"
        @busy="onReclaimBusy"
        @close="onReclaimClosed"
      />
    </Transition>

    <!-- Hard app-wide lock while a reclaim operation is in flight. Sits above
         everything except the reclaim modal and swallows all interaction. -->
    <div v-if="reclaimBusy" class="busy-lock" @click.stop @contextmenu.prevent></div>

    <ToastStack />
  </div>
</template>

<style scoped>
.app-shell {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: radial-gradient(120% 120% at 0% 0%, #14202e 0%, #0a0c10 55%, #050608 100%);
  background-size: cover;
  background-position: center;
}
.app-shell__overlay {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(90% 70% at 20% 0%, rgba(59, 130, 246, 0.18), transparent 55%),
    linear-gradient(180deg, rgba(8, 10, 14, 0.58) 0%, rgba(8, 10, 14, 0.32) 30%, rgba(8, 10, 14, 0.66) 100%);
  pointer-events: none;
  z-index: 1;
}
.modal-frost {
  position: fixed;
  inset: 0;
  z-index: 45;
  background:
    linear-gradient(135deg, rgba(10, 14, 22, 0.62), rgba(24, 34, 48, 0.34)),
    rgba(5, 6, 8, 0.28);
  backdrop-filter: blur(18px) saturate(170%);
  -webkit-backdrop-filter: blur(18px) saturate(170%);
  pointer-events: none;
}
.busy-lock {
  position: fixed;
  inset: 0;
  z-index: 105;
  cursor: wait;
  background: transparent;
}

.terminal-slide-enter-active,
.terminal-slide-leave-active {
  transition: transform var(--t-slow) var(--ease-out), opacity var(--t-med) var(--ease);
}
.terminal-slide-enter-from,
.terminal-slide-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
