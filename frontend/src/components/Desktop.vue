<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { Folder, FileText, FileCode, FileImage, File, Pin, PinOff, Pencil, Eye, ChevronRight, Trash2, FolderOpen, Loader2 } from '@lucide/vue'
import * as App from '../../wailsjs/go/main/App'
import { useToast } from '../composables/useToast'

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
const contextMenu = ref({ open: false, x: 0, y: 0, entry: null })
const pendingDelete = ref(null)
const deleting = ref(false)
const { notify } = useToast()
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
  closeContextMenu()
  emit('toggle-pin', entry.path)
}

function clearSelection() {
  selected.value = null
}

function showContextMenu(entry, e) {
  selected.value = entry
  selectedAt.value = { x: e.clientX, y: e.clientY, time: Date.now() }
  const width = 178
  const height = entry.isDir ? 108 : 148
  contextMenu.value = {
    open: true,
    x: Math.min(e.clientX, window.innerWidth - width - 8),
    y: Math.min(e.clientY, window.innerHeight - height - 8),
    entry,
  }
}

function closeContextMenu() {
  contextMenu.value.open = false
}

function openEditor(entry) {
  closeContextMenu()
  emit('open-editor', entry)
}

function openViewer(entry) {
  closeContextMenu()
  emit('open-viewer', entry)
}

function requestDelete(entry) {
  closeContextMenu()
  pendingDelete.value = entry
}

function cancelDelete() {
  if (deleting.value) return
  pendingDelete.value = null
}

async function confirmDelete() {
  if (!pendingDelete.value || deleting.value) return
  const entry = pendingDelete.value
  deleting.value = true
  try {
    const elevated = props.superUser || props.currentUser === 'root'
    if (entry.isDir) {
      await App.DeleteDirAs(entry.path, props.currentDistro || '', props.currentUser || '', elevated)
    } else {
      await App.DeleteFileAs(entry.path, props.currentDistro || '', props.currentUser || '', elevated)
    }
    if (entry.isDir && isPinned(entry.path)) {
      emit('toggle-pin', entry.path)
    }
    entries.value = entries.value.filter((item) => item.path !== entry.path)
    pendingDelete.value = null
    notify(`${entry.name} deleted`, 'success')
    await load()
  } catch (e) {
    notify(`Could not delete ${entry.name}: ${String(e?.message || e || 'unknown error')}`)
  } finally {
    deleting.value = false
  }
}

</script>

<template>
  <div class="desktop" @click="closeContextMenu" @click.self="clearSelection" @contextmenu.prevent="closeContextMenu">
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
        @contextmenu.prevent.stop="showContextMenu(entry, $event)"
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
          <button v-if="!entry.isDir" class="mini" title="Open in editor" @click.stop="openEditor(entry)">
            <Pencil :size="13" />
          </button>
          <button v-if="!entry.isDir" class="mini" title="Open in viewer" @click.stop="openViewer(entry)">
            <Eye :size="13" />
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="contextMenu.open"
      class="context-menu glass-strong"
      :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
      @click.stop
      @contextmenu.prevent.stop
    >
      <button v-if="contextMenu.entry?.isDir" class="context-menu__item" @click="openEntry(contextMenu.entry); closeContextMenu()">
        <FolderOpen :size="15" />
        <span>Open</span>
      </button>
      <button v-if="contextMenu.entry?.isDir" class="context-menu__item" @click="togglePin(contextMenu.entry)">
        <component :is="isPinned(contextMenu.entry.path) ? PinOff : Pin" :size="15" />
        <span>{{ isPinned(contextMenu.entry.path) ? 'Unpin' : 'Pin' }}</span>
      </button>
      <button v-if="!contextMenu.entry?.isDir" class="context-menu__item" @click="openEditor(contextMenu.entry)">
        <Pencil :size="15" />
        <span>Edit</span>
      </button>
      <button v-if="!contextMenu.entry?.isDir" class="context-menu__item" @click="openViewer(contextMenu.entry)">
        <Eye :size="15" />
        <span>Preview</span>
      </button>
      <div class="context-menu__sep"></div>
      <button class="context-menu__item context-menu__item--danger" @click="requestDelete(contextMenu.entry)">
        <Trash2 :size="15" />
        <span>Delete</span>
      </button>
    </div>

    <div v-if="pendingDelete" class="delete-confirm" @click.self="cancelDelete">
      <div class="delete-confirm__panel glass-strong" @click.stop>
        <div class="delete-confirm__icon">
          <Trash2 :size="22" />
        </div>
        <div class="delete-confirm__copy">
          <h2>Delete {{ pendingDelete.isDir ? 'folder' : 'file' }}?</h2>
          <p>{{ pendingDelete.name }}</p>
          <span v-if="pendingDelete.isDir">This will remove the folder and everything inside it.</span>
          <span v-else>This will remove the file from WSL.</span>
        </div>
        <div class="delete-confirm__actions">
          <button class="delete-confirm__btn" :disabled="deleting" @click="cancelDelete">Cancel</button>
          <button class="delete-confirm__btn delete-confirm__btn--danger" :disabled="deleting" @click="confirmDelete">
            <Loader2 v-if="deleting" :size="14" class="spin" />
            <Trash2 v-else :size="14" />
            {{ deleting ? 'Deleting' : 'Delete' }}
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

