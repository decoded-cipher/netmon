<template>
  <table class="target-table">
    <thead v-if="targets?.length">
      <tr><th>Host</th><th>Latency</th><th>Loss</th></tr>
    </thead>
    <tbody>
      <tr v-if="!targets?.length">
        <td colspan="3" style="border-bottom:none">
          <div class="card-empty">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>
            <span>No targets yet</span>
          </div>
        </td>
      </tr>
      <tr v-else v-for="t in targets" :key="t.host">
        <td>
          <div class="host-cell">
            <span class="inline-dot" :class="t.status === 'up' ? 'up' : 'down'" />
            <span>{{ t.host }}</span>
          </div>
          <div class="ip-cell">{{ t.ip }}</div>
        </td>
        <td class="font-semibold">{{ t.latency }} ms</td>
        <td class="font-semibold">{{ t.loss }}%</td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
defineProps({ targets: Array })
</script>
