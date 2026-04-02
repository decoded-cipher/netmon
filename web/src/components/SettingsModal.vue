<template>
  <div class="modal-backdrop" :class="{ open }" @click.self="close">
    <div class="modal settings-modal" role="dialog" aria-modal="true">

      <div class="modal-header">
        <h2>Settings</h2>
        <button class="modal-close" @click="close">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <div class="settings-layout">
        <!-- Sidebar nav -->
        <nav class="settings-nav">
          <button
            v-for="tab in tabs" :key="tab.id"
            class="settings-nav-item"
            :class="{ active: activeTab === tab.id }"
            @click="activeTab = tab.id"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="tab.icon" />
            {{ tab.label }}
          </button>
        </nav>

        <!-- Monitor tab -->
        <div v-if="activeTab === 'monitor'" class="settings-pane">
          <div>
            <div class="modal-section-title">
              <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>
              Internet Probes
            </div>
            <div class="targets-grid">
              <div class="field-group">
                <label class="field-label" for="cfgPingTargets">Probe Hosts</label>
                <textarea class="field-input" id="cfgPingTargets" rows="4" placeholder="google.com&#10;cloudflare.com" spellcheck="false" v-model="pingTargets" />
                <div class="field-hint">Pinged to test your internet connection. Your gateway is always added.</div>
              </div>
              <div class="field-group">
                <label class="field-label" for="cfgDnsTargets">DNS Probe Hosts</label>
                <textarea class="field-input" id="cfgDnsTargets" rows="4" placeholder="google.com&#10;cloudflare.com" spellcheck="false" v-model="dnsTargets" />
                <div class="field-hint">Resolved to measure DNS lookup speed on your network.</div>
              </div>
            </div>
          </div>

          <div>
            <div class="modal-section-title">
              <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
              Intervals &amp; Probing
            </div>
            <div class="intervals-grid">
              <div class="interval-card" @click="$el.querySelector('input').focus()">
                <div class="interval-card-label">Ping every</div>
                <div class="interval-input-row">
                  <input class="interval-input" type="number" min="10" max="3600" v-model="pingInterval" />
                  <span class="interval-unit">sec</span>
                </div>
                <div class="field-hint" style="margin-top:0">Min 10 s</div>
              </div>
              <div class="interval-card" @click="$el.querySelector('input').focus()">
                <div class="interval-card-label">Speed test every</div>
                <div class="interval-input-row">
                  <input class="interval-input" type="number" min="5" max="1440" v-model="speedInterval" />
                  <span class="interval-unit">min</span>
                </div>
                <div class="field-hint" style="margin-top:0">Min 5 min</div>
              </div>
              <div class="interval-card" @click="$el.querySelector('input').focus()">
                <div class="interval-card-label">Pings per cycle</div>
                <div class="interval-input-row">
                  <input class="interval-input" type="number" min="1" max="20" v-model="pingCount" />
                  <span class="interval-unit">pings</span>
                </div>
                <div class="field-hint" style="margin-top:0">1 – 20</div>
              </div>
            </div>
          </div>
        </div>

        <!-- About tab -->
        <div v-else-if="activeTab === 'about'" class="settings-pane">
          <!-- App identity -->
          <div class="about-app">
            <div class="about-app-icon">
              <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><circle cx="12" cy="20" r="1" fill="currentColor" stroke="none"/></svg>
            </div>
            <div>
              <div class="about-app-name">netmon</div>
              <div class="about-app-desc">Lightweight self-hosted network monitoring</div>
            </div>
          </div>

          <div class="about-divider" />

          <!-- Version info -->
          <div class="about-version-row">
            <div>
              <div class="about-version-label">Current version</div>
              <div class="about-version-value">{{ versionInfo?.current || '—' }}</div>
              <div v-if="versionInfo" class="about-update-status" :class="updateStatusClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" v-html="updateStatusIcon" />
                {{ updateStatusText }}
              </div>
              <div v-else-if="checking" class="about-update-status muted">
                Checking for updates…
              </div>
            </div>
            <div v-if="versionInfo?.update_available">
              <div class="about-version-label">Latest</div>
              <div class="about-version-value" style="color: var(--accent)">{{ versionInfo.latest }}</div>
              <a
                :href="`https://github.com/decoded-cipher/netmon/releases/tag/${versionInfo.latest}`"
                target="_blank"
                rel="noopener"
                class="about-update-status avail"
                style="text-decoration:none; margin-top: 0.25rem;"
              >View release notes →</a>
            </div>
          </div>

          <!-- Update action -->
          <div v-if="updateDone" class="about-update-done">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
            Updated to {{ versionInfo?.latest }} — reload to apply
            <button class="btn btn-primary" style="margin-left: 0.5rem" @click="reloadPage">Reload</button>
          </div>
          <div v-else-if="updating" class="about-update-status muted" style="align-self:flex-start">
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="spin"><line x1="12" y1="2" x2="12" y2="6"/><line x1="12" y1="18" x2="12" y2="22"/><line x1="4.93" y1="4.93" x2="7.76" y2="7.76"/><line x1="16.24" y1="16.24" x2="19.07" y2="19.07"/><line x1="2" y1="12" x2="6" y2="12"/><line x1="18" y1="12" x2="22" y2="12"/><line x1="4.93" y1="19.07" x2="7.76" y2="16.24"/><line x1="16.24" y1="7.76" x2="19.07" y2="4.93"/></svg>
            Updating… server will restart momentarily
          </div>
          <div v-else-if="updateError" class="about-update-status" style="color:var(--red);align-self:flex-start">
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
            {{ updateError }}
          </div>
          <div v-else class="about-btn-row">
            <button
              v-if="versionInfo?.update_available"
              class="btn btn-primary"
              @click="doUpdate"
            >Update to {{ versionInfo.latest }}</button>
            <button class="btn btn-ghost" :disabled="checking" @click="checkUpdate">
              {{ checking ? 'Checking…' : 'Check for updates' }}
            </button>
          </div>

          <div class="about-divider" />

          <!-- Links -->
          <div>
            <div class="about-version-label" style="margin-bottom: 0.625rem">Source &amp; License</div>
            <div class="about-links">
              <a href="https://github.com/decoded-cipher/netmon" target="_blank" rel="noopener" class="about-link">
                <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"/></svg>
                GitHub
              </a>
              <a href="https://github.com/decoded-cipher/netmon/blob/master/LICENSE" target="_blank" rel="noopener" class="about-link">
                <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
                MIT License
              </a>
              <a href="https://github.com/decoded-cipher/netmon/releases" target="_blank" rel="noopener" class="about-link">
                <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
                Releases
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer — only for monitor tab -->
      <div v-if="activeTab === 'monitor'" class="modal-footer">
        <div class="modal-note">
          <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          Interval changes take effect after restart
        </div>
        <div class="modal-actions">
          <span class="save-status" :class="saveOk ? 'ok' : 'err'">{{ saveStatus }}</span>
          <button class="btn btn-ghost" @click="close" type="button">Cancel</button>
          <button class="btn btn-primary" :disabled="saving" @click="save" type="button">Save changes</button>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({ open: Boolean })
