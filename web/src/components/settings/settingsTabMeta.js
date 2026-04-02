/**
 * Single source of truth for settings navigation.
 *
 * To add a section:
 * 1. Add an entry here (id, label, icon SVG).
 * 2. Import the panel component in SettingsModal.vue and register it in the template.
 */

export const DEFAULT_SETTINGS_TAB = 'monitor'

export const SETTINGS_TABS = [
  {
    id: 'monitor',
    label: 'Monitor',
    icon: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>`,
  },
  {
    id: 'about',
    label: 'About',
    icon: `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>`,
  },
]

