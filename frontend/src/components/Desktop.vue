<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { Folder, FileText, FileCode, FileImage, File, Pin, PinOff, Pencil, Eye, ChevronRight } from '@lucide/vue'
import * as App from '../../wailsjs/go/main/App'

const props = defineProps({
  cwd: String,
  config: Object,
  currentUser: String,
  currentDistro: String,
  superUser: Boolean,
  refreshKey: Number,
})
const emit = defineEmits(['navigate', 'open-editor', 'open-viewer', 'toggle-pin'])

const entries = ref([])
const loading = ref(false)
const error = ref('')
const selected = ref(null)
const showHidden = ref(false)
const selectedAt = ref({ x: 0, y: 0, time: 0 })
let loadId = 0

onMounted(load)
watch(() => [props.cwd, props.currentUser, props.currentDistro, props.superUser, props.refreshKey], load)

async function load() {
  const id = ++loadId
  loading.value = true
  error.value = ''
  try {
    const nextEntries = await App.ListDirAs(props.cwd, props.currentDistro || '', props.currentUser || '', props.superUser || props.currentUser === 'root')
    if (id !== loadId) return
    entries.value = nextEntries
    selected.value = null
  } catch (e) {
    if (id !== loadId) return
    error.value = String(e?.message || e || 'Failed to load directory')
    entries.value = []
  } finally {
    if (id !== loadId) return
    loading.value = false
  }
}

const visibleEntries = computed(() => {
  if (showHidden.value) return entries.value
  return entries.value.filter((e) => !e.isHidden)
})

const breadcrumbs = computed(() => {
  const parts = (props.cwd || '/').split('/').filter(Boolean)
  const crumbs = [{ name: 'root', path: '/' }]
  let acc = ''
  for (const p of parts) {
    acc += '/' + p
    crumbs.push({ name: p, path: acc })
  }
  return crumbs
})

function iconFor(entry) {
  if (entry.isDir) return Folder
  const n = entry.name.toLowerCase()
  if (n.match(/\.(png|jpe?g|gif|webp|svg|bmp)$/)) return FileImage
  if (n.match(/\.(md|txt|sh|go|js|ts|py|php|json|ya?ml|conf)$/)) return FileCode
  if (n.match(/\.(pdf|docx?|rtf)$/)) return FileText
  return File
}

function isPinned(path) {
  return (props.config.pinnedFolders || []).includes(path)
}

function handleClick(entry, e) {
  const now = Date.now()
  if (selected.value === entry && now - selectedAt.value.time < 350) {
    openEntry(entry)
  } else {
    selected.value = entry
    selectedAt.value = { x: e.clientX, y: e.clientY, time: now }
  }
}

function openEntry(entry) {
  if (entry.isDir) {
    emit('navigate', entry.path)
    return
  }
  const n = entry.name.toLowerCase()
  if (n.match(/\.(png|jpe?g|gif|webp|svg|bmp|pdf|docx?)$/)) {
    emit('open-viewer', entry)
  } else {
    emit('open-editor', entry)
  }
}

function doubleClick(entry) {
  openEntry(entry)
}

function togglePin(entry) {
  emit('toggle-pin', entry.path)
}

function clearSelection() {
  selected.value = null
}

</script>

<template>
  <div class="desktop" @click.self="clearSelection">
    <div class="desktop__breadcrumbs glass">
      <button
        v-for="(c, i) in breadcrumbs"
        :key="c.path"
        class="crumb"
        :class="{ active: i === breadcrumbs.length - 1 }"
        @click="emit('navigate', c.path)"
      >
        <ChevronRight v-if="i > 0" :size="12" class="crumb__sep" />
        {{ c.name }}
      </button>
    </div>

    <div v-if="error" class="desktop__empty">
      <p class="desktop__empty-title">Couldn't open this folder</p>
      <p class="desktop__empty-sub">{{ error }}</p>
    </div>

    <div v-else-if="loading && !entries.length" class="desktop__empty">
      <p class="desktop__empty-title">Loading…</p>
    </div>

    <div v-else-if="!visibleEntries.length" class="desktop__empty">
      <p class="desktop__empty-title">This folder is empty</p>
    </div>

    <div v-else class="desktop__grid">
      <div
        v-for="entry in visibleEntries"
        :key="entry.path"
        class="file-item"
        :class="{ selected: selected === entry, pinned: isPinned(entry.path) }"
        @click="handleClick(entry, $event)"
        @dblclick="doubleClick(entry)"
      >
        <div class="file-item__icon">
          <component :is="iconFor(entry)" :size="34" />
          <Pin v-if="isPinned(entry.path)" :size="12" class="file-item__pin" />
        </div>
        <div class="file-item__name" :title="entry.name">{{ entry.name }}</div>
        <div class="file-item__actions" v-if="selected === entry">
          <button v-if="entry.isDir" class="mini" title="Pin/Unpin" @click.stop="togglePin(entry)">
            <component :is="isPinned(entry.path) ? PinOff : Pin" :size="13" />
          </button>
          <button v-if="!entry.isDir" class="mini" title="Open in editor" @click.stop="emit('open-editor', entry)">
            <Pencil :size="13" />
          </button>
          <button v-if="!entry.isDir" class="mini" title="Open in viewer" @click.stop="emit('open-viewer', entry)">
            <Eye :size="13" />
          </button>
        </div>
      </div>
    </div>

    <button class="hidden-toggle" :class="{ on: showHidden }" @click.stop="showHidden = !showHidden" title="Toggle hidden files">
      {{ showHidden ? 'Hide' : 'Show' }} hidden
    </button>
  </div>
