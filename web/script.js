const $ = (s) => document.querySelector(s);
let chartInstances = {};
let lastUpdateTime = null;
let isLoading = false;
let currentMinutes = 60;
let prevSummary = null;
let prevStatus  = null;
let refreshTimerStart = null;
const REFRESH_MS = 15000;

function getChartHeight() {
  if (window.innerWidth >= 1024) {
    const grid = document.getElementById('chartGrid');
    if (grid && grid.offsetHeight > 0) {
      // 2 rows with 12px gap; each card has ~52px overhead (padding + header)
      return Math.max(80, Math.floor((grid.offsetHeight - 12) / 2) - 52);
    }
  }
  return window.innerWidth < 640 ? 160 : 185;
}

// ── Live "Xs ago" ticker ──────────────────────────────────────────────
setInterval(() => {
  if (!lastUpdateTime) return;
  const sec = Math.floor((Date.now() - lastUpdateTime) / 1000);
  let label;
  if (sec < 5)        label = "just now";
  else if (sec < 60)  label = sec + "s ago";
  else if (sec < 3600) label = Math.floor(sec / 60) + "m ago";
  else                label = Math.floor(sec / 3600) + "h ago";
  const el = $("#lastUpdated");
  if (el && !el.classList.contains("refreshing")) el.textContent = label;
}, 1000);

// ── Refresh bar ticker ────────────────────────────────────────────────
setInterval(() => {
  if (!refreshTimerStart) return;
  const pct = Math.min(100, ((Date.now() - refreshTimerStart) / REFRESH_MS) * 100);
  const fill = document.getElementById('refreshBarFill');
  if (fill) fill.style.width = pct + '%';
}, 100);

// ── SVG icons ─────────────────────────────────────────────────────────
const SVG = {
  clock:     `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>`,
  arrowDown: `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><polyline points="19 12 12 19 5 12"/></svg>`,
  arrowUp:   `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/></svg>`,
  activity:  `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>`,
};

// ── Data loading ──────────────────────────────────────────────────────
async function loadData(forceRecreate = false) {
  if (isLoading) return;
  isLoading = true;
  const lu = $("#lastUpdated");
  if (lu) { lu.classList.add("refreshing"); lu.textContent = "Refreshing…"; }

  const [data, cfg] = await Promise.all([
    fetch(`/api/data?minutes=${currentMinutes}`).then(r => r.json()).catch(() => null),
    fetch("/api/config").then(r => r.json()).catch(() => null),
  ]);

  isLoading = false;
  if (lu) lu.classList.remove("refreshing");

  if (!data) {
    if (lu) lu.textContent = "Error";
    return;
  }

  lastUpdateTime = Date.now();
  refreshTimerStart = Date.now();
  checkThresholds(data.summary, prevSummary);

  // Filter to current config only — gateway is excluded from the table
  // (it's a local router probe, not an internet connectivity target)
  if (cfg) {
    const allowedPing = new Set(cfg.ping_targets.map(h => h.toLowerCase()));
    data.targets = (data.targets || []).filter(t => allowedPing.has(t.host.toLowerCase()));
    const allowedDns = new Set(cfg.dns_targets.map(h => h.toLowerCase()));
    data.dns = (data.dns || []).filter(d => allowedDns.has(d.host.toLowerCase()));
  }

  renderHeader(data, cfg);
  renderSummary(data.summary, data.history);
  renderCharts(data.history, data.summary, forceRecreate);
  renderHealth(data.summary);
  renderTargets(data.targets);
  renderDNS(data.dns);
  renderQuickStats(data.summary, data.history);
  prevSummary = data.summary;
}

// ── Header ────────────────────────────────────────────────────────────
function renderHeader(data, cfg) {
  const targetCount = cfg ? cfg.ping_targets.length : (data.targets || []).length;
  const intervalLabel = cfg ? ` · ping ${cfg.ping_interval_s}s` : "";
  $("#targetCount").textContent = targetCount;
  $("#networkLabel").textContent = (data.network_id || "unknown") + intervalLabel;

  const loss = data.summary.packet_loss;
  const badge = $("#statusBadge");
  if (loss > 5) {
    badge.className = "status-badge offline";
    badge.innerHTML = '<span class="status-dot"></span>Offline';
  } else if (loss > 1) {
    badge.className = "status-badge degraded";
    badge.innerHTML = '<span class="status-dot"></span>Degraded';
  } else {
    badge.className = "status-badge online";
    badge.innerHTML = '<span class="status-dot"></span>Online';
  }
}

