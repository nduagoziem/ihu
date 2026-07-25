<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import {
  Terminal, BookOpen, Image as ImageIcon, X, RotateCw, Settings,
} from '@lucide/vue'
import * as App from '../wailsjs/go/main/App'
import WelcomeScreen from './components/WelcomeScreen.vue'
import TopBar from './components/TopBar.vue'
import Desktop from './components/Desktop.vue'
import BottomDock from './components/BottomDock.vue'
import TerminalDrawer from './components/TerminalDrawer.vue'
import CommandHelp from './components/CommandHelp.vue'
import TextEditor from './components/TextEditor.vue'
import FileViewer from './components/FileViewer.vue'
import BackgroundPicker from './components/BackgroundPicker.vue'

const config = reactive({
  welcomeDisabled: false,
  defaultLinuxPath: '/root',
  defaultLinuxUser: 'root',
  defaultLinuxDistro: '',
  pinnedFolders: [],
  backgroundImage: '',
  backgroundMode: 'gradient',
})

const ui = reactive({
  showWelcome: false,
  showTerminal: false,
  showHelp: false,
  showEditor: false,
  showViewer: false,
  showBackgroundPicker: false,
})

const editorFile = ref(null)
const viewerFile = ref(null)
const cwd = ref('/root')
const currentUser = ref('root')
const currentDistro = ref('')
const bootData = ref(null)
const bgRefresh = ref(0)

const desktopStyle = computed(() => {
  void bgRefresh.value
  const img = config.backgroundImage
  const mode = config.backgroundMode || 'gradient'
  if (img) {
    const cover = mode === 'cover' || mode === 'gradient'
    return {
      backgroundImage: `url("${img}")`,
      backgroundSize: cover ? 'cover' : 'contain',
      backgroundPosition: 'center',
      backgroundRepeat: mode === 'contain' ? 'no-repeat' : 'no-repeat',
    }
  }
  return {}
})

onMounted(async () => {
  const cfg = await App.GetConfig().catch(() => null)
  if (cfg) Object.assign(config, cfg)
  cwd.value = config.defaultLinuxPath || '/root'
  currentUser.value = config.defaultLinuxUser || 'root'
  bootData.value = await App.GetBootData().catch(() => null)
  if (bootData.value && !config.welcomeDisabled) ui.showWelcome = true
  window.addEventListener('keydown', onKey)
  window.addEventListener('keydown', onNavKey)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('keydown', onNavKey)
})

function onKey(e) {
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
    ui.showBackgroundPicker = false
  }
}

function onNavKey(e) {
  if (ui.showTerminal || ui.showHelp || ui.showEditor || ui.showViewer || ui.showBackgroundPicker) return
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
  const cfg = await App.TogglePinnedFolder(path).catch(() => null)
  if (cfg) Object.assign(config, cfg)
}
async function onSetBackground(image, mode) {
  const cfg = await App.SetBackground(image, mode).catch(() => null)
  if (cfg) Object.assign(config, cfg)
  bgRefresh.value++
  ui.showBackgroundPicker = false
}
async function onConfigUpdate(partial) {
  Object.assign(config, partial)
}
</script>

<template>
  <div class="app-shell" :style="desktopStyle">
    <div class="app-shell__overlay"></div>

    <TopBar
      :current-user="currentUser"
      :current-distro="currentDistro"
      :config="config"
      :cwd="cwd"
      :can-go-back="canGoBack"
      :can-go-forward="canGoForward"
      @navigate="navigateTo"
      @back="navigateBack"
      @forward="navigateForward"
      @update:user="currentUser = $event"
      @update:distro="currentDistro = $event"
      @config-update="onConfigUpdate"
      @show-welcome="ui.showWelcome = true"
      @show-background="ui.showBackgroundPicker = true"
    />

    <Desktop
      :cwd="cwd"
      :config="config"
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

    <div class="floating-actions">
      <button class="fab" :class="{ active: ui.showTerminal }" title="Terminal (Ctrl/Cmd+T)" @click="ui.showTerminal = !ui.showTerminal">
        <Terminal :size="18" />
      </button>
      <button class="fab" :class="{ active: ui.showHelp }" title="Command Help (Ctrl/Cmd+H)" @click="ui.showHelp = !ui.showHelp">
        <BookOpen :size="18" />
      </button>
      <button class="fab" title="Background (Shift+I)" @click="ui.showBackgroundPicker = true">
        <ImageIcon :size="18" />
      </button>
    </div>

    <Transition name="fade">
      <WelcomeScreen
        v-if="ui.showWelcome"
        :boot-data="bootData"
        @close="ui.showWelcome = false"
        @disable="async (v) => { await App.SetWelcomeDisabled(v); config.welcomeDisabled = v; ui.showWelcome = false }"
      />
    </Transition>

    <Transition name="terminal-slide">
      <TerminalDrawer v-if="ui.showTerminal" :cwd="cwd" :user="currentUser" @close="ui.showTerminal = false" />
    </Transition>

    <Transition name="fade">
      <CommandHelp v-if="ui.showHelp" @close="ui.showHelp = false" />
    </Transition>

    <Transition name="scale">
      <TextEditor v-if="ui.showEditor" :file="editorFile" @close="ui.showEditor = false" />
    </Transition>

    <Transition name="scale">
      <FileViewer v-if="ui.showViewer" :file="viewerFile" @close="ui.showViewer = false" />
    </Transition>

    <Transition name="scale">
      <BackgroundPicker v-if="ui.showBackgroundPicker" :config="config" @apply="onSetBackground" @close="ui.showBackgroundPicker = false" />
    </Transition>
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
  background: linear-gradient(180deg, rgba(8, 10, 14, 0.45) 0%, rgba(8, 10, 14, 0.2) 30%, rgba(8, 10, 14, 0.5) 100%);
  pointer-events: none;
  z-index: 1;
}

.floating-actions {
  position: fixed;
  right: 20px;
  bottom: 96px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  z-index: 40;
}
.fab {
  width: 44px;
  height: 44px;
  border-radius: var(--r-full);
  border: 1px solid var(--frost-border);
  background: var(--frost-bg);
  backdrop-filter: blur(20px) saturate(160%);
  -webkit-backdrop-filter: blur(20px) saturate(160%);
  color: var(--text-secondary);
  display: grid;
  place-items: center;
  transition: all var(--t-med) var(--ease);
  box-shadow: var(--shadow-md);
}
.fab:hover {
  color: var(--text-primary);
  background: var(--info-100);
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}
.fab.active {
  color: #fff;
  background: var(--accent);
  border-color: transparent;
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
