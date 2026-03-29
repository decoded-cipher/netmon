<template>
  <div class="flex flex-col h-screen px-4 sm:px-6 lg:px-8 py-4 max-w-[1680px] mx-auto overflow-hidden">

    <AppHeader
      :status="statusBadge"
      :target-count="targetCount"
      :network-label="networkLabel"
      :ago-label="agoLabel"
      :is-refreshing="isLoading"
      :is-dark="isDark"
      :current-minutes="currentMinutes"
      @toggle-theme="$emit('toggle-theme')"
      @open-settings="settingsOpen = true"
    >
      <template #duration-picker>
        <DurationPicker v-model="currentMinutes" @change="onDurationChange" />
      </template>
    </AppHeader>

    <SettingsModal :open="settingsOpen" @close="settingsOpen = false" @saved="loadData()" />

    <!-- Scroll wrapper -->
    <div class="flex-1 min-h-0 overflow-y-auto lg:overflow-hidden">
      <div class="grid grid-cols-1 lg:grid-cols-[1fr_300px] gap-3 lg:h-full">

        <!-- Left column -->
        <div class="flex flex-col gap-3 lg:min-h-0">

          <!-- Metric cards -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 flex-shrink-0">
            <MetricCard v-for="(card, i) in metricCards" :key="i" v-bind="card" />
          </div>

          <!-- Charts 2×2 — :key forces full remount on theme/resize -->
          <div id="chartGrid" class="grid grid-cols-1 md:grid-cols-2 gap-3 lg:flex-1 lg:min-h-0">
            <ChartPanel :key="'lat-' + chartKey" :config="chartConfigs.latency" :label="chartLabels.latency" title="Latency"               dot="var(--accent)"  label-id="latencyChart" :icon="CHART_ICONS.latency" />
            <ChartPanel :key="'spd-' + chartKey" :config="chartConfigs.speed"   :label="chartLabels.speed"   title="Throughput"            dot="var(--green)"   label-id="speedChart"   :icon="CHART_ICONS.speed"   />
            <ChartPanel :key="'los-' + chartKey" :config="chartConfigs.loss"    :label="chartLabels.loss"    title="Packet Loss &amp; Jitter" dot="var(--red)"  label-id="lossChart"    :icon="CHART_ICONS.loss"    />
            <ChartPanel :key="'dns-' + chartKey" :config="chartConfigs.dns"     :label="chartLabels.dns"     title="DNS Resolution"        dot="var(--purple)"  label-id="dnsChart"     :icon="CHART_ICONS.dns"    />
          </div>
        </div>

        <!-- Right sidebar -->
        <div class="sidebar-scroll flex flex-col gap-3 lg:min-h-0 lg:overflow-y-auto pb-4 lg:pb-0">

          <AccordionCard section="health" padding="p-3">
            <template #title>
              <span class="flex items-center gap-1.5">
                <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
                Connection Health
              </span>
            </template>
            <HealthRing :summary="summary" />
          </AccordionCard>

          <AccordionCard section="connectivity" padding="p-3">
            <template #title>
              <span class="flex items-center gap-1.5">
                <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>
                Connectivity
              </span>
            </template>
            <ConnectivityTable :targets="targets" />
          </AccordionCard>

          <AccordionCard section="dns" padding="p-3">
            <template #title>
              <span class="flex items-center gap-1.5">
                <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                DNS Lookups
              </span>
            </template>
            <DnsTable :checks="dns" />
          </AccordionCard>

          <ConnectionCard :history="history" />
          <QuickStats :summary="summary" :history="history" />

        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import AppHeader         from './AppHeader.vue'
import DurationPicker    from './DurationPicker.vue'
import SettingsModal     from './SettingsModal.vue'
import MetricCard        from './MetricCard.vue'
import ChartPanel        from './ChartPanel.vue'
import AccordionCard     from './AccordionCard.vue'
import HealthRing        from './HealthRing.vue'
import ConnectivityTable from './ConnectivityTable.vue'
import DnsTable          from './DnsTable.vue'
import ConnectionCard    from './ConnectionCard.vue'
import QuickStats        from './QuickStats.vue'
import { useToasts }     from '../composables/useToasts.js'
import { sparklineSVG }  from '../utils/sparkline.js'

