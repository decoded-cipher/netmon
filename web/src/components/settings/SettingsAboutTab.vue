<template>
  <div class="ab">
    <header class="ab-hero">
      <div class="ab-hero-bg" aria-hidden="true" />
      <div class="ab-hero-inner">
        <span
          v-if="versionInfo?.current"
          class="ab-ver-badge"
          role="status"
          :aria-label="'Live, version ' + versionInfo.current"
        >
          <span class="ab-ver-badge-dot" aria-hidden="true" />
          <span class="ab-ver-badge-live" aria-hidden="true">Live</span>
          <span class="ab-ver-badge-sep" aria-hidden="true">·</span>
          <span class="ab-ver-badge-ver" aria-hidden="true">{{ versionInfo.current }}</span>
        </span>
        <div class="ab-hero-top">
          <div class="ab-hero-logo" aria-hidden="true">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.65" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><circle cx="12" cy="20" r="1" fill="currentColor" stroke="none"/></svg>
          </div>
          <div class="ab-hero-titles">
            <p class="ab-eyebrow">Network monitor</p>
            <h2 id="ab-intro-heading" class="ab-name">netmon</h2>
          </div>
        </div>
        <p class="ab-lead">
          A lightweight, self-hosted dashboard for latency, jitter, packet loss, DNS resolution, and bandwidth
          — all in your browser, with data kept on the machine that runs the server.
        </p>
      </div>
    </header>

    <div class="ab-updates" aria-label="Updates">
      <p v-if="updateDone" class="ab-msg ab-msg--ok">Update installed — reload to finish.</p>
      <p v-else-if="updating" class="ab-msg ab-msg--muted">Installing update…</p>
      <p v-else-if="updateError" class="ab-msg ab-msg--err">{{ updateError }}</p>

      <div class="ab-updates-actions">
        <template v-if="updateDone">
          <button type="button" class="btn btn-primary" @click="$emit('reload')">Reload</button>
        </template>
        <template v-else-if="!updating">
          <button
            type="button"
            class="btn btn-ghost ab-ico-btn"
            :disabled="checking"
            @click="$emit('check')"
          >
            <svg
              v-if="checking"
              class="ab-ico spin"
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <line x1="12" y1="2" x2="12" y2="6" />
              <line x1="12" y1="18" x2="12" y2="22" />
              <line x1="4.93" y1="4.93" x2="7.76" y2="7.76" />
              <line x1="16.24" y1="16.24" x2="19.07" y2="19.07" />
              <line x1="2" y1="12" x2="6" y2="12" />
              <line x1="18" y1="12" x2="22" y2="12" />
              <line x1="4.93" y1="19.07" x2="7.76" y2="16.24" />
              <line x1="16.24" y1="7.76" x2="19.07" y2="4.93" />
            </svg>
            <svg
              v-else
              class="ab-ico"
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8" />
              <path d="M21 3v5h-5" />
              <path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16" />
              <path d="M8 16H3v5" />
            </svg>
            {{ checking ? 'Checking…' : 'Check for updates' }}
          </button>
          <button
            v-if="versionInfo?.update_available"
            type="button"
            class="btn btn-primary ab-ico-btn"
            @click="$emit('update')"
          >
            <svg
              class="ab-ico"
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <path d="M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 3.5 7.5" />
              <path d="M12 12v9" />
              <path d="m16 16-4 4-4-4" />
            </svg>
            Update to {{ versionInfo.latest }}
          </button>
        </template>
      </div>
    </div>

    <nav class="ab-links" aria-labelledby="ab-links-heading">
      <h3 id="ab-links-heading" class="ab-links-heading">Resources</h3>
      <ul class="ab-linklist">
        <li v-for="(link, i) in projectLinks" :key="i" class="ab-linkitem">
          <a
            :href="link.href"
            target="_blank"
            rel="noopener noreferrer"
            class="ab-linkrow"
          >
            <span
              class="ab-linkrow-ico"
              :class="'ab-linkrow-ico--' + link.tone"
              aria-hidden="true"
              v-html="link.icon"
            />
            <span class="ab-linkrow-body">
              <span class="ab-linkrow-title">{{ link.title }}</span>
              <span class="ab-linkrow-meta">{{ link.meta }}</span>
            </span>
            <span class="ab-linkrow-ext-wrap" aria-hidden="true">
              <svg
                class="ab-linkrow-ext"
                xmlns="http://www.w3.org/2000/svg"
                width="15"
                height="15"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                <polyline points="15 3 21 3 21 9" />
                <line x1="10" y1="14" x2="21" y2="3" />
              </svg>
            </span>
          </a>
        </li>
      </ul>
    </nav>
  </div>
