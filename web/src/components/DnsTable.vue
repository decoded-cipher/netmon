<template>
  <table class="target-table">
    <thead>
      <tr><th>Host</th><th>Time</th></tr>
    </thead>
    <tbody>
      <tr v-for="d in checks" :key="d.host">
        <td class="font-medium">{{ d.host }}</td>
        <td class="font-semibold">{{ d.time_ms }} ms</td>
      </tr>
    </tbody>
  </table>
</template>

<script setup>
defineProps({ checks: Array })
</script>
