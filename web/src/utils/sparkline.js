export function sparklineSVG(values, color) {
  if (!values || values.length < 2) return ''
  const clean = values.filter(v => v != null && !isNaN(v) && v >= 0)
  if (clean.length < 2) return ''
  const w = 72, h = 32
  const min = Math.min(...clean), max = Math.max(...clean)
  const range = max - min || 1
  const pts = clean.map((v, i) => {
    const x = (i / (clean.length - 1)) * w
    const y = (h - 2) - ((v - min) / range) * (h - 6)
    return x.toFixed(1) + ',' + y.toFixed(1)
  }).join(' ')
  return `<svg width="${w}" height="${h}" style="position:absolute;bottom:6px;right:10px;pointer-events:none" xmlns="http://www.w3.org/2000/svg">`
    + `<polyline points="${pts}" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" opacity="0.35"/>`
    + `</svg>`
}
