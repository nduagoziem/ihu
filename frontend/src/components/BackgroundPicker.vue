<script setup>
import { ref, computed } from 'vue'
import { X, Check } from '@lucide/vue'
import bgImage1 from '../assets/images/app-bg-img-1.jpg'
import bgImage2 from '../assets/images/app-bg-img-2.jpg'

const props = defineProps({ config: Object })
const emit = defineEmits(['apply', 'close'])

const gradients = [
  { id: 'deep-sea', name: 'Deep Sea', css: 'radial-gradient(120% 120% at 0% 0%, #14202e 0%, #0a0c10 55%, #050608 100%)' },
  { id: 'aurora', name: 'Aurora', css: 'radial-gradient(100% 100% at 20% 10%, #0e2a33 0%, #0a1a22 50%, #060a10 100%)' },
  { id: 'ember', name: 'Ember', css: 'radial-gradient(120% 120% at 80% 20%, #2a1410 0%, #1a0c0a 55%, #080404 100%)' },
  { id: 'forest', name: 'Forest', css: 'radial-gradient(120% 120% at 30% 80%, #102214 0%, #0a1410 55%, #040805 100%)' },
  { id: 'graphite', name: 'Graphite', css: 'linear-gradient(135deg, #1a1d24 0%, #101319 100%)' },
  { id: 'steel', name: 'Steel Blue', css: 'radial-gradient(120% 120% at 70% 30%, #16242e 0%, #0c1620 55%, #060a10 100%)' },
]

const presets = [
  bgImage1,
  bgImage2,
]

const selected = ref(resolvePreset(props.config?.backgroundImage || ''))
const mode = ref(props.config?.backgroundMode || 'gradient')
const selectedGradient = ref(gradients.find((g) => g.css === props.config?.backgroundImage) || gradients[0])

function pickGradient(g) {
  selectedGradient.value = g
  selected.value = g.css
  mode.value = 'gradient'
}
function pickImage(url) {
  selected.value = url
  selectedGradient.value = null
  mode.value = 'cover'
}
function apply() {
  emit('apply', selected.value, mode.value)
}
function resolvePreset(value) {
  if (value === '../assets/images/app-bg-img-1.jpg') return bgImage1
  if (value === '../assets/images/app-bg-img-2.jpg') return bgImage2
  return value
}
</script>

<template>
  <div class="picker-overlay" @click.self="emit('close')">
    <div class="picker glass-strong">
      <div class="picker__header">
        <span class="picker__title">Background</span>
        <button class="picker__close" @click="emit('close')"><X :size="18" /></button>
      </div>

      <div class="picker__section">
        <div class="picker__label">Solid & gradient aesthetics</div>
        <div class="picker__grid">
          <button
            v-for="g in gradients"
            :key="g.id"
            class="swatch"
            :class="{ active: selectedGradient?.id === g.id }"
            :style="{ background: g.css }"
            @click="pickGradient(g)"
          >
            <Check v-if="selectedGradient?.id === g.id" :size="18" class="swatch__check" />
            <span class="swatch__name">{{ g.name }}</span>
          </button>
        </div>
      </div>

      <div class="picker__section">
        <div class="picker__label">Photo backgrounds</div>
        <div class="picker__grid picker__grid--photos">
          <button
            v-for="(url, i) in presets"
            :key="url"
            class="photo"
            :class="{ active: selected === url }"
            :style="{ backgroundImage: `linear-gradient(135deg, rgba(5, 8, 12, 0.58), rgba(20, 32, 46, 0.42)), url('${url}')` }"
            @click="pickImage(url)"
          >
            <Check v-if="selected === url" :size="20" class="photo__check" />
          </button>
        </div>
      </div>

      <div class="picker__footer">
        <span class="picker__hint">Press Shift + I anytime to open this.</span>
        <button class="picker__apply" @click="apply">Apply</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.picker-overlay {
  position: fixed; inset: 0; z-index: 82;
  display: grid; place-items: center;
  background: rgba(5, 6, 8, 0.5);
  animation: fadeIn var(--t-med) var(--ease);
}
.picker {
  width: min(620px, 92vw);
  max-height: 84vh;
  border-radius: var(--r-xl);
  display: flex; flex-direction: column;
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-slow) var(--ease-out);
}
.picker__header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
}
.picker__title { font-weight: 700; font-size: 15px; }
.picker__close {
  width: 32px; height: 32px; border-radius: var(--r-sm); border: none;
  background: transparent; color: var(--text-muted);
  display: grid; place-items: center;
}
.picker__close:hover { background: rgba(255,255,255,0.08); color: var(--text-primary); }

.picker__section { padding: 16px 20px; }
.picker__label {
  font-size: 11px; text-transform: uppercase; letter-spacing: 0.08em;
  color: var(--text-muted); margin-bottom: 12px;
}
.picker__grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
.picker__grid--photos { grid-template-columns: repeat(4, 1fr); }
.swatch {
  position: relative;
  height: 76px;
  border-radius: var(--r-md);
  border: 2px solid transparent;
  cursor: pointer;
  overflow: hidden;
  transition: all var(--t-fast) var(--ease);
}
.swatch:hover { transform: translateY(-2px); }
.swatch.active { border-color: var(--accent); box-shadow: 0 0 0 2px var(--accent-soft); }
.swatch__check {
  position: absolute; top: 8px; right: 8px;
  color: #fff; filter: drop-shadow(0 1px 2px rgba(0,0,0,0.6));
}
.swatch__name {
  position: absolute; bottom: 6px; left: 8px;
  font-size: 11px; color: #fff;
  text-shadow: 0 1px 2px rgba(0,0,0,0.7);
}
.photo {
  position: relative;
  height: 76px;
  border-radius: var(--r-md);
  border: 2px solid transparent;
  background-size: cover; background-position: center;
  cursor: pointer;
  transition: all var(--t-fast) var(--ease);
}
.photo:hover { transform: translateY(-2px); }
.photo.active { border-color: var(--accent); box-shadow: 0 0 0 2px var(--accent-soft); }
.photo__check {
  position: absolute; top: 8px; right: 8px;
  color: #fff; filter: drop-shadow(0 1px 2px rgba(0,0,0,0.7));
}

.picker__footer {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.07);
}
.picker__hint { font-size: 11px; color: var(--text-muted); }
.picker__apply {
  padding: 8px 22px;
  border-radius: var(--r-md);
  border: none;
  background: var(--accent);
  color: #fff;
  font-weight: 600; font-size: 13px;
  transition: all var(--t-fast) var(--ease);
}
.picker__apply:hover { background: var(--accent-hover); transform: translateY(-1px); }
</style>
