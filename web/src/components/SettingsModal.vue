<template>
  <div class="modal-backdrop" :class="{ open }" @click.self="close">
    <div
      class="modal sm-root"
      role="dialog"
      aria-modal="true"
      aria-labelledby="sm-dialog-title"
    >
      <header class="sm-head">
        <h1 id="sm-dialog-title" class="sm-title">Settings</h1>
        <Transition name="sm-fade">
          <p
            v-if="autosaveLine"
            class="sm-autosave"
            :class="{
              'sm-autosave--muted': autosaveTone === 'muted',
              'sm-autosave--ok': autosaveTone === 'ok',
              'sm-autosave--err': autosaveTone === 'err',
            }"
            role="status"
            aria-live="polite"
          >
            {{ autosaveLine }}
          </p>
        </Transition>
      </header>

      <div class="sm-main">
        <aside class="sm-aside" aria-label="Settings categories">
          <SettingsNav v-model="activeTab" :tabs="SETTINGS_TABS" />
        </aside>

        <div
          class="sm-panel-host"
          :aria-labelledby="'settings-tab-' + activeTab"
        >
          <div
            v-show="activeTab === 'monitor'"
            id="settings-panel-monitor"
            class="sm-panel"
            role="tabpanel"
            tabindex="0"
            aria-labelledby="settings-tab-monitor"
          >
            <SettingsMonitorTab
              v-model:ping-targets="pingTargets"
              v-model:dns-targets="dnsTargets"
              v-model:ping-interval="pingInterval"
              v-model:speed-interval="speedInterval"
              v-model:ping-count="pingCount"
            />
          </div>

          <div
            v-show="activeTab === 'about'"
            id="settings-panel-about"
            class="sm-panel"
            role="tabpanel"
            tabindex="0"
            aria-labelledby="settings-tab-about"
          >
            <SettingsAboutTab
              :version-info="versionInfo"
              :checking="checking"
              :updating="updating"
              :update-done="updateDone"
              :update-error="updateError"
              @check="checkUpdate"
              @update="doUpdate"
              @reload="reloadPage"
            />
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, watch, onUnmounted } from 'vue'
import SettingsNav from './settings/SettingsNav.vue'
import SettingsMonitorTab from './settings/SettingsMonitorTab.vue'
import SettingsAboutTab from './settings/SettingsAboutTab.vue'
import { SETTINGS_TABS, DEFAULT_SETTINGS_TAB } from './settings/settingsTabMeta.js'

const AUTOSAVE_DEBOUNCE_MS = 550
/** Minimum time "Saving..." stays visible before switching to "Saved" */
const AUTOSAVE_SAVING_MIN_MS = 450
/** How long "Saved" stays visible before fading out */
const AUTOSAVE_SAVED_CLEAR_MS = 4000
const AUTOSAVE_ERR_CLEAR_MS = 5000

const props = defineProps({ open: Boolean })
const emit = defineEmits(['close', 'saved'])

const activeTab = ref(DEFAULT_SETTINGS_TAB)

const pingTargets = ref('')
const dnsTargets = ref('')
const pingInterval = ref(60)
const speedInterval = ref(30)
const pingCount = ref(5)

/** '', 'muted' (Saving…), 'ok', 'err' */
const autosaveTone = ref('')
const autosaveLine = ref('')

let lastSavedSerialized = ''
let autosaveClearTimer = null
let autosaveDebounceTimer = null
let saveGeneration = 0

function serializeConfig() {
  const pt = pingTargets.value.split('\n').map(s => s.trim()).filter(Boolean)
  const dt = dnsTargets.value.split('\n').map(s => s.trim()).filter(Boolean)
  const pis = parseInt(pingInterval.value, 10)
  const sim = parseInt(speedInterval.value, 10)
  const pc = parseInt(pingCount.value, 10)
  return JSON.stringify({ pt, dt, pis, sim, pc })
}