// ── Summary cards ─────────────────────────────────────────────────────
function renderSummary(s, history) {
  const lossColor = s.packet_loss > 1 ? "red" : "green";
  const last = (history || []).slice(-20);
  const sc = { blue: "#3b82f6", green: "#10b981", purple: "#8b5cf6", red: "#ef4444" };
  const p = prevSummary;
  const items = [
    { label: "Avg Latency",  value: s.latency_avg + " ms",   icon: SVG.clock,     color: "blue",    sub: "p95 " + s.latency_p95 + " ms",           delta: deltaStr(s.latency_avg,  p ? p.latency_avg  : null, true),  spark: sparklineSVG(last.map(h => h.latency),  sc.blue)   },
    { label: "Download",     value: s.download_avg + " Mbps", icon: SVG.arrowDown, color: "green",   sub: "avg throughput",                          delta: deltaStr(s.download_avg, p ? p.download_avg : null, false), spark: sparklineSVG(last.map(h => h.download), sc.green)  },
    { label: "Upload",       value: s.upload_avg + " Mbps",   icon: SVG.arrowUp,   color: "purple",  sub: "avg throughput",                          delta: deltaStr(s.upload_avg,   p ? p.upload_avg   : null, false), spark: sparklineSVG(last.map(h => h.upload),   sc.purple) },
    { label: "Packet Loss",  value: s.packet_loss + "%",      icon: SVG.activity,  color: lossColor, sub: s.packet_loss < 1 ? "healthy" : "elevated", delta: deltaStr(s.packet_loss,  p ? p.packet_loss  : null, true),  spark: sparklineSVG(last.map(h => h.loss),     sc[lossColor]) },
  ];
  $("#summaryCards").innerHTML = items.map(m =>
    `<div class="card metric-card" data-color="${m.color}">
      <div class="flex items-center justify-between">
        <div class="metric-label">${m.label}</div>
        <div class="metric-icon ${m.color}">${m.icon}</div>
      </div>
      <div class="metric-value">${m.value}${m.delta}</div>
      <div class="metric-sub">${m.sub}</div>
      ${m.spark}
    </div>`
  ).join("");
  // Flash each card to signal new data
  document.querySelectorAll("#summaryCards .card").forEach(c => {
    c.classList.remove("flash-update");
    void c.offsetWidth;
    c.classList.add("flash-update");
  });
}

