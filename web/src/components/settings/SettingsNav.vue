<template>
  <nav
    class="sn"
    role="tablist"
    :aria-label="ariaLabel"
  >
    <button
      v-for="tab in tabs"
      :key="tab.id"
      type="button"
      role="tab"
      class="sn-item"
      :class="{ active: modelValue === tab.id }"
      :aria-selected="modelValue === tab.id"
      :aria-controls="panelIdPrefix + tab.id"
      :id="tabIdPrefix + tab.id"
      :tabindex="modelValue === tab.id ? 0 : -1"
      @click="$emit('update:modelValue', tab.id)"
    >
      <span class="sn-ico" aria-hidden="true" v-html="tab.icon" />
      <span class="sn-label">{{ tab.label }}</span>
    </button>
  </nav>
</template>

<script setup>
defineProps({
  tabs: { type: Array, required: true },
  modelValue: { type: String, required: true },
  ariaLabel: { type: String, default: 'Settings sections' },
  tabIdPrefix: { type: String, default: 'settings-tab-' },
  panelIdPrefix: { type: String, default: 'settings-panel-' },
})

defineEmits(['update:modelValue'])
</script>

<style scoped>
.sn {
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 0.65rem 0.45rem;
  min-width: 0;
}

.sn-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  text-align: left;
  padding: 0.45rem 0.5rem;
  border: 1px solid transparent;
  border-radius: 10px;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  font-family: inherit;
  transition: color 0.12s ease, background 0.12s ease, border-color 0.12s ease;
}

.sn-item:hover {
  color: var(--fg);
  background: color-mix(in srgb, var(--fg) 5%, transparent);
  border-color: var(--border);
}

.sn-item:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--accent) 50%, transparent);
  outline-offset: 1px;
}

.sn-item.active {
  color: var(--fg);
  background: var(--accent-soft);
  border-color: color-mix(in srgb, var(--accent) 28%, var(--border));
}

.sn-item.active .sn-ico {
  color: var(--accent);
}

.sn-ico {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  color: inherit;
}

.sn-ico :deep(svg) {
  display: block;
  width: 17px;
  height: 17px;
}

.sn-label {
  min-width: 0;
  font-size: 0.75rem;
  font-weight: 500;
  letter-spacing: -0.01em;
  line-height: 1.35;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sn-item.active .sn-label {
  font-weight: 600;
}

@media (max-width: 560px) {
  .sn {
    flex-direction: row;
    flex-wrap: nowrap;
    overflow-x: auto;
    overflow-y: hidden;
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;
    padding: 0.55rem 0.5rem;
    gap: 6px;
    border-bottom: 1px solid var(--border);
  }

  .sn::-webkit-scrollbar {
    display: none;
  }

  .sn-item {
    flex: 0 0 auto;
    width: auto;
    padding: 0.45rem 0.75rem;
    border-radius: 10px;
  }

  .sn-label {
    font-size: 0.78rem;
  }
}
</style>