function scheduleAutosaveClear(ms) {
  clearTimeout(autosaveClearTimer)
  autosaveClearTimer = setTimeout(() => {
    autosaveLine.value = ''
    autosaveTone.value = ''
  }, ms)
}

function showAutosaveError(msg) {
  autosaveLine.value = msg
  autosaveTone.value = 'err'
  scheduleAutosaveClear(AUTOSAVE_ERR_CLEAR_MS)
}

async function load() {
  try {
    const cfg = await fetch('/api/config').then(r => r.json())
    pingTargets.value = (cfg.ping_targets || []).join('\n')
    dnsTargets.value = (cfg.dns_targets || []).join('\n')
    pingInterval.value = cfg.ping_interval_s || 60
    speedInterval.value = cfg.speed_interval_m || 30
    pingCount.value = cfg.ping_count || 5
    lastSavedSerialized = serializeConfig()
  } catch {
    showAutosaveError('Failed to load config')
  }
}

async function persistConfig() {
  if (!props.open) return

  const cur = serializeConfig()
  if (cur === lastSavedSerialized) return

  const pt = pingTargets.value.split('\n').map(s => s.trim()).filter(Boolean)
  const dt = dnsTargets.value.split('\n').map(s => s.trim()).filter(Boolean)
  const pis = parseInt(pingInterval.value, 10)
  const sim = parseInt(speedInterval.value, 10)
  const pc = parseInt(pingCount.value, 10)

  if (!pt.length) {
    showAutosaveError('Probe hosts cannot be empty')
    return
  }
  if (pis < 10) {
    showAutosaveError('Ping interval must be ≥ 10 s')
    return
  }
  if (sim < 5) {
    showAutosaveError('Speed interval must be ≥ 5 min')
    return
  }
  if (pc < 1 || pc > 20) {
    showAutosaveError('Ping count must be 1–20')
    return
  }

  const gen = ++saveGeneration
  clearTimeout(autosaveClearTimer)
  autosaveLine.value = 'Saving...'
  autosaveTone.value = 'muted'

  const savingStarted = Date.now()

  try {
    const res = await fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ping_targets: pt,
        dns_targets: dt,
        ping_interval_s: pis,
        speed_interval_m: sim,
        ping_count: pc,
      }),
    })
    const body = await res.json().catch(() => ({}))
    if (gen !== saveGeneration) return

    if (res.ok) {
      const elapsed = Date.now() - savingStarted
      const pad = Math.max(0, AUTOSAVE_SAVING_MIN_MS - elapsed)
      if (pad > 0) {
        await new Promise(r => setTimeout(r, pad))
      }
      if (gen !== saveGeneration) return

      lastSavedSerialized = serializeConfig()
      autosaveLine.value = 'Saved'
      autosaveTone.value = 'ok'
      emit('saved')
      scheduleAutosaveClear(AUTOSAVE_SAVED_CLEAR_MS)
    } else {
      showAutosaveError(body.error || 'Save failed')
    }
  } catch {
    if (gen !== saveGeneration) return
    showAutosaveError('Network error')
  }
}

function queueAutosave() {
  if (!props.open) return
  clearTimeout(autosaveDebounceTimer)
  autosaveDebounceTimer = setTimeout(persistConfig, AUTOSAVE_DEBOUNCE_MS)
}

watch(
  [pingTargets, dnsTargets, pingInterval, speedInterval, pingCount],
  () => {
    if (!props.open) return
    const cur = serializeConfig()
    if (cur === lastSavedSerialized) return
    queueAutosave()
  }
)

const versionInfo = ref(null)
const checking = ref(false)
const updating = ref(false)
const updateDone = ref(false)
const updateError = ref('')

async function checkUpdate() {
  checking.value = true
  try {
    versionInfo.value = await fetch('/api/version').then(r => r.json())
  } catch {
    /* ignore */
  }
  checking.value = false
}

