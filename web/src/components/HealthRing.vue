<template>
  <div class="flex flex-col items-center">
    <div v-if="!summary" class="card-empty">
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
      <span>No data yet</span>
    </div>

    <template v-else>
      <div class="health-ring">
        <svg viewBox="0 0 100 100" width="100" height="100">
          <circle class="ring-bg" cx="50" cy="50" r="42" />
          <circle class="ring-fg" cx="50" cy="50" r="42"
            stroke-dasharray="263.89"
            :stroke-dashoffset="dashOffset"
            :style="{ stroke: color }"
          />
        </svg>
        <div class="health-score-text">
          <span class="health-score-num" :style="{ color }">{{ score }}</span>
          <span class="health-score-label">Score</span>
        </div>
      </div>

      <div class="grid grid-cols-3 gap-2 w-full mt-3">
        <div class="flex flex-col items-center gap-0.5">
          <span class="health-metric-value" :style="{ color: lat?.color }">{{ lat?.text ?? '—' }}</span>
          <span class="health-metric-label">Latency</span>
        </div>
        <div class="flex flex-col items-center gap-0.5">
          <span class="health-metric-value" :style="{ color: loss?.color }">{{ loss?.text ?? '—' }}</span>
          <span class="health-metric-label">Loss</span>
        </div>
        <div class="flex flex-col items-center gap-0.5">
          <span class="health-metric-value" :style="{ color: spd?.color }">{{ spd?.text ?? '—' }}</span>
          <span class="health-metric-label">Speed</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { computeHealth, rateMetric } from '../utils/health.js'

const props = defineProps({ summary: Object })

const CIRC = 2 * Math.PI * 42

const score = computed(() => props.summary ? computeHealth(props.summary) : 0)
const color = computed(() => {
  if (!props.summary) return 'var(--border)'
  return score.value >= 80 ? 'var(--green)' : score.value >= 50 ? 'var(--yellow)' : 'var(--red)'
})
const dashOffset = computed(() => CIRC - (score.value / 100) * CIRC)

const lat  = computed(() => props.summary ? rateMetric(props.summary.latency_avg, 20, 50) : null)
const loss = computed(() => props.summary ? rateMetric(props.summary.packet_loss, 0.5, 2) : null)
const spd  = computed(() => {
  if (!props.summary) return null
  const d = props.summary.download_avg
  return d >= 50 ? { text: 'Good', color: 'var(--green)' }
       : d >= 25 ? { text: 'Fair', color: 'var(--yellow)' }
       :           { text: 'Poor', color: 'var(--red)' }
})
</script>