// ── Charts ────────────────────────────────────────────────────────────
function renderCharts(history, summary, forceRecreate = false) {
  if (!history || history.length === 0) return;

  const times      = history.map(h => h.time);
  const isSmall    = window.innerWidth < 640;
  const chartH     = getChartHeight();
  const tickCount  = Math.min(times.length, isSmall ? 4 : 6);
  const dark       = document.body.classList.contains("dark-mode");
  const labelStyle = { colors: "var(--muted)", fontSize: isSmall ? "9px" : "10px" };

  $("#latencyAvgLabel").textContent = "avg " + summary.latency_avg + " ms";
  $("#speedAvgLabel").textContent   = "↓" + summary.download_avg + "  ↑" + summary.upload_avg + " Mbps";
  $("#lossAvgLabel").textContent    = "loss " + summary.packet_loss + "% · jitter " + summary.jitter_avg + " ms";
  $("#dnsAvgLabel").textContent     = "avg " + summary.dns_avg + " ms";

  // Smooth update when charts already exist (no flicker)
  if (!forceRecreate && chartInstances.latency) {
    const xUpd = { xaxis: { categories: times, tickAmount: tickCount, labels: { style: labelStyle } } };
    chartInstances.latency.updateOptions(xUpd, false, true);
    chartInstances.latency.updateSeries([{ name: "Latency (ms)", data: history.map(h => h.latency) }], true);
    chartInstances.speed.updateOptions(xUpd, false, true);
    chartInstances.speed.updateSeries([
      { name: "Download", data: history.map(h => h.download) },
      { name: "Upload",   data: history.map(h => h.upload) },
    ], true);
    chartInstances.loss.updateOptions(xUpd, false, true);
    chartInstances.loss.updateSeries([
      { name: "Loss (%)",   data: history.map(h => h.loss) },
      { name: "Jitter (ms)", data: history.map(h => h.jitter) },
    ], true);
    chartInstances.dns.updateOptions(xUpd, false, true);
    chartInstances.dns.updateSeries([{ name: "DNS (ms)", data: history.map(h => h.dns) }], true);
    return;
  }

  // Full create (first load or after theme change)
  Object.values(chartInstances).forEach(c => c.destroy());
  chartInstances = {};

  const base = {
    chart: { type: "area", background: "transparent", fontFamily: "'Inter',sans-serif", toolbar: { show: false }, height: chartH, animations: { speed: 500 }, zoom: { enabled: false } },
    stroke: { curve: "smooth", width: 2 },
    markers: { size: times.length <= 5 ? 4 : 0, hover: { size: 5 } },
    grid: { borderColor: "var(--border)", xaxis: { lines: { show: false } }, yaxis: { lines: { show: true } }, padding: { top: -8, right: 8, bottom: 0, left: 8 } },
    xaxis: {
      categories: times, tickAmount: tickCount,
      labels: { style: labelStyle, rotate: 0, hideOverlappingLabels: true, maxHeight: 30 },
      axisBorder: { show: false }, axisTicks: { show: false },
    },
    yaxis: { labels: { style: labelStyle }, forceNiceScale: true },
    legend: { position: "top", fontSize: "10px", labels: { colors: "var(--subtle)" }, markers: { width: 6, height: 6, radius: 2 }, itemMargin: { horizontal: 8 } },
    tooltip: { theme: dark ? "dark" : "light", style: { fontSize: "11px" }, x: { show: true } },
    dataLabels: { enabled: false },
  };
  const fillArea = (o = 0.12) => ({ type: "gradient", gradient: { shadeIntensity: 0, opacityFrom: o, opacityTo: 0.01, stops: [0, 95] } });

  chartInstances.latency = new ApexCharts($("#latencyChart"), { ...base, series: [{ name: "Latency (ms)", data: history.map(h => h.latency) }], colors: ["#3b82f6"], fill: fillArea(), yaxis: { ...base.yaxis, min: 0 } });
  chartInstances.latency.render();

  chartInstances.speed = new ApexCharts($("#speedChart"), { ...base, series: [{ name: "Download", data: history.map(h => h.download) }, { name: "Upload", data: history.map(h => h.upload) }], colors: ["#10b981", "#8b5cf6"], fill: fillArea(0.08) });
  chartInstances.speed.render();

  chartInstances.loss = new ApexCharts($("#lossChart"), {
    ...base,
    series: [{ name: "Loss (%)", data: history.map(h => h.loss) }, { name: "Jitter (ms)", data: history.map(h => h.jitter) }],
    colors: ["#ef4444", "#f59e0b"], fill: fillArea(0.1),
    yaxis: [
      { forceNiceScale: true, min: 0, title: { text: "Loss %",   style: { fontSize: "10px", color: "var(--muted)" } }, labels: { style: { colors: "var(--muted)", fontSize: "10px" } } },
      { opposite: true, forceNiceScale: true, min: 0, title: { text: "Jitter ms", style: { fontSize: "10px", color: "var(--muted)" } }, labels: { style: { colors: "var(--muted)", fontSize: "10px" } } },
    ],
  });
  chartInstances.loss.render();

  chartInstances.dns = new ApexCharts($("#dnsChart"), { ...base, series: [{ name: "DNS (ms)", data: history.map(h => h.dns) }], colors: ["#8b5cf6"], fill: fillArea(), yaxis: { ...base.yaxis, min: 0 } });
  chartInstances.dns.render();
}

