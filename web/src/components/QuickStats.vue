<template>
  <AccordionCard section="summary" padding="p-3">
    <template #title>
      <span class="flex items-center gap-1.5">
        <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
        24h Summary
      </span>
    </template>
    <div v-if="!summary" class="card-empty">
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
      <span>No data yet</span>
    </div>
    <div v-else class="flex flex-col gap-1.5">
      <div class="card-row">
        <span class="card-row-label">Uptime</span>
        <span class="card-row-value">{{ summary.uptime_24h }}%</span>
      </div>
      <div class="card-row">
        <span class="card-row-label">Outages</span>
        <span class="card-row-value">{{ summary.outages_24h }}</span>
      </div>
      <div class="card-divider" />
      <div class="card-row">
        <span class="card-row-label">Peak Download</span>
        <span class="card-row-value">{{ peakDl }}</span>
      </div>
      <div class="card-row">
        <span class="card-row-label">Min Latency</span>
        <span class="card-row-value">{{ fmt(summary?.latency_min, ' ms') }}</span>
      </div>
      <div class="card-row">
        <span class="card-row-label">Max Latency</span>
        <span class="card-row-value">{{ fmt(summary?.latency_max, ' ms') }}</span>
      </div>
      <div class="card-row">
        <span class="card-row-label">Avg Jitter</span>
        <span class="card-row-value">{{ fmt(summary.jitter_avg, ' ms') }}</span>
      </div>
    </div>
  </AccordionCard>
</template>

<script setup>
import { computed } from 'vue'
import AccordionCard from './AccordionCard.vue'
import { fmt } from '../utils/format.js'

const props = defineProps({ summary: Object, history: Array })

const peakDl = computed(() => {
  const vals = props.history?.map(h => h.download).filter(v => v != null) ?? []
  return vals.length ? Math.max(...vals) + ' Mbps' : '—'
})
</script>
