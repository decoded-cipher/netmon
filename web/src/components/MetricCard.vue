<template>
  <div class="card metric-card flash-trigger" :data-color="color">
    <div class="flex items-center justify-between">
      <div class="metric-label">{{ label }}</div>
      <div class="metric-icon" :class="color" v-html="icon" />
    </div>
    <div class="metric-value">{{ value }}</div>
    <div class="metric-sub">{{ sub }}</div>
    <span v-if="spark" v-html="spark" />
  </div>
</template>

<script setup>
defineProps({
  label:  String,
  value:  String,
  sub:    String,
  color:  String,
  icon:   String,
  spark:  String,
})
</script>
