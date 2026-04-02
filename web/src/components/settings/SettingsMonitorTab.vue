<template>
  <div class="sm-monitor">
    <section class="sm-section" aria-labelledby="sm-targets-heading">
      <div class="sm-section-head">
        <h2 id="sm-targets-heading" class="sm-section-title">Probe targets</h2>
        <p class="sm-section-desc">Hosts to ping and resolve. The gateway is always probed.</p>
      </div>

      <div class="sm-target-grid">
        <div class="sm-target-card sm-target-card--ping">
          <div class="sm-target-card-top">
            <span class="sm-target-badge" aria-hidden="true">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
            </span>
            <div class="sm-target-card-head">
              <label class="sm-target-label" for="cfgPing">Ping</label>
              <span class="sm-count">{{ pingTargetCount }} hosts</span>
            </div>
          </div>
          <textarea
            id="cfgPing"
            class="sm-textarea"
            rows="4"
            placeholder="google.com&#10;cloudflare.com&#10;1.1.1.1"
            spellcheck="false"
            :value="pingTargets"
            @input="$emit('update:pingTargets', $event.target.value)"
          />
          <p class="sm-field-hint">One hostname or IP per line.</p>
        </div>

        <div class="sm-target-card sm-target-card--dns">
          <div class="sm-target-card-top">
            <span class="sm-target-badge sm-target-badge--dns" aria-hidden="true">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
            </span>
            <div class="sm-target-card-head">
              <label class="sm-target-label" for="cfgDns">DNS</label>
              <span class="sm-count">{{ dnsTargetCount }} hosts</span>
            </div>
          </div>
          <textarea
            id="cfgDns"
            class="sm-textarea"
            rows="4"
            placeholder="google.com&#10;cloudflare.com&#10;github.com"
            spellcheck="false"
            :value="dnsTargets"
            @input="$emit('update:dnsTargets', $event.target.value)"
          />
          <p class="sm-field-hint">Names are resolved to measure DNS latency.</p>
        </div>
      </div>
    </section>

    <section class="sm-section" aria-labelledby="sm-timing-heading">
      <div class="sm-section-head">
        <h2 id="sm-timing-heading" class="sm-section-title">Timing & sampling</h2>
        <p class="sm-section-desc">How often measurements run and how many samples to average.</p>
      </div>

      <div class="sm-timing-grid">
        <div class="sm-timing-card">
          <div class="sm-timing-card-text">
            <span class="sm-timing-name">Ping interval</span>
            <span class="sm-timing-sub">Between probe cycles</span>
          </div>
          <div class="sm-stp" role="group" :aria-label="'Ping interval, ' + pingInterval + ' seconds'">
            <button type="button" class="sm-stp-btn" @click="decPingInterval" :disabled="pingInterval <= 10" aria-label="Decrease ping interval">
              <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </button>
            <div class="sm-stp-val">
              <input
                class="sm-stp-num"
                type="number"
                min="10"
                max="3600"
                :value="pingInterval"
                @input="$emit('update:pingInterval', clampNum($event.target.value, 10, 3600))"
              />
              <span class="sm-stp-unit">sec</span>
            </div>
            <button type="button" class="sm-stp-btn" @click="incPingInterval" :disabled="pingInterval >= 3600" aria-label="Increase ping interval">
              <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </button>
          </div>
        </div>

        <div class="sm-timing-card">
          <div class="sm-timing-card-text">
            <span class="sm-timing-name">Speed test</span>
            <span class="sm-timing-sub">Bandwidth check cadence</span>
          </div>
          <div class="sm-stp" role="group" :aria-label="'Speed test every ' + speedInterval + ' minutes'">
            <button type="button" class="sm-stp-btn" @click="decSpeedInterval" :disabled="speedInterval <= 5" aria-label="Decrease speed test interval">
              <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </button>
            <div class="sm-stp-val">
              <input
                class="sm-stp-num"
                type="number"
                min="5"
                max="1440"
                :value="speedInterval"
                @input="$emit('update:speedInterval', clampNum($event.target.value, 5, 1440))"
              />
              <span class="sm-stp-unit">min</span>
            </div>
            <button type="button" class="sm-stp-btn" @click="incSpeedInterval" :disabled="speedInterval >= 1440" aria-label="Increase speed test interval">
              <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </button>
          </div>
        </div>

        <div class="sm-timing-card">
          <div class="sm-timing-card-text">
            <span class="sm-timing-name">Pings per cycle</span>
            <span class="sm-timing-sub">Averaged to smooth spikes</span>
          </div>
          <div class="sm-stp" role="group" :aria-label="pingCount + ' pings per cycle'">
            <button type="button" class="sm-stp-btn" @click="decPingCount" :disabled="pingCount <= 1" aria-label="Decrease pings per cycle">
              <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </button>
            <div class="sm-stp-val">
              <input
                class="sm-stp-num sm-stp-num--narrow"
                type="number"
                min="1"
                max="20"
                :value="pingCount"
                @input="$emit('update:pingCount', clampNum($event.target.value, 1, 20))"
              />
              <span class="sm-stp-unit">each</span>
            </div>
            <button type="button" class="sm-stp-btn" @click="incPingCount" :disabled="pingCount >= 20" aria-label="Increase pings per cycle">
              <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  pingTargets: { type: String, default: '' },
  dnsTargets: { type: String, default: '' },
  pingInterval: { type: Number, default: 60 },
  speedInterval: { type: Number, default: 30 },
  pingCount: { type: Number, default: 5 },
})