async function doUpdate() {
  updating.value = true
  updateError.value = ''
  const targetVersion = versionInfo.value?.latest
  try {
    const res = await fetch('/api/update', { method: 'POST' })
    const body = await res.json()
    if (!res.ok) {
      updating.value = false
      updateError.value = body.error || 'Update failed'
      return
    }
    if (body.status === 'up_to_date') {
      updating.value = false
      return
    }
  } catch {
    /* server may have restarted */
  }

  let attempts = 0
  const poll = setInterval(async () => {
    if (++attempts > 30) {
      clearInterval(poll)
      updating.value = false
      updateError.value = 'Update timed out. Restart the server manually.'
      return
    }
    try {
      const res = await fetch('/api/version').then(r => r.json())
      if (res.current === targetVersion) {
        clearInterval(poll)
        versionInfo.value = res
        updating.value = false
        updateDone.value = true
      }
    } catch {
      /* still restarting */
    }
  }, 3000)
}

function reloadPage() {
  window.location.reload()
}

function runTabOpenHook(tabId) {
  if (tabId === 'about' && !versionInfo.value) checkUpdate()
}

function onDocKeydown(e) {
  if (e.key === 'Escape' && props.open) {
    e.preventDefault()
    close()
  }
}

watch(
  () => props.open,
  v => {
    if (v) {
      load()
      activeTab.value = DEFAULT_SETTINGS_TAB
      versionInfo.value = null
      updateDone.value = false
      updateError.value = ''
      document.addEventListener('keydown', onDocKeydown)
    } else {
      document.removeEventListener('keydown', onDocKeydown)
    }
  }
)

watch(activeTab, (tabId, prev) => {
  if (tabId === prev) return
  runTabOpenHook(tabId)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onDocKeydown)
})

function close() {
  clearTimeout(autosaveDebounceTimer)
  clearTimeout(autosaveClearTimer)
  saveGeneration++
  autosaveLine.value = ''
  autosaveTone.value = ''
  emit('close')
}
</script>

<style scoped>
/* One height for all tabs — body scrolls inside .sm-panel-host */
.sm-root {
  --sm-fixed-height: min(580px, calc(100vh - 2rem));

  width: 100%;
  max-width: 760px !important;
  height: var(--sm-fixed-height) !important;
  min-height: var(--sm-fixed-height) !important;
  max-height: var(--sm-fixed-height) !important;

  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important;
  border-radius: 14px !important;
  border: 1px solid var(--border) !important;
  background: var(--card) !important;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.04) inset,
    0 24px 56px rgba(0, 0, 0, 0.2) !important;
}

.sm-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
  padding: 1rem 1.25rem;
  flex-shrink: 0;
  border-bottom: 1px solid var(--border);
}

.sm-title {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: var(--fg);
  line-height: 1.2;
}

.sm-autosave {
  margin: 0;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.01em;
  min-height: 1.2em;
}

.sm-autosave--muted {
  color: var(--muted);
}

.sm-autosave--ok {
  color: var(--green);
}

.sm-autosave--err {
  color: var(--red);
}

.sm-main {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: row;
  align-items: stretch;
}

.sm-aside {
  width: 168px;
  flex-shrink: 0;
  border-right: 1px solid var(--border);
  background: color-mix(in srgb, var(--bg) 88%, var(--card));
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.sm-aside :deep(.sn) {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-gutter: stable;
}

.sm-panel-host {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-gutter: stable;
}

.sm-panel {
  padding: 1.125rem 1.25rem 1.25rem;
  outline: none;
}

.sm-fade-enter-active,
.sm-fade-leave-active {
  transition: opacity 0.15s ease;
}

.sm-fade-enter-from,
.sm-fade-leave-to {
  opacity: 0;
}

@media (max-width: 560px) {
  .sm-main {
    flex-direction: column;
  }

  .sm-aside {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
    overflow: hidden;
  }

  .sm-aside :deep(.sn) {
    overflow-y: hidden;
  }
}
</style>