const props = defineProps({ isDark: Boolean })
defineEmits(['toggle-theme'])

const REFRESH_MS = 15000

const { show: showToast } = useToasts()

// ── State ───────────────────────────────────────────────────────────────
const apiData        = ref(null)
const apiConfig      = ref(null)
const currentMinutes = ref(60)
const isLoading      = ref(false)
const lastUpdateTime = ref(null)
const prevSummary    = ref(null)
const prevStatus     = ref(null)
const chartKey       = ref(0)
const settingsOpen   = ref(false)
const agoLabel       = ref('—')

// Recreate charts when theme changes
watch(() => props.isDark, () => { chartKey.value++ })

// ── "X ago" ticker ──────────────────────────────────────────────────────
const agoTimer = setInterval(() => {
  if (!lastUpdateTime.value) return
  const sec = Math.floor((Date.now() - lastUpdateTime.value) / 1000)
  if      (sec < 5)    agoLabel.value = 'just now'
  else if (sec < 60)   agoLabel.value = sec + 's ago'
  else if (sec < 3600) agoLabel.value = Math.floor(sec / 60) + 'm ago'
  else                 agoLabel.value = Math.floor(sec / 3600) + 'h ago'
}, 1000)

// ── Data loading ─────────────────────────────────────────────────────────
async function loadData(forceRecreate = false) {
  if (isLoading.value) return
  isLoading.value = true
  try {
    const [d, cfg] = await Promise.all([
      fetch(`/api/data?minutes=${currentMinutes.value}`).then(r => r.json()),
      fetch('/api/config').then(r => r.json()),
    ])
    lastUpdateTime.value = Date.now()
    checkThresholds(d.summary, prevSummary.value)

    if (cfg) {
      const allowedPing = new Set(cfg.ping_targets.map(h => h.toLowerCase()))
      d.targets = (d.targets || []).filter(t => allowedPing.has(t.host.toLowerCase()))
      const allowedDns = new Set(cfg.dns_targets.map(h => h.toLowerCase()))
      d.dns = (d.dns || []).filter(x => allowedDns.has(x.host.toLowerCase()))
    }

    prevSummary.value = d.summary
    apiData.value     = d
    apiConfig.value   = cfg
    if (forceRecreate) chartKey.value++
  } catch (_) { /* keep stale data on error */ }
  isLoading.value = false
}

function checkThresholds(s, prev) {
  const curr = s.packet_loss > 5 ? 'offline' : s.packet_loss > 1 ? 'degraded' : 'online'
  if (prevStatus.value !== null && prevStatus.value !== curr) {
    if      (curr === 'offline')  showToast('st-offline',  'Connection Offline',  `Packet loss at ${s.packet_loss}%`, 'error', 8000)
    else if (curr === 'degraded') showToast('st-degraded', 'Connection Degraded', `Packet loss at ${s.packet_loss}%`, 'warn',  7000)
    else                          showToast('st-ok',       'Connection Restored', 'All metrics back to normal',       'ok',    5000)
  }
  prevStatus.value = curr
  if (prev && s.latency_avg > 150 && prev.latency_avg <= 150)
    showToast('lat-spike', 'Latency Spike', `${s.latency_avg}ms avg (was ${prev.latency_avg}ms)`, 'warn', 6000)
}

// ── Polling & resize ──────────────────────────────────────────────────────
let pollTimer = null, resizeTimer = null

onMounted(() => {
  loadData()
  pollTimer = setInterval(loadData, REFRESH_MS)
  window.addEventListener('resize', onResize)
})
onUnmounted(() => {
  clearInterval(pollTimer)
  clearInterval(agoTimer)
  window.removeEventListener('resize', onResize)
})

function onResize() {
  clearTimeout(resizeTimer)
  resizeTimer = setTimeout(() => loadData(true), 250)
}

function onDurationChange(minutes) {
  currentMinutes.value = minutes
  loadData(true)
}