// ── Health ring ───────────────────────────────────────────────────────
function renderHealth(s) {
  let score = 100;
  if (s.latency_avg > 50) score -= 20; else if (s.latency_avg > 30) score -= 10; else if (s.latency_avg > 20) score -= 5;
  if (s.packet_loss > 5) score -= 30; else if (s.packet_loss > 1) score -= 15; else if (s.packet_loss > 0.5) score -= 5;
  if (s.download_avg < 25) score -= 20; else if (s.download_avg < 50) score -= 10;
  if (s.jitter_avg > 10) score -= 10; else if (s.jitter_avg > 5) score -= 5;
  score = Math.max(0, Math.min(100, score));

  const circ = 2 * Math.PI * 42;
  const arc  = $("#healthArc");
  const color = score >= 80 ? "var(--green)" : score >= 50 ? "var(--yellow)" : "var(--red)";
  arc.style.stroke = color;
  requestAnimationFrame(() => { arc.style.strokeDashoffset = circ - (score / 100) * circ; });
  $("#healthNum").textContent = score;
  $("#healthNum").style.color = color;

  const rate = (val, good, warn) => val <= good ? { text: "Good", color: "var(--green)" } : val <= warn ? { text: "Fair", color: "var(--yellow)" } : { text: "Poor", color: "var(--red)" };
  const lat = rate(s.latency_avg, 20, 50);
  $("#hbLatency").textContent = lat.text; $("#hbLatency").style.color = lat.color;
  const loss = rate(s.packet_loss, 0.5, 2);
  $("#hbLoss").textContent = loss.text; $("#hbLoss").style.color = loss.color;
  const spd = s.download_avg >= 50 ? { text: "Good", color: "var(--green)" } : s.download_avg >= 25 ? { text: "Fair", color: "var(--yellow)" } : { text: "Poor", color: "var(--red)" };
  $("#hbSpeed").textContent = spd.text; $("#hbSpeed").style.color = spd.color;
}

// ── Tables ────────────────────────────────────────────────────────────
function renderTargets(targets) {
  if (!targets) return;
  $("#targetBody").innerHTML = targets.map(t =>
    `<tr>
      <td>
        <div class="host-cell"><span class="inline-dot ${t.status === 'up' ? 'up' : 'down'}"></span><span>${t.host}</span></div>
        <div class="ip-cell">${t.ip}</div>
      </td>
      <td class="font-semibold">${t.latency} ms</td>
      <td>${t.loss}%</td>
    </tr>`
  ).join("");
}

function renderDNS(dns) {
  if (!dns) return;
  $("#dnsBody").innerHTML = dns.map(d =>
    `<tr>
      <td class="font-medium">${d.host}</td>
      <td class="font-semibold">${d.time_ms} ms</td>
    </tr>`
  ).join("");
}

function renderQuickStats(s, history) {
  if (!history || history.length === 0) return;
  const peakDl = Math.max(...history.map(h => h.download));
  $("#qs-uptime").textContent   = s.uptime_24h + "%";
  $("#qs-outages").textContent  = s.outages_24h;
  $("#qs-peak-dl").textContent  = peakDl + " Mbps";
  $("#qs-min-lat").textContent  = (s.latency_min || Math.min(...history.map(h => h.latency))) + " ms";
  $("#qs-max-lat").textContent  = (s.latency_max || Math.max(...history.map(h => h.latency))) + " ms";
  $("#qs-jitter").textContent   = s.jitter_avg + " ms";
}

// ── Sparkline helper ──────────────────────────────────────────────────
function sparklineSVG(values, color) {
  if (!values || values.length < 2) return '';
  const clean = values.filter(v => v != null && !isNaN(v) && v >= 0);
  if (clean.length < 2) return '';
  const w = 72, h = 32;
  const min = Math.min(...clean), max = Math.max(...clean);
  const range = max - min || 1;
  const pts = clean.map((v, i) => {
    const x = (i / (clean.length - 1)) * w;
    const y = (h - 2) - ((v - min) / range) * (h - 6);
    return x.toFixed(1) + ',' + y.toFixed(1);
  }).join(' ');
  return '<svg width="' + w + '" height="' + h + '" style="position:absolute;bottom:6px;right:10px;pointer-events:none" xmlns="http://www.w3.org/2000/svg">'
    + '<polyline points="' + pts + '" fill="none" stroke="' + color + '" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" opacity="0.35"/></svg>';
}

