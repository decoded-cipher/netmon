<template>
  <div class="modal-backdrop" :class="{ open }" @click.self="close">
    <div class="modal" role="dialog" aria-modal="true">

      <div class="modal-header">
        <h2>
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="color:var(--accent)"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          Monitor Settings
        </h2>
        <button class="modal-close" @click="close">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <div class="modal-body">
        <!-- Targets -->
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

        <!-- Intervals -->
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

      <div class="modal-footer">
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
import { ref } from 'vue'

const props = defineProps({ open: Boolean })
const emit  = defineEmits(['close', 'saved'])

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

  if (!pt.length)         return showStatus('Probe hosts cannot be empty', false)
  if (pis < 10)           return showStatus('Ping interval must be ≥ 10 s', false)
  if (sim < 5)            return showStatus('Speed interval must be ≥ 5 min', false)
  if (pc < 1 || pc > 20)  return showStatus('Ping count must be 1–20', false)

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

function close() {
  saveStatus.value = ''
  emit('close')
}

// Load config when opened
import { watch } from 'vue'
watch(() => props.open, (v) => { if (v) load() })
</script>
