<template>
  <AccordionCard section="conn" padding="p-3">
    <template #title>
      <span class="flex items-center gap-1.5">
        <!-- WiFi icon -->
        <span v-if="conn && type === 'wifi'">
          <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><circle cx="12" cy="20" r="1" fill="currentColor" stroke="none"/></svg>
        </span>
        <!-- Ethernet icon -->
        <span v-else>
          <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect x="16" y="16" width="6" height="6" rx="1"/><rect x="2" y="16" width="6" height="6" rx="1"/><rect x="9" y="2" width="6" height="6" rx="1"/><path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"/><path d="M12 12V8"/></svg>
        </span>
        {{ !conn ? 'Connection' : type === 'wifi' ? 'WiFi Signal' : 'Ethernet' }}
      </span>
    </template>

    <!-- No data -->
    <div v-if="!conn" class="flex flex-col items-center gap-1.5 py-4" style="color:var(--muted)">
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" style="opacity:0.4"><rect x="16" y="16" width="6" height="6" rx="1"/><rect x="2" y="16" width="6" height="6" rx="1"/><rect x="9" y="2" width="6" height="6" rx="1"/><path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"/><path d="M12 12V8"/></svg>
      <span style="font-size:0.6875rem;font-weight:600;opacity:0.6">No connection data</span>
    </div>

    <!-- WiFi details -->
    <div v-else-if="type === 'wifi'" class="flex flex-col gap-1.5 text-xs">
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Signal (RSSI)</span>
        <span class="font-bold" :style="{ color: signalColor }">{{ rssi }} dBm</span>
      </div>
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">SNR</span>
        <span class="font-bold">{{ snr ? snr + ' dB' : '—' }}</span>
      </div>
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Band · Channel</span>
        <span class="font-bold">{{ bandChannel }}</span>
      </div>
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Link Rate</span>
        <span class="font-bold">{{ rate ? rate + ' Mbps' : '—' }}</span>
      </div>
      <div style="height:1px;background:var(--border);margin:0.25rem 0" />
      <div class="flex items-center gap-2">
        <div style="flex:1;height:4px;background:var(--border);border-radius:2px;overflow:hidden">
          <div :style="{ width: signalPct + '%', background: signalColor, height: '100%', borderRadius: '2px', transition: 'width 0.5s ease' }" />
        </div>
        <span :style="{ fontSize: '10px', minWidth: '52px', textAlign: 'right', color: signalColor }">{{ signalQuality }}</span>
      </div>
    </div>

    <!-- Ethernet details -->
    <div v-else-if="type === 'ethernet'" class="flex flex-col gap-1.5 text-xs">
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Link Speed</span>
        <span class="font-bold" :style="{ color: ethSpeedColor }">{{ rate ? ethSpeed : '—' }}</span>
      </div>
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Duplex</span>
        <span class="font-bold">{{ duplexLabel }}</span>
      </div>
    </div>
  </AccordionCard>
</template>

<script setup>
import { computed } from 'vue'
import AccordionCard from './AccordionCard.vue'

const props = defineProps({ history: Array })

const conn = computed(() => {
  if (!props.history?.length) return null
  const latest = props.history[props.history.length - 1]
  return latest.conn_type ? latest : null
})

const type   = computed(() => conn.value?.conn_type ?? '')
const rssi   = computed(() => conn.value?.conn_rssi ?? 0)
const snr    = computed(() => conn.value?.conn_snr ?? 0)
const band   = computed(() => conn.value?.conn_band ?? '')
const ch     = computed(() => conn.value?.conn_channel ?? 0)
const rate   = computed(() => conn.value?.conn_link_rate ?? 0)
const duplex = computed(() => conn.value?.conn_duplex ?? '')

const signalPct  = computed(() => Math.max(0, Math.min(100, ((rssi.value + 90) / 60) * 100)))
const signalColor = computed(() => {
  if (rssi.value >= -50) return 'var(--green)'
  if (rssi.value >= -65) return 'var(--green)'
  if (rssi.value >= -75) return 'var(--yellow)'
  return 'var(--red)'
})
const signalQuality = computed(() => {
  if (rssi.value >= -50) return 'Excellent'
  if (rssi.value >= -65) return 'Good'
  if (rssi.value >= -75) return 'Fair'
  return 'Poor'
})

const ethSpeed = computed(() => rate.value >= 1000 ? (rate.value / 1000) + ' Gbps' : rate.value + ' Mbps')
const ethSpeedColor = computed(() => rate.value >= 1000 ? 'var(--green)' : 'var(--fg)')
const duplexLabel = computed(() => duplex.value ? duplex.value.charAt(0).toUpperCase() + duplex.value.slice(1) + ' duplex' : '—')
const bandChannel = computed(() => [band.value, ch.value ? 'ch. ' + ch.value : ''].filter(Boolean).join(' · ') || '—')
</script>