// ── Delta indicator helper ─────────────────────────────────────────────
function deltaStr(curr, prev, lowerIsBetter) {
  if (prev == null) return '';
  const diff = curr - prev;
  if (Math.abs(diff) < 0.05) return '';
  const better = lowerIsBetter ? diff < 0 : diff > 0;
  const color  = better ? 'var(--green)' : 'var(--red)';
  const arrow  = diff > 0 ? '↑' : '↓';
  const abs    = Math.abs(diff);
  const disp   = abs < 10 ? abs.toFixed(1) : Math.round(abs);
  return '<span class="metric-delta" style="color:' + color + '">' + arrow + disp + '</span>';
}

// ── Toast notifications ────────────────────────────────────────────────
const activeToasts = new Set();

function showToast(id, title, msg, type, duration) {
  duration = duration || 5000;
  if (activeToasts.has(id)) return;
  activeToasts.add(id);
  const container = document.getElementById('toastContainer');
  if (!container) return;
  const icons = {
    warn:  '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
    error: '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>',
    ok:    '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>',
  };
  const el = document.createElement('div');
  el.className = 'toast ' + type;
  el.innerHTML = '<span class="toast-icon">' + (icons[type] || icons.warn) + '</span>'
    + '<div class="toast-body"><div class="toast-title">' + title + '</div>'
    + (msg ? '<div class="toast-msg">' + msg + '</div>' : '') + '</div>';
  container.appendChild(el);
  const dismiss = () => {
    el.classList.add('dismiss');
    setTimeout(() => { el.remove(); activeToasts.delete(id); }, 250);
  };
  el.addEventListener('click', dismiss);
  setTimeout(dismiss, duration);
}

// ── Threshold checks ──────────────────────────────────────────────────
function checkThresholds(s, prev) {
  const curr = s.packet_loss > 5 ? 'offline' : s.packet_loss > 1 ? 'degraded' : 'online';
  if (prevStatus !== null && prevStatus !== curr) {
    if (curr === 'offline')
      showToast('st-offline', 'Connection Offline', 'Packet loss at ' + s.packet_loss + '%', 'error', 8000);
    else if (curr === 'degraded')
      showToast('st-degraded', 'Connection Degraded', 'Packet loss at ' + s.packet_loss + '%', 'warn', 7000);
    else
      showToast('st-ok', 'Connection Restored', 'All metrics back to normal', 'ok', 5000);
  }
  prevStatus = curr;
  if (prev && s.latency_avg > 150 && prev.latency_avg <= 150)
    showToast('lat-spike', 'Latency Spike', s.latency_avg + 'ms avg (was ' + prev.latency_avg + 'ms)', 'warn', 6000);
}

// ── Duration picker ───────────────────────────────────────────────────
const durationBtn   = $("#durationBtn");
const durationPanel = $("#durationPanel");
const durationLabel = $("#durationLabel");

function setDuration(minutes, label, reload = true) {
  currentMinutes = minutes;
  durationLabel.textContent = label;
  document.querySelectorAll(".dur-chip").forEach(c => {
    c.classList.toggle("active", parseInt(c.dataset.minutes) === minutes);
  });
  closeDurationPanel();
  if (reload) loadData(true);
}

function closeDurationPanel() {
  durationPanel.classList.remove("open");
  durationBtn.classList.remove("open");
}

durationBtn.addEventListener("click", e => {
  e.stopPropagation();
  const isOpen = durationPanel.classList.toggle("open");
  durationBtn.classList.toggle("open", isOpen);
});

document.querySelectorAll(".dur-chip").forEach(chip => {
  chip.addEventListener("click", () => {
    setDuration(parseInt(chip.dataset.minutes), chip.dataset.label);
  });
});

$("#durApply").addEventListener("click", () => {
  const n    = Math.max(1, parseInt($("#durNum").value) || 1);
  const unit = parseInt($("#durUnit").value);
  const mins = n * unit;
  const unitLabel = unit === 1 ? "m" : unit === 60 ? "h" : "d";
  setDuration(mins, n + unitLabel);
});

$("#durNum").addEventListener("keydown", e => { if (e.key === "Enter") $("#durApply").click(); });

