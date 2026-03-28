<template>
  <div class="flex flex-col items-center">
    <div class="health-ring">
      <svg viewBox="0 0 100 100" width="100" height="100">
        <circle class="ring-bg" cx="50" cy="50" r="42" />
        <circle
          class="ring-fg"
          cx="50" cy="50" r="42"
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
    <div class="grid grid-cols-3 gap-3 w-full mt-4 text-center">
      <div>
        <div class="text-xs font-bold" :style="{ color: lat?.color }">{{ lat?.text ?? '—' }}</div>
        <div class="text-[10px] font-medium" style="color:var(--muted)">Latency</div>
      </div>
      <div>
        <div class="text-xs font-bold" :style="{ color: loss?.color }">{{ loss?.text ?? '—' }}</div>
        <div class="text-[10px] font-medium" style="color:var(--muted)">Loss</div>
      </div>
      <div>
        <div class="text-xs font-bold" :style="{ color: spd?.color }">{{ spd?.text ?? '—' }}</div>
        <div class="text-[10px] font-medium" style="color:var(--muted)">Speed</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { computeHealth, rateMetric } from '../utils/health.js'

const props = defineProps({ summary: Object })

const CIRC = 2 * Math.PI * 42

const score = computed(() => props.summary ? computeHealth(props.summary) : 0)
const color = computed(() => score.value >= 80 ? 'var(--green)' : score.value >= 50 ? 'var(--yellow)' : 'var(--red)')
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
