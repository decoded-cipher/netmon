<template>
  <div class="duration-wrapper" @focusout="closeOnOutside">
    <button class="duration-btn" :class="{ open }" @click="open = !open" type="button">
      <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
      <span class="duration-label">{{ label }}</span>
      <svg class="dur-chevron" xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
    </button>

    <div class="duration-panel" :class="{ open }">
      <div class="dur-presets">
        <button
          v-for="p in presets" :key="p.minutes"
          class="dur-chip" :class="{ active: modelValue === p.minutes }"
          @click="select(p.minutes, p.label)"
          type="button"
        >{{ p.label }}</button>
      </div>
      <div class="dur-divider">Custom</div>
      <div class="dur-custom">
        <span class="dur-custom-label">Last</span>
        <input class="dur-num" type="number" v-model="durNum" min="1" max="9999" @keydown.enter="applyCustom" />
        <select class="dur-unit" v-model="durUnit">
          <option :value="1">min</option>
          <option :value="60">hr</option>
          <option :value="1440">day</option>
        </select>
      </div>
      <button class="dur-apply" @click="applyCustom" type="button">Apply</button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({ modelValue: { type: Number, default: 60 } })
const emit  = defineEmits(['update:modelValue', 'change'])

const open     = ref(false)
const label    = ref('1h')
const durNum   = ref(1)
const durUnit  = ref(60)

const presets  = [
  { minutes: 15,   label: '15m' },
  { minutes: 60,   label: '1h' },
  { minutes: 360,  label: '6h' },
  { minutes: 1440, label: '24h' },
]

function select(minutes, l) {
  label.value = l
  open.value  = false
  emit('update:modelValue', minutes)
  emit('change', minutes)
}

function applyCustom() {
  const n    = Math.max(1, parseInt(durNum.value) || 1)
  const unit = parseInt(durUnit.value)
  const mins = n * unit
  const ul   = unit === 1 ? 'm' : unit === 60 ? 'h' : 'd'
  select(mins, n + ul)
}

function closeOnOutside(e) {
  if (!e.currentTarget.contains(e.relatedTarget)) open.value = false
}
</script>
