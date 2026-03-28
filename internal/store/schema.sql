CREATE TABLE IF NOT EXISTS measurements (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	ts             TEXT    NOT NULL,
	network_id     TEXT    NOT NULL DEFAULT '',
	latency        REAL    NOT NULL DEFAULT 0,
	jitter         REAL    NOT NULL DEFAULT 0,
	packet_loss    REAL    NOT NULL DEFAULT 0,
	download       REAL    NOT NULL DEFAULT 0,
	upload         REAL    NOT NULL DEFAULT 0,
	dns            REAL    NOT NULL DEFAULT 0,
	conn_type      TEXT    NOT NULL DEFAULT '',
	conn_rssi      INTEGER NOT NULL DEFAULT 0,
	conn_noise     INTEGER NOT NULL DEFAULT 0,
	conn_snr       INTEGER NOT NULL DEFAULT 0,
	conn_channel   INTEGER NOT NULL DEFAULT 0,
	conn_band      TEXT    NOT NULL DEFAULT '',
	conn_link_rate INTEGER NOT NULL DEFAULT 0,
	conn_duplex    TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS ping_targets (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	host       TEXT NOT NULL UNIQUE,
	ip         TEXT NOT NULL DEFAULT '',
	latency    REAL NOT NULL DEFAULT 0,
	loss       REAL NOT NULL DEFAULT 0,
	status     TEXT NOT NULL DEFAULT 'unknown',
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS dns_checks (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	host       TEXT NOT NULL UNIQUE,
	time_ms    REAL NOT NULL DEFAULT 0,
	resolver   TEXT NOT NULL DEFAULT 'system',
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS network_events (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	ts         TEXT    NOT NULL,
	network_id TEXT    NOT NULL DEFAULT '',
	ssid       TEXT    NOT NULL DEFAULT '',
	gateway    TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS settings (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_measurements_ts    ON measurements(ts);
CREATE INDEX IF NOT EXISTS idx_network_events_ts  ON network_events(ts);