</template>

<script setup>
const REPO = 'https://github.com/decoded-cipher/netmon'
const ISSUES = 'https://github.com/decoded-cipher/netmon/issues'
const LICENSE = 'https://github.com/decoded-cipher/netmon/blob/master/LICENSE'

const svgGithub = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22"/></svg>`
const svgIssue = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>`
const svgFile = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>`

const projectLinks = [
  { href: REPO, title: 'Source code', meta: 'Repository on GitHub', icon: svgGithub, tone: 'slate' },
  { href: ISSUES, title: 'Report an issue', meta: 'GitHub Issues', icon: svgIssue, tone: 'amber' },
  { href: LICENSE, title: 'License', meta: 'MIT — view on GitHub', icon: svgFile, tone: 'muted' },
]

defineProps({
  versionInfo: { type: Object, default: null },
  checking: { type: Boolean, default: false },
  updating: { type: Boolean, default: false },
  updateDone: { type: Boolean, default: false },
  updateError: { type: String, default: '' },
})

defineEmits(['check', 'update', 'reload'])
</script>

<style scoped>
.ab {
  --ab-font: 'Inter', ui-sans-serif, system-ui, sans-serif;
  --ab-mono: ui-monospace, 'SF Mono', 'Cascadia Code', 'Segoe UI Mono', monospace;

  width: 100%;
  max-width: 42rem;
  display: flex;
  flex-direction: column;
  gap: 0;
  font-family: var(--ab-font);
  font-feature-settings: 'kern' 1, 'liga' 1;
  -webkit-font-smoothing: antialiased;
}

.ab-hero {
  position: relative;
  border-radius: 10px;
  border: 1px solid var(--border);
  overflow: hidden;
  box-shadow: 0 1px 2px var(--shadow);
}

.ab-hero-bg {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 120% 80% at 100% -20%, color-mix(in srgb, var(--accent) 18%, transparent), transparent 50%),
    radial-gradient(ellipse 80% 60% at 0% 100%, color-mix(in srgb, var(--purple) 12%, transparent), transparent 45%),
    linear-gradient(165deg, color-mix(in srgb, var(--fg) 3.5%, var(--card)), var(--card));
  pointer-events: none;
}

.ab-hero-inner {
  position: relative;
  padding: 0.85rem 1rem 0.95rem;
}

.ab-ver-badge {
  position: absolute;
  top: 0.65rem;
  right: 0.65rem;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  max-width: min(14.5rem, calc(100% - 1.25rem));
  padding: 0.32rem 0.7rem 0.32rem 0.55rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--green) 32%, var(--border));
  background: linear-gradient(
    165deg,
    color-mix(in srgb, var(--green-soft) 85%, var(--card)),
    var(--card)
  );
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.05) inset;
  line-height: 1;
  overflow: hidden;
  min-width: 0;
}

.ab-ver-badge-dot {
  flex-shrink: 0;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--green);
  box-shadow: 0 0 0 0 color-mix(in srgb, var(--green) 45%, transparent);
  animation: ab-ver-live-pulse 2.2s ease-in-out infinite;
}

.ab-ver-badge-live {
  flex-shrink: 0;
  font-size: 0.5625rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--green);
}

.ab-ver-badge-sep {
  flex-shrink: 0;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--muted);
  opacity: 0.65;
  margin: 0 -0.05rem;
}

.ab-ver-badge-ver {
  min-width: 0;
  font-family: var(--ab-mono);
  font-size: 0.6875rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.03em;
  color: var(--fg);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (prefers-reduced-motion: reduce) {
  .ab-ver-badge-dot {
    animation: none;
    box-shadow: none;
  }
}

@keyframes ab-ver-live-pulse {
  0%,
  100% {
    opacity: 1;
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--green) 40%, transparent);
  }
  50% {
    opacity: 0.92;
    box-shadow: 0 0 0 5px transparent;
  }
}

.ab-hero-top {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.ab-hero-logo {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: linear-gradient(145deg, var(--accent-soft), color-mix(in srgb, var(--purple) 12%, transparent));
  color: var(--accent);
  border: 1px solid color-mix(in srgb, var(--accent) 22%, var(--border));
  box-shadow: 0 3px 14px color-mix(in srgb, var(--accent) 10%, transparent);
}

.ab-hero-titles {
  min-width: 0;
}

.ab-eyebrow {
  margin: 0;
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--muted);
  line-height: 1.3;
}