// ── Computed: header ──────────────────────────────────────────────────────
const summary     = computed(() => apiData.value?.summary)
const history     = computed(() => apiData.value?.history || [])
const targets     = computed(() => apiData.value?.targets || [])
const dns         = computed(() => apiData.value?.dns || [])
const networkId   = computed(() => apiData.value?.network_id || 'unknown')
const targetCount = computed(() => apiConfig.value ? apiConfig.value.ping_targets.length : targets.value.length)
const networkLabel = computed(() => {
  const interval = apiConfig.value ? ` · ping ${apiConfig.value.ping_interval_s}s` : ''
  return networkId.value + interval
})
const statusBadge = computed(() => {
  const loss = summary.value?.packet_loss ?? 0
  return loss > 5 ? 'offline' : loss > 1 ? 'degraded' : 'online'
})

// ── Computed: metric cards ────────────────────────────────────────────────
const SVG = (d) => `<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${d}</svg>`
const CHART_ICONS = {
  latency: SVG('<circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>'),
  speed:   SVG('<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>'),
  loss:    SVG('<polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>'),
  dns:     SVG('<rect x="2" y="2" width="20" height="8" rx="2" ry="2"/><rect x="2" y="14" width="20" height="8" rx="2" ry="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/>'),
}

const ICONS = {
  clock:     `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>`,
  arrowDown: `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><polyline points="19 12 12 19 5 12"/></svg>`,
  arrowUp:   `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/></svg>`,
  activity:  `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>`,
}

const metricCards = computed(() => {
  if (!summary.value) return [
    { label: 'Avg Latency', value: '—', icon: ICONS.clock,     color: 'blue',   sub: 'waiting for data', spark: null },
    { label: 'Download',    value: '—', icon: ICONS.arrowDown, color: 'green',  sub: 'waiting for data', spark: null },
    { label: 'Upload',      value: '—', icon: ICONS.arrowUp,   color: 'purple', sub: 'waiting for data', spark: null },
    { label: 'Packet Loss', value: '—', icon: ICONS.activity,  color: 'yellow', sub: 'waiting for data', spark: null },
  ]
  const s  = summary.value
  const l20 = history.value.slice(-20)
  const lc  = s.packet_loss > 1 ? 'red' : 'yellow'
  const C   = { blue: '#3b82f6', green: '#10b981', purple: '#8b5cf6', red: '#ef4444', yellow: '#f59e0b' }
  return [
    { label: 'Avg Latency', value: s.latency_avg  + ' ms',   icon: ICONS.clock,     color: 'blue',   sub: 'p95 ' + s.latency_p95 + ' ms',            spark: sparklineSVG(l20.map(h => h.latency),  C.blue)   },
    { label: 'Download',    value: s.download_avg + ' Mbps', icon: ICONS.arrowDown, color: 'green',  sub: 'avg throughput',                           spark: sparklineSVG(l20.map(h => h.download), C.green)  },
    { label: 'Upload',      value: s.upload_avg   + ' Mbps', icon: ICONS.arrowUp,   color: 'purple', sub: 'avg throughput',                           spark: sparklineSVG(l20.map(h => h.upload),   C.purple) },
    { label: 'Packet Loss', value: s.packet_loss  + '%',     icon: ICONS.activity,  color: lc,       sub: s.packet_loss < 1 ? 'healthy' : 'elevated', spark: sparklineSVG(l20.map(h => h.loss),     C[lc])    },
  ]
})

// ── Computed: charts ──────────────────────────────────────────────────────
function getChartHeight(hasData) {
  if (window.innerWidth >= 1024) {
    if (!hasData) return '100%'
    const grid = document.getElementById('chartGrid')
    if (grid && grid.offsetHeight > 0)
      return Math.max(80, Math.floor((grid.offsetHeight - 12) / 2) - 52)
    return '100%'
  }
  return window.innerWidth < 640 ? 160 : 185
}