const emit  = defineEmits(['close', 'saved'])

const activeTab = ref('monitor')

const tabs = [
  {
    id: 'monitor',
    label: 'Monitor',
    icon: '<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>',
  },
  {
    id: 'about',
    label: 'About',
    icon: '<circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>',
  },
]

// ── Monitor tab state ────────────────────────────────────────────────────
const pingTargets   = ref('')
const dnsTargets    = ref('')
const pingInterval  = ref(60)
const speedInterval = ref(30)
const pingCount     = ref(5)
const saveStatus    = ref('')
const saveOk        = ref(false)
const saving        = ref(false)

async function load() {
  try {
    const cfg = await fetch('/api/config').then(r => r.json())
    pingTargets.value   = (cfg.ping_targets  || []).join('\n')
    dnsTargets.value    = (cfg.dns_targets   || []).join('\n')
    pingInterval.value  = cfg.ping_interval_s  || 60
    speedInterval.value = cfg.speed_interval_m || 30
    pingCount.value     = cfg.ping_count || 5
  } catch { showStatus('Failed to load config', false) }
}

async function save() {
  const pt  = pingTargets.value.split('\n').map(s => s.trim()).filter(Boolean)
  const dt  = dnsTargets.value.split('\n').map(s => s.trim()).filter(Boolean)
  const pis = parseInt(pingInterval.value, 10)
  const sim = parseInt(speedInterval.value, 10)
  const pc  = parseInt(pingCount.value, 10)

  if (!pt.length)        return showStatus('Probe hosts cannot be empty', false)
  if (pis < 10)          return showStatus('Ping interval must be ≥ 10 s', false)
  if (sim < 5)           return showStatus('Speed interval must be ≥ 5 min', false)
  if (pc < 1 || pc > 20) return showStatus('Ping count must be 1–20', false)

  saving.value = true
  try {
    const res  = await fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ping_targets: pt, dns_targets: dt, ping_interval_s: pis, speed_interval_m: sim, ping_count: pc }),
    })
    const body = await res.json()
    if (res.ok) {
      showStatus('Saved!', true)
      setTimeout(() => { close(); emit('saved') }, 900)
    } else {
      showStatus(body.error || 'Save failed', false)
    }
  } catch { showStatus('Network error', false) }
  saving.value = false
}