.ab-name {
  margin: 0.12rem 0 0;
  font-size: 1.3125rem;
  font-weight: 700;
  letter-spacing: -0.035em;
  color: var(--fg);
  line-height: 1.15;
}

.ab-lead {
  margin: 0.6rem 0 0;
  font-size: 0.8125rem;
  font-weight: 400;
  line-height: 1.5;
  color: var(--subtle);
}

.ab-updates {
  margin-top: 1.5rem;
  padding-top: 0.5rem;
  padding-bottom: 0.35rem;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.ab-msg {
  margin: 0;
  font-size: 0.75rem;
  font-weight: 500;
  line-height: 1.4;
}

.ab-msg--ok {
  color: var(--green);
}

.ab-msg--muted {
  color: var(--muted);
  font-weight: 400;
}

.ab-msg--err {
  color: var(--red);
  font-weight: 500;
}

.ab-updates-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  padding-top: 0.65rem;
  padding-bottom: 0.65rem;
}

.ab-updates-actions .btn {
  font-size: 0.75rem;
  font-weight: 500;
}

.ab-ico-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.ab-ico {
  flex-shrink: 0;
  opacity: 0.9;
}

.ab-links {
  margin-top: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.ab-links-heading {
  margin: 0;
  font-size: 0.625rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--muted);
  line-height: 1.2;
}

.ab-linklist {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.ab-linkitem {
  margin: 0;
  border-radius: 11px;
  border: 1px solid var(--border);
  background: linear-gradient(
    165deg,
    color-mix(in srgb, var(--fg) 2.5%, var(--card)),
    var(--card)
  );
  box-shadow: 0 1px 2px var(--shadow);
  overflow: hidden;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.ab-linkitem:has(.ab-linkrow:hover) {
  border-color: color-mix(in srgb, var(--accent) 22%, var(--border));
  box-shadow:
    0 1px 2px var(--shadow),
    0 4px 14px color-mix(in srgb, var(--accent) 6%, transparent);
}

.ab-linkrow {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0.65rem 0.75rem;
  text-decoration: none;
  color: inherit;
  transition: background 0.15s ease;
  font-size: 0.8125rem;
}

.ab-linkrow:hover {
  background: color-mix(in srgb, var(--accent) 5%, var(--input-hover-bg));
}

.ab-linkrow:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--accent) 55%, transparent);
  outline-offset: 2px;
}

.ab-linkrow-ico {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.04) inset;
}

.ab-linkrow-ico--slate {
  background: linear-gradient(
    155deg,
    color-mix(in srgb, var(--fg) 8%, transparent),
    color-mix(in srgb, var(--fg) 4%, transparent)
  );
  color: var(--subtle);
  border: 1px solid color-mix(in srgb, var(--fg) 12%, var(--border));
}

.ab-linkrow-ico--amber {
  background: linear-gradient(
    155deg,
    color-mix(in srgb, var(--yellow) 14%, var(--card)),
    var(--yellow-soft)
  );
  color: var(--yellow);
  border: 1px solid color-mix(in srgb, var(--yellow) 30%, var(--border));
}

.ab-linkrow-ico--muted {
  background: linear-gradient(
    155deg,
    color-mix(in srgb, var(--fg) 6%, transparent),
    color-mix(in srgb, var(--fg) 3%, transparent)
  );
  color: var(--muted);
  border: 1px solid var(--border);
}

.ab-linkrow-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.ab-linkrow-title {
  font-size: 0.8125rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--fg);
  line-height: 1.25;
}

.ab-linkrow-meta {
  font-size: 0.6875rem;
  font-weight: 400;
  line-height: 1.35;
  color: var(--muted);
}

.ab-linkrow-ext-wrap {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: color-mix(in srgb, var(--fg) 4%, transparent);
  border: 1px solid color-mix(in srgb, var(--border) 80%, transparent);
}

.ab-linkrow-ext {
  color: var(--muted);
  opacity: 0.75;
  transition: opacity 0.15s ease, color 0.15s ease, transform 0.15s ease;
}

.ab-linkrow:hover .ab-linkrow-ext {
  opacity: 1;
  color: var(--accent);
}

.ab-linkrow:hover .ab-linkrow-ext-wrap {
  border-color: color-mix(in srgb, var(--accent) 25%, var(--border));
  background: color-mix(in srgb, var(--accent) 8%, transparent);
}

.ab-linkrow:hover .ab-linkrow-ext {
  transform: translate(1px, -1px);
}
</style>
