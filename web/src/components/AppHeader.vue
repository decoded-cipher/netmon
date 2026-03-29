<template>
  <header class="flex justify-between items-center gap-2 mb-3 flex-shrink-0 min-w-0">
    <div class="flex items-center gap-2.5 min-w-0">
      <div class="app-logo flex-shrink-0">
        <!-- WiFi -->
        <svg v-if="connType !== 'ethernet'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><circle cx="12" cy="20" r="1" fill="currentColor" stroke="none"/></svg>
        <!-- Ethernet -->
        <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="16" y="16" width="6" height="6" rx="1"/><rect x="2" y="16" width="6" height="6" rx="1"/><rect x="9" y="2" width="6" height="6" rx="1"/><path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"/><path d="M12 12V8"/></svg>
      </div>
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <h1 class="text-lg sm:text-2xl font-extrabold tracking-tight leading-none">netmon</h1>
          <span class="status-badge flex-shrink-0" :class="status">
            <span class="status-dot" />
            {{ status.charAt(0).toUpperCase() + status.slice(1) }}
          </span>
        </div>
        <p class="text-xs mt-0.5 flex items-center gap-1.5 truncate" style="color:var(--subtle)">
          <span class="live-dot flex-shrink-0" />
          <span class="font-semibold flex-shrink-0" style="color:var(--fg)">{{ targetCount }}</span>
          <span class="flex-shrink-0">probes</span>
          <span class="flex-shrink-0" style="color:var(--border)">·</span>
          <span class="flex-shrink-0" :class="{ refreshing: isRefreshing }" style="color:var(--muted)">
            {{ isRefreshing ? 'Refreshing…' : agoLabel }}
          </span>
          <span class="hidden sm:inline flex-shrink-0" style="color:var(--border)">·</span>
          <span class="font-semibold hidden sm:inline flex-shrink-0" style="color:var(--accent)">{{ networkLabel }}</span>
        </p>
      </div>
    </div>

    <div class="flex gap-1.5 items-center flex-shrink-0">
      <slot name="duration-picker" />

      <button class="theme-toggle icon-btn" :title="isDark ? 'Switch to light mode' : 'Switch to dark mode'" @click="$emit('toggle-theme')">
        <svg v-if="!isDark" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
      </button>

      <button class="settings-btn icon-btn" title="Settings" @click="$emit('open-settings')">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
      </button>
    </div>
  </header>
</template>

<script setup>
defineProps({
  status:       { type: String, default: 'online' },
  targetCount:  { type: Number, default: 0 },
  networkLabel: { type: String, default: '—' },
  agoLabel:     { type: String, default: '—' },
  isRefreshing: { type: Boolean, default: false },
  isDark:       { type: Boolean, default: false },
  currentMinutes: { type: Number, default: 60 },
  connType:     { type: String, default: '' },
})

defineEmits(['toggle-theme', 'open-settings', 'set-duration'])
</script>
