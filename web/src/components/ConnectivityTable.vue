<template>
  <table class="target-table">
    <thead>
      <tr><th>Host</th><th>Latency</th><th>Loss</th></tr>
    </thead>
    <tbody>
      <tr v-for="t in targets" :key="t.host">
        <td>
          <div class="host-cell">
            <span class="inline-dot" :class="t.status === 'up' ? 'up' : 'down'" />
            <span>{{ t.host }}</span>
          </div>
          <div class="ip-cell">{{ t.ip }}</div>
        </td>
        <td class="font-semibold">{{ t.latency }} ms</td>
        <td>{{ t.loss }}%</td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
defineProps({ targets: Array })
</script>
