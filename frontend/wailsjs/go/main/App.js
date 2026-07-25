// @ts-check
// Wails bindings for the App struct. Generated bindings extended for ihu.
// When running inside Wails these delegate to the Go backend; otherwise they
// fall back to an in-memory mock so the UI is fully explorable standalone.

const isWails = typeof window !== 'undefined' && window.go && window.go.main && window.go.main.App

const go = isWails ? window.go.main.App : null

export function Greet(arg1) {
  if (go) return go.Greet(arg1)
  return Promise.resolve(`Hello ${arg1}, It's show time!`)
}

export function GetBootData() {
  if (go) return go.GetBootData()
  return Promise.resolve({
    systemStats: {
      arch: 'x86_64',
      distro: 'Ubuntu 24.04 LTS',
      kernel: '6.6.36.3-microsoft',
      cpuUsage: 18,
      memoryUsage: 42,
      diskUsage: 55,
      temperature: 52,
      networkStatus: 'active',
      timestamp: new Date().toString(),
    },
    bootedAt: new Date().toString(),
  })
}

export function GetConfig() {
  if (go) return go.GetConfig()
  return Promise.resolve({
    welcomeDisabled: false,
    defaultLinuxPath: '/root',
    defaultLinuxUser: 'root',
    defaultLinuxDistro: '',
    pinnedFolders: [],
    backgroundImage: '',
    backgroundMode: 'gradient',
  })
}

export function SetWelcomeDisabled(arg1) {
  if (go) return go.SetWelcomeDisabled(arg1)
  return Promise.resolve({ ...mockConfig(), welcomeDisabled: arg1 })
}

export function SetDefaultLinuxPath(arg1) {
  if (go) return go.SetDefaultLinuxPath(arg1)
  return Promise.resolve({ ...mockConfig(), defaultLinuxPath: arg1 })
}

export function SetDefaultLinuxUser(arg1) {
  if (go) return go.SetDefaultLinuxUser(arg1)
  return Promise.resolve({ ...mockConfig(), defaultLinuxUser: arg1 })
}

export function SetDefaultLinuxDistro(arg1) {
  if (go) return go.SetDefaultLinuxDistro(arg1)
  return Promise.resolve({ ...mockConfig(), defaultLinuxDistro: arg1 })
}

export function TogglePinnedFolder(arg1) {
  if (go) return go.TogglePinnedFolder(arg1)
  const cfg = mockConfig()
  const i = cfg.pinnedFolders.indexOf(arg1)
  if (i >= 0) cfg.pinnedFolders.splice(i, 1)
  else cfg.pinnedFolders.push(arg1)
  return Promise.resolve(cfg)
}

export function SetBackground(arg1, arg2) {
  if (go) return go.SetBackground(arg1, arg2)
  return Promise.resolve({ ...mockConfig(), backgroundImage: arg1, backgroundMode: arg2 })
}

export function ListDir(arg1) {
  if (go) return go.ListDir(arg1)
  return Promise.resolve(mockDir(arg1))
}

export function HomePath(arg1) {
  if (go) return go.HomePath(arg1)
  const user = arg1 || 'root'
  return Promise.resolve(user === 'root' ? '/root' : `/home/${user}`)
}

export function ListDistros() {
  if (go) return go.ListDistros()
  return Promise.resolve(['Ubuntu', 'Debian', 'Kali-Linux'])
}

export function ListUsers() {
  if (go) return go.ListUsers()
  return Promise.resolve(['root', 'nduagoziem', 'guest'])
}

export function ReadFile(arg1) {
  if (go) return go.ReadFile(arg1)
  return Promise.resolve(mockFile(arg1))
}

// --- mock helpers (non-Wails only) ----------------------------------------

function mockConfig() {
  return {
    welcomeDisabled: false,
    defaultLinuxPath: '/root',
    defaultLinuxUser: 'root',
    defaultLinuxDistro: '',
    pinnedFolders: [],
    backgroundImage: '',
    backgroundMode: 'gradient',
  }
}

function mockDir(dir) {
  const base = dir || '/root'
  const mk = (name, isDir, extra = {}) => ({
    name,
    path: base === '/' ? `/${name}` : `${base}/${name}`,
    isDir,
    isSymlink: false,
    isHidden: name.startsWith('.'),
    size: isDir ? 0 : 2048,
    modified: '2025-07-24 12:00',
    perm: isDir ? 'drwxr-xr-x' : '-rw-r--r--',
    ...extra,
  })
  return [
    mk('Documents', true),
    mk('Downloads', true),
    mk('Projects', true),
    mk('Pictures', true),
    mk('.bashrc', false, { size: 3821 }),
    mk('.profile', false, { size: 887 }),
    mk('notes.md', false, { size: 1432 }),
    mk('todo.txt', false, { size: 612 }),
    mk('script.sh', false, { size: 921 }),
    mk('photo.png', false, { size: 84210 }),
    mk('report.pdf', false, { size: 201334 }),
  ]
}

function mockFile(path) {
  const name = (path || '').split('/').pop() || ''
  if (name.endsWith('.md')) {
    return Promise.resolve(`# ${name.replace('.md', '')}\n\nThis is a sample markdown document rendered by ihu.\n\n## Features\n\n- File browsing\n- Terminal\n- Editor\n`)
  }
  if (name.endsWith('.sh')) {
    return Promise.resolve(`#!/usr/bin/env bash\nset -euo pipefail\n\necho "Hello from ${name}"\nfor i in {1..5}; do\n  echo "  line $i"\ndone\n`)
  }
  return Promise.resolve(`Contents of ${name}\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit.\n`)
}
