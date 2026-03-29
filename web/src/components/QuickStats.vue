<template>
  <AccordionCard section="summary" padding="p-3">
    <template #title>
      <span class="flex items-center gap-1.5">
        <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
        24h Summary
      </span>
    </template>
    <div v-if="!summary" class="flex flex-col items-center gap-1.5 py-4" style="color:var(--muted)">
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" style="opacity:0.4"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
      <span style="font-size:0.6875rem;font-weight:600;opacity:0.6">No data yet</span>
    </div>
    <div v-else class="flex flex-col gap-1.5 text-xs">
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Uptime</span>
        <span class="font-bold">{{ summary.uptime_24h }}%</span>
      </div>
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Outages</span>
        <span class="font-bold">{{ summary.outages_24h }}</span>
      </div>
      <div style="height:1px;background:var(--border);margin:0.25rem 0" />
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Peak Download</span>
        <span class="font-bold">{{ peakDl }}</span>
      </div>
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Min Latency</span>
        <span class="font-bold">{{ minLat }} ms</span>
      </div>
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Max Latency</span>
        <span class="font-bold">{{ maxLat }} ms</span>
      </div>
      <div class="flex justify-between items-center">
        <span style="color:var(--muted)">Avg Jitter</span>
        <span class="font-bold">{{ summary.jitter_avg }} ms</span>
      </div>
    </div>
  </AccordionCard>
</template>

<script setup>
import { computed } from 'vue'
import AccordionCard from './AccordionCard.vue'

const props = defineProps({ summary: Object, history: Array })

const peakDl = computed(() => {
  const vals = props.history?.map(h => h.download).filter(v => v > 0) ?? []
  return vals.length ? Math.max(...vals) + ' Mbps' : '—'
})
const minLat = computed(() => props.summary?.latency_min ?? (props.history?.length ? Math.min(...props.history.map(h => h.latency)) : '—'))
const maxLat = computed(() => props.summary?.latency_max ?? (props.history?.length ? Math.max(...props.history.map(h => h.latency)) : '—'))
</script>
