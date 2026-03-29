<template>
  <table class="target-table">
    <thead v-if="checks?.length">
      <tr><th>Host</th><th>Time</th></tr>
    </thead>
    <tbody>
      <tr v-if="!checks?.length">
        <td colspan="2" class="text-center py-5" style="color:var(--muted);border-bottom:none">
          <div class="flex flex-col items-center gap-1.5">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" style="opacity:0.45"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
            <span style="font-size:0.6875rem;font-weight:600;opacity:0.6">No lookups yet</span>
          </div>
        </td>
      </tr>
      <tr v-else v-for="d in checks" :key="d.host">
        <td class="font-medium">{{ d.host }}</td>
        <td class="font-semibold">{{ d.time_ms }} ms</td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
defineProps({ checks: Array })
</script>
