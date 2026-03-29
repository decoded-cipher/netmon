<template>
  <div class="card" :class="[padding, { 's-collapsed': collapsed }]">
    <div class="section-label accordion-toggle" @click="toggle">
      <slot name="title" />
      <svg class="accordion-chevron" xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
    </div>
    <div v-if="!collapsed" class="accordion-body-inner">
      <slot />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const props = defineProps({
  section: { type: String, required: true },
  padding: { type: String, default: 'p-3' },
})

const collapsed = ref(false)

onMounted(() => {
  collapsed.value = localStorage.getItem('acc-' + props.section) === '1'
})

function toggle() {
  collapsed.value = !collapsed.value
  localStorage.setItem('acc-' + props.section, collapsed.value ? '1' : '0')
}
</script>