.context-menu {
  position: fixed;
  z-index: 82;
  width: 178px;
  padding: 6px;
  border-radius: var(--r-md);
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-fast) var(--ease-out);
}
.context-menu__item {
  width: 100%;
  height: 32px;
  display: flex; align-items: center; gap: 9px;
  padding: 0 9px;
  border: none;
  border-radius: var(--r-sm);
  background: transparent;
  color: var(--text-secondary);
  font-size: 12.5px;
  text-align: left;
}
.context-menu__item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-primary);
}
.context-menu__item--danger { color: var(--rose); }
.context-menu__item--danger:hover {
  background: rgba(251, 113, 133, 0.14);
  color: #fecdd3;
}
.context-menu__sep {
  height: 1px;
  margin: 5px 3px;
  background: rgba(255, 255, 255, 0.08);
}

.delete-confirm {
  position: fixed;
  inset: 0;
  z-index: 84;
  display: grid;
  place-items: center;
  background: rgba(5, 6, 8, 0.5);
  backdrop-filter: blur(18px) saturate(160%);
  -webkit-backdrop-filter: blur(18px) saturate(160%);
  animation: fadeIn var(--t-med) var(--ease);
}
.delete-confirm__panel {
  width: min(380px, calc(100vw - 32px));
  border-radius: var(--r-lg);
  padding: 22px;
  display: grid;
  grid-template-columns: 46px 1fr;
  gap: 14px;
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-med) var(--ease-out);
}
.delete-confirm__icon {
  width: 42px; height: 42px;
  border-radius: var(--r-md);
  display: grid; place-items: center;
  background: rgba(251, 113, 133, 0.14);
  color: var(--rose);
}
.delete-confirm__copy {
  min-width: 0;
}
.delete-confirm__copy h2 {
  margin: 0;
  font-size: 17px;
  line-height: 1.25;
  color: var(--text-primary);
}
.delete-confirm__copy p {
  margin: 8px 0 3px;
  color: var(--text-secondary);
  font-size: 13px;
  overflow-wrap: anywhere;
}
.delete-confirm__copy span {
  color: var(--text-muted);
  font-size: 12.5px;
}
.delete-confirm__actions {
  grid-column: 1 / -1;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}
.delete-confirm__btn {
  height: 32px;
  display: flex; align-items: center; gap: 7px;
  padding: 0 13px;
  border-radius: var(--r-sm);
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-secondary);
  font-size: 12.5px;
}
.delete-confirm__btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}
.delete-confirm__btn--danger {
  border-color: rgba(251, 113, 133, 0.28);
  background: rgba(251, 113, 133, 0.16);
  color: #fecdd3;
}
.delete-confirm__btn--danger:hover:not(:disabled) {
  background: rgba(251, 113, 133, 0.24);
}
.delete-confirm__btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

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