document.addEventListener("click", e => {
  if (!$("#durationWrapper").contains(e.target)) closeDurationPanel();
});

// ── Theme toggle ──────────────────────────────────────────────────────
const toggle = $("#themeToggle");
function applyTheme(dark) {
  document.body.classList.toggle("dark-mode", dark);
  $("#iconMoon").style.display = dark ? "none" : "";
  $("#iconSun").style.display  = dark ? "" : "none";
}
applyTheme(localStorage.getItem("theme") === "dark");
toggle.addEventListener("click", () => {
  const dark = document.body.classList.toggle("dark-mode");
  applyTheme(dark);
  localStorage.setItem("theme", dark ? "dark" : "light");
  loadData(true); // recreate charts so tooltip theme updates
});

// ── Settings modal ────────────────────────────────────────────────────
const backdrop = $("#modalBackdrop");

function openSettings() { backdrop.classList.add("open"); loadConfig(); }
function closeSettings() { backdrop.classList.remove("open"); $("#saveStatus").textContent = ""; }

async function loadConfig() {
  try {
    const cfg = await fetch("/api/config").then(r => r.json());
    $("#cfgPingTargets").value  = (cfg.ping_targets || []).join("\n");
    $("#cfgDnsTargets").value   = (cfg.dns_targets  || []).join("\n");
    $("#cfgPingInterval").value = cfg.ping_interval_s  || 60;
    $("#cfgSpeedInterval").value = cfg.speed_interval_m || 30;
    $("#cfgPingCount").value    = cfg.ping_count || 5;
  } catch { showStatus("Failed to load config", false); }
}

async function saveConfig() {
  const saveBtn = $("#saveBtn");
  saveBtn.disabled = true;
  const pingTargets  = $("#cfgPingTargets").value.split("\n").map(s => s.trim()).filter(Boolean);
  const dnsTargets   = $("#cfgDnsTargets").value.split("\n").map(s => s.trim()).filter(Boolean);
  const pingIntervalS  = parseInt($("#cfgPingInterval").value, 10);
  const speedIntervalM = parseInt($("#cfgSpeedInterval").value, 10);
  const pingCount    = parseInt($("#cfgPingCount").value, 10);

  if (!pingTargets.length)           { showStatus("Probe hosts cannot be empty", false); saveBtn.disabled = false; return; }
  if (pingIntervalS < 10)            { showStatus("Ping interval must be ≥ 10 s", false); saveBtn.disabled = false; return; }
  if (speedIntervalM < 5)            { showStatus("Speed interval must be ≥ 5 min", false); saveBtn.disabled = false; return; }
  if (pingCount < 1 || pingCount > 20) { showStatus("Ping count must be 1–20", false); saveBtn.disabled = false; return; }

  try {
    const res = await fetch("/api/config", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ping_targets: pingTargets, dns_targets: dnsTargets, ping_interval_s: pingIntervalS, speed_interval_m: speedIntervalM, ping_count: pingCount }),
    });
    const body = await res.json();
    if (res.ok) { showStatus("Saved!", true); setTimeout(closeSettings, 900); }
    else showStatus(body.error || "Save failed", false);
  } catch { showStatus("Network error", false); }

  saveBtn.disabled = false;
}

function showStatus(msg, ok) {
  const el = $("#saveStatus");
  el.textContent = msg;
  el.className = "save-status " + (ok ? "ok" : "err");
}

$("#settingsBtn").addEventListener("click", openSettings);
$("#modalClose").addEventListener("click", closeSettings);
$("#modalCancel").addEventListener("click", closeSettings);
$("#saveBtn").addEventListener("click", saveConfig);
backdrop.addEventListener("click", e => { if (e.target === backdrop) closeSettings(); });
document.addEventListener("keydown", e => { if (e.key === "Escape") closeSettings(); });

// ── Init ──────────────────────────────────────────────────────────────
document.addEventListener("DOMContentLoaded", () => {
  loadData();
  setInterval(loadData, 15000); // poll every 15s
});

let resizeTimer;
window.addEventListener("resize", () => {
  clearTimeout(resizeTimer);
  resizeTimer = setTimeout(() => loadData(true), 250);
});