const emit = defineEmits([
  'update:pingTargets',
  'update:dnsTargets',
  'update:pingInterval',
  'update:speedInterval',
  'update:pingCount',
])

const pingTargetCount = computed(() =>
  props.pingTargets.split('\n').map(s => s.trim()).filter(Boolean).length
)
const dnsTargetCount = computed(() =>
  props.dnsTargets.split('\n').map(s => s.trim()).filter(Boolean).length
)

function clampNum(raw, min, max) {
  let n = parseInt(raw, 10)
  if (Number.isNaN(n)) n = min
  return Math.min(max, Math.max(min, n))
}

function decPingInterval() {
  emit('update:pingInterval', Math.max(10, props.pingInterval - 10))
}
function incPingInterval() {
  emit('update:pingInterval', Math.min(3600, props.pingInterval + 10))
}
function decSpeedInterval() {
  emit('update:speedInterval', Math.max(5, props.speedInterval - 5))
}
function incSpeedInterval() {
  emit('update:speedInterval', Math.min(1440, props.speedInterval + 5))
}
function decPingCount() {
  emit('update:pingCount', Math.max(1, props.pingCount - 1))
}
function incPingCount() {
  emit('update:pingCount', Math.min(20, props.pingCount + 1))
}
</script>

<style scoped>
.sm-monitor {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.sm-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.sm-section + .sm-section {
  margin-top: 1rem;
}

.sm-section-head {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.sm-section-title {
  margin: 0;
  font-size: 0.8125rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--fg);
}

.sm-section-desc {
  margin: 0;
  font-size: 0.75rem;
  line-height: 1.45;
  color: var(--muted);
  max-width: 42rem;
}

.sm-target-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.875rem;
}

@media (max-width: 560px) {
  .sm-target-grid { grid-template-columns: 1fr; }
  .sm-textarea {
    height: 4.25rem;
    min-height: 4.25rem;
    max-height: 4.25rem;
  }
}

.sm-target-card {
  border-radius: 12px;
  border: 1px solid var(--border);
  background: linear-gradient(165deg, var(--card) 0%, var(--bg) 100%);
  padding: 0.75rem 0.875rem 0.875rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.04) inset;
}

.sm-target-card--ping {
  border-color: color-mix(in srgb, var(--accent) 30%, var(--border));
  box-shadow: 0 8px 24px var(--shadow);
}

.sm-target-card--dns {
  border-color: color-mix(in srgb, var(--purple) 30%, var(--border));
  box-shadow: 0 8px 24px var(--shadow);
}

.sm-target-card-top {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
}

.sm-target-badge {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: var(--accent-soft);
  color: var(--accent);
  border: 1px solid color-mix(in srgb, var(--accent) 22%, transparent);
}

.sm-target-badge--dns {
  background: var(--purple-soft);
  color: var(--purple);
  border-color: color-mix(in srgb, var(--purple) 25%, transparent);
}

.sm-target-card-head {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
  flex: 1;
}

.sm-target-label {
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--fg);
  letter-spacing: -0.01em;
}

.sm-count {
  font-size: 0.625rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--muted);
}

.sm-textarea {
  box-sizing: border-box;
  width: 100%;
  height: 5rem;
  min-height: 5rem;
  max-height: 5rem;
  resize: none;
  overflow-y: auto;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--card);
  padding: 0.45rem 0.6rem;
  color: var(--fg);
  font-size: 0.8125rem;
  font-family: 'SF Mono', 'Fira Code', ui-monospace, monospace;
  line-height: 1.55;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.sm-textarea:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-soft);
}

.sm-target-card--dns .sm-textarea:focus {
  border-color: var(--purple);
  box-shadow: 0 0 0 3px var(--purple-soft);
}

.sm-field-hint {
  margin: 0;
  font-size: 0.6875rem;
  color: var(--muted);
  line-height: 1.4;
}

.sm-timing-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
}

@media (max-width: 520px) {
  .sm-timing-grid { grid-template-columns: 1fr; }
}

.sm-timing-card {
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--card);
  padding: 0.65rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
}

.sm-timing-card-text {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.sm-timing-name {
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--fg);
  letter-spacing: -0.01em;
}

.sm-timing-sub {
  font-size: 0.6875rem;
  color: var(--muted);
  line-height: 1.35;
}

.sm-stp {
  display: flex;
  align-items: center;
  gap: 2px;
  align-self: flex-start;
  width: 100%;
  max-width: 100%;
  justify-content: space-between;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 3px;
}

.sm-stp-btn {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: var(--muted);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: color 0.12s, background 0.12s;
  flex-shrink: 0;
}

.sm-stp-btn:hover:not(:disabled) {
  color: var(--fg);
  background: var(--input-hover-bg);
}

.sm-stp-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.sm-stp-val {
  display: flex;
  align-items: baseline;
  gap: 3px;
  min-width: 0;
  flex: 1;
  justify-content: center;
  padding: 0 4px;
}

.sm-stp-num {
  background: transparent;
  border: none;
  outline: none;
  font-size: 1.125rem;
  font-weight: 800;
  color: var(--fg);
  width: 2.5rem;
  text-align: center;
  font-family: 'Inter', sans-serif;
  letter-spacing: -0.03em;
  -moz-appearance: textfield;
}

.sm-stp-num--narrow {
  width: 2rem;
}

.sm-stp-num::-webkit-outer-spin-button,
.sm-stp-num::-webkit-inner-spin-button {
  -webkit-appearance: none;
}

.sm-stp-unit {
  font-size: 0.5625rem;
  font-weight: 700;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  white-space: nowrap;
}
</style>