function showStatus(msg, ok) {
  saveStatus.value = msg
  saveOk.value = ok
}

// ── About tab state ──────────────────────────────────────────────────────
const versionInfo  = ref(null)
const checking     = ref(false)
const updating     = ref(false)
const updateDone   = ref(false)
const updateError  = ref('')

async function checkUpdate() {
  checking.value = true
  try {
    versionInfo.value = await fetch('/api/version').then(r => r.json())
  } catch {}
  checking.value = false
}

async function doUpdate() {
  updating.value   = true
  updateError.value = ''
  const targetVersion = versionInfo.value?.latest

  try {
    const res  = await fetch('/api/update', { method: 'POST' })
    const body = await res.json()
    if (!res.ok) {
      updating.value  = false
      updateError.value = body.error || 'Update failed'
      return
    }
    if (body.status === 'up_to_date') {
      updating.value = false
      return
    }
  } catch {
    // Server may have restarted before we got a response — continue polling.
  }

  // Poll until the server comes back running the new version.
  let attempts = 0
  const poll = setInterval(async () => {
    attempts++
    if (attempts > 30) {
      clearInterval(poll)
      updating.value    = false
      updateError.value = 'Update timed out. Restart the server manually if needed.'
      return
    }
    try {
      const res = await fetch('/api/version').then(r => r.json())
      if (res.current === targetVersion) {
        clearInterval(poll)
        versionInfo.value = res
        updating.value    = false
        updateDone.value  = true
      }
    } catch { /* server still restarting */ }
  }, 3000)
}

function reloadPage() {
  window.location.reload()
}

const updateStatusClass = computed(() => {
  if (!versionInfo.value) return 'muted'
  return versionInfo.value.update_available ? 'avail' : 'ok'
})

const updateStatusText = computed(() => {
  if (!versionInfo.value) return ''
  return versionInfo.value.update_available
    ? `Update available: ${versionInfo.value.latest}`
    : 'Up to date'
})

const updateStatusIcon = computed(() => {
  if (!versionInfo.value || !versionInfo.value.update_available)
    return '<polyline points="20 6 9 17 4 12"/>'
  return '<line x1="12" y1="5" x2="12" y2="19"/><polyline points="19 12 12 19 5 12"/>'
})

// ── Watchers ─────────────────────────────────────────────────────────────
watch(() => props.open, (v) => {
  if (v) {
    load()
    activeTab.value   = 'monitor'
    versionInfo.value = null
    updateDone.value  = false
    updateError.value = ''
  }
})

watch(activeTab, (v) => {
  if (v === 'about' && !versionInfo.value) checkUpdate()
})

function close() {
  saveStatus.value = ''
  emit('close')
}
</script>
