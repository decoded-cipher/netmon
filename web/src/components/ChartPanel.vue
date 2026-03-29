<template>
  <div class="card chart-box">
    <div class="chart-header">
      <h3>
        <span class="chart-dot" :style="{ background: dot }" />
        {{ title }}
      </h3>
      <div v-if="label || icon" class="chart-badge" :style="{ color: dot }">
        <span v-if="icon" class="chart-badge-icon" v-html="icon" />
        <span>{{ label }}</span>
      </div>
    </div>
    <div :id="labelId" ref="chartEl" style="height:100%" />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import ApexCharts from 'apexcharts'

const props = defineProps({
  config: { type: Object, required: true },
  label:  { type: String, default: '' },
  title:  { type: String, required: true },
  dot:    { type: String, required: true },
  labelId:{ type: String, required: true },
  icon:   { type: String, default: '' },
})

const chartEl = ref(null)
let chart = null

function build() {
  if (chart) { chart.destroy(); chart = null }
  if (!chartEl.value) return
  chart = new ApexCharts(chartEl.value, props.config)
  chart.render()
}

onMounted(build)
onUnmounted(() => { if (chart) chart.destroy() })

// Smooth update when only data changes; full rebuild handled by parent via :key
watch(() => props.config, (cfg) => {
  if (!chart) { build(); return }
  chart.updateOptions(cfg, false, false)
}, { deep: true })
</script>
