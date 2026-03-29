<template>
  <table class="target-table">
    <thead v-if="targets?.length">
      <tr><th>Host</th><th>Latency</th><th>Loss</th></tr>
    </thead>
    <tbody>
      <tr v-if="!targets?.length">
        <td colspan="3" class="text-center py-5" style="color:var(--muted);border-bottom:none">
          <div class="flex flex-col items-center gap-1.5">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" style="opacity:0.45"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>
            <span style="font-size:0.6875rem;font-weight:600;opacity:0.6">No targets yet</span>
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
        <td>{{ t.loss }}%</td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
defineProps({ targets: Array })
</script>
