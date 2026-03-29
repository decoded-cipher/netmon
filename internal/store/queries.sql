-- name: insert_measurement
INSERT INTO measurements
	(ts, network_id, latency, jitter, packet_loss, download, upload, dns,
	 conn_type, conn_rssi, conn_noise, conn_snr, conn_channel, conn_band, conn_link_rate, conn_duplex)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: get_history
SELECT id, ts, network_id, latency, jitter, packet_loss, download, upload, dns,
       conn_type, conn_rssi, conn_noise, conn_snr, conn_channel, conn_band, conn_link_rate, conn_duplex
FROM measurements ORDER BY ts DESC LIMIT ?;

-- name: get_history_window
SELECT id, ts, network_id, latency, jitter, packet_loss, download, upload, dns,
       conn_type, conn_rssi, conn_noise, conn_snr, conn_channel, conn_band, conn_link_rate, conn_duplex
FROM measurements WHERE ts >= ? ORDER BY ts ASC;

-- name: get_summary_stats
SELECT
	COALESCE(AVG(latency),    0),
	COALESCE(MIN(latency),    0),
	COALESCE(MAX(latency),    0),
	COALESCE(AVG(jitter),     0),
	COALESCE(AVG(packet_loss),0),
	COALESCE(AVG(dns),        0),
	COALESCE(AVG(download),   0),
	COALESCE(AVG(upload),     0)
FROM measurements WHERE ts >= ?;

-- name: get_latencies
SELECT latency FROM measurements WHERE ts >= ? ORDER BY latency;

-- name: count_measurements
SELECT COUNT(*) FROM measurements WHERE ts >= ?;

-- name: count_up_measurements
SELECT COUNT(*) FROM measurements WHERE ts >= ? AND packet_loss < 100;

-- name: count_outages
SELECT COUNT(*) FROM measurements WHERE ts >= ? AND packet_loss >= 50;

-- name: upsert_ping_target
INSERT INTO ping_targets (host, ip, latency, loss, status, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(host) DO UPDATE SET
	ip         = excluded.ip,
	latency    = excluded.latency,
	loss       = excluded.loss,
	status     = excluded.status,
	updated_at = excluded.updated_at;

-- name: get_ping_targets
SELECT host, ip, latency, loss, status FROM ping_targets ORDER BY host;

-- name: upsert_dns_check
INSERT INTO dns_checks (host, time_ms, resolver, updated_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(host) DO UPDATE SET
	time_ms    = excluded.time_ms,
	resolver   = excluded.resolver,
	updated_at = excluded.updated_at;

-- name: get_dns_checks
SELECT host, time_ms, resolver FROM dns_checks ORDER BY host;

-- name: log_network_event
INSERT INTO network_events (ts, network_id, ssid, gateway) VALUES (?, ?, ?, ?);

-- name: get_current_network_id
SELECT network_id FROM network_events ORDER BY ts DESC LIMIT 1;

-- name: get_config
SELECT value FROM settings WHERE key = 'config';

-- name: save_config
INSERT INTO settings (key, value) VALUES ('config', ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value;