</template>

<style scoped>
.desktop {
  position: absolute;
  top: 68px; left: 12px; right: 12px; bottom: 116px;
  z-index: 5;
}
.desktop__breadcrumbs {
  position: absolute;
  top: -6px; left: 0;
  height: 34px;
  display: flex; align-items: center; gap: 2px;
  padding: 0 10px;
  border-radius: var(--r-md);
  font-size: 12.5px;
  z-index: 6;
}
.crumb {
  display: flex; align-items: center; gap: 2px;
  background: none; border: none;
  color: var(--text-muted);
  padding: 4px 6px;
  border-radius: var(--r-sm);
  transition: all var(--t-fast) var(--ease);
}
.crumb:hover { color: var(--text-secondary); background: rgba(255, 255, 255, 0.05); }
.crumb.active { color: var(--text-primary); font-weight: 600; }
.crumb__sep { opacity: 0.5; }

.desktop__grid {
  position: absolute;
  top: 36px; left: 0; right: 0; bottom: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(104px, 1fr));
  grid-auto-rows: 116px;
  gap: 8px;
  align-content: start;
  padding: 8px 4px 60px;
  overflow-y: auto;
}

.file-item {
  position: relative;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 8px;
  padding: 10px 6px;
  border-radius: var(--r-md);
  border: 1px solid transparent;
  cursor: default;
  transition: all var(--t-fast) var(--ease);
}
.file-item:hover { background: rgba(255, 255, 255, 0.05); border-color: rgba(255, 255, 255, 0.06); }
.file-item.selected {
  background: var(--info-100);
  border-color: var(--info-300);
  box-shadow: 0 0 0 1px var(--info-300);
}
.file-item__icon { position: relative; color: var(--text-secondary); transition: color var(--t-med) var(--ease); }
.file-item:hover .file-item__icon,
.file-item.selected .file-item__icon { color: var(--accent-hover); }
.file-item__pin { position: absolute; top: -2px; right: -4px; color: var(--amber); background: var(--surface-1); border-radius: var(--r-full); padding: 1px; }
.file-item__name {
  font-size: 12px; text-align: center;
  max-width: 100%;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  color: var(--text-secondary);
}
.file-item.selected .file-item__name { color: var(--text-primary); }

.file-item__actions {
  position: absolute;
  bottom: 6px; left: 50%; transform: translateX(-50%);
  display: flex; gap: 4px;
  animation: fadeIn var(--t-fast) var(--ease);
}
.mini {
  width: 24px; height: 24px;
  border-radius: var(--r-sm);
  border: none;
  background: rgba(0, 0, 0, 0.5);
  color: var(--text-primary);
  display: grid; place-items: center;
  transition: all var(--t-fast) var(--ease);
}
.mini:hover { background: var(--accent); color: #fff; }

.desktop__empty {
  position: absolute;
  top: 50%; left: 50%; transform: translate(-50%, -50%);
  text-align: center;
}
.desktop__empty-title { margin: 0; font-size: 16px; color: var(--text-secondary); font-weight: 600; }
.desktop__empty-sub { margin: 6px 0 0; font-size: 13px; color: var(--text-muted); }

.hidden-toggle {
  position: absolute;
  bottom: 4px; right: 4px;
  padding: 5px 12px;
  border-radius: var(--r-full);
  border: 1px solid var(--frost-border);
  background: var(--frost-bg);
  backdrop-filter: blur(12px);
  color: var(--text-muted);
  font-size: 11px;
  transition: all var(--t-fast) var(--ease);
}
.hidden-toggle:hover { color: var(--text-primary); }
.hidden-toggle.on { color: var(--accent-hover); border-color: var(--info-300); }
</style>
