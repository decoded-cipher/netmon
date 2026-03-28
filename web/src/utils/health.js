export function computeHealth(s) {
  let score = 100
  if (s.latency_avg > 50)       score -= 20
  else if (s.latency_avg > 30)  score -= 10
  else if (s.latency_avg > 20)  score -= 5
  if (s.packet_loss > 5)        score -= 30
  else if (s.packet_loss > 1)   score -= 15
  else if (s.packet_loss > 0.5) score -= 5
  if (s.download_avg < 25)      score -= 20
  else if (s.download_avg < 50) score -= 10
  if (s.jitter_avg > 10)        score -= 10
  else if (s.jitter_avg > 5)    score -= 5
  return Math.max(0, Math.min(100, score))
}

export function rateMetric(val, good, warn) {
  if (val <= good) return { text: 'Good', color: 'var(--green)' }
  if (val <= warn) return { text: 'Fair', color: 'var(--yellow)' }
  return { text: 'Poor', color: 'var(--red)' }
}