const chartConfigs = computed(() => {
  const h      = history.value
  const times  = h.map(x => x.time)
  const isSmall = window.innerWidth < 640
  const chartH  = getChartHeight(h.length > 0)
  const tick    = Math.min(times.length, isSmall ? 4 : 6)
  const ls      = { colors: 'var(--muted)', fontSize: isSmall ? '9px' : '10px' }
  const dark    = props.isDark

  const noData = {
    text: 'Waiting for data…',
    align: 'center', verticalAlign: 'middle',
    style: { color: dark ? '#7b8ba4' : '#9ca3af', fontSize: '12px', fontFamily: "'Inter',sans-serif" },
  }

  const base = {
    chart: { type: 'area', background: 'transparent', fontFamily: "'Inter',sans-serif", toolbar: { show: false }, height: chartH, animations: { speed: 500 }, zoom: { enabled: false } },
    stroke: { curve: 'smooth', width: 2 },
    markers: { size: times.length > 0 && times.length <= 5 ? 4 : 0, hover: { size: 5 } },
    grid: { borderColor: 'var(--border)', strokeDashArray: 4, xaxis: { lines: { show: false } }, yaxis: { lines: { show: true } }, padding: { top: -8, right: 0, bottom: 0, left: 13.5 } },
    xaxis: { categories: times, tickAmount: tick, labels: { style: ls, rotate: 0, hideOverlappingLabels: true, maxHeight: 30 }, axisBorder: { show: true, color: 'var(--border)' }, axisTicks: { show: false } },
    yaxis: { labels: { style: ls }, forceNiceScale: true, axisBorder: { show: true, color: 'var(--border)' }, axisTicks: { show: false } },
    legend: { position: 'top', fontSize: '10px', labels: { colors: 'var(--subtle)' }, markers: { width: 6, height: 6, radius: 2 }, itemMargin: { horizontal: 8 } },
    tooltip: { theme: dark ? 'dark' : 'light', style: { fontSize: '11px' }, x: { show: true } },
    dataLabels: { enabled: false },
    noData,
  }
  const fill = (o = 0.12) => ({ type: 'gradient', gradient: { shadeIntensity: 0, opacityFrom: o, opacityTo: 0.01, stops: [0, 95] } })

  return {
    latency: { ...base, series: [{ name: 'Latency (ms)', data: h.map(x => x.latency) }], colors: ['#3b82f6'], fill: fill(), yaxis: { ...base.yaxis, min: 0 } },
    speed:   { ...base, series: [{ name: 'Download', data: h.map(x => x.download || null) }, { name: 'Upload', data: h.map(x => x.upload || null) }], colors: ['#10b981', '#8b5cf6'], fill: fill(0.08) },
    loss: {
      ...base,
      series: [{ name: 'Loss (%)', data: h.map(x => x.loss) }, { name: 'Jitter (ms)', data: h.map(x => x.jitter) }],
      colors: ['#ef4444', '#f59e0b'], fill: fill(0.1),
      yaxis: [
        { forceNiceScale: true, min: 0, axisBorder: { show: true, color: 'var(--border)' }, axisTicks: { show: false }, title: { text: 'Loss %',   style: { fontSize: '10px', color: 'var(--muted)' } }, labels: { style: { colors: 'var(--muted)', fontSize: '10px' } } },
        { opposite: true, forceNiceScale: true, min: 0, axisBorder: { show: false }, axisTicks: { show: false }, title: { text: 'Jitter ms', style: { fontSize: '10px', color: 'var(--muted)' } }, labels: { style: { colors: 'var(--muted)', fontSize: '10px' } } },
      ],
    },
    dns: { ...base, series: [{ name: 'DNS (ms)', data: h.map(x => x.dns || null) }], colors: ['#8b5cf6'], fill: fill(), yaxis: { ...base.yaxis, min: 0 } },
  }
})

const chartLabels = computed(() => {
  if (!summary.value) return { latency: '—', speed: '—', loss: '—', dns: '—' }
  const s = summary.value
  return {
    latency: s.latency_avg + ' ms',
    speed:   '↓ ' + s.download_avg + '  ↑ ' + s.upload_avg + ' Mbps',
    loss:    s.packet_loss + '% · ' + s.jitter_avg + ' ms',
    dns:     s.dns_avg + ' ms',
  }
})
</script>
