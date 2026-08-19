package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type EmbeddedStore struct {
	db *sql.DB
}

func NewEmbeddedStore(dbPath string) (*EmbeddedStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open embedded db: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip TEXT NOT NULL,
			mac TEXT DEFAULT '',
			hostname TEXT DEFAULT '',
			vendor TEXT DEFAULT '',
			device_type TEXT DEFAULT 'unknown',
			os TEXT DEFAULT '',
			open_ports TEXT DEFAULT '[]',
			services TEXT DEFAULT '[]',
			shares TEXT DEFAULT '[]',
			vulnerabilities TEXT DEFAULT '[]',
			last_seen TEXT DEFAULT '',
			status TEXT DEFAULT 'unknown',
			created_at TEXT DEFAULT (datetime('now')),
			updated_at TEXT DEFAULT (datetime('now'))
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_ip ON devices(ip);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot init embedded db schema: %w", err)
	}

	// migrate columns for existing databases
	for _, col := range []string{
		"domain TEXT DEFAULT ''",
		"shares TEXT DEFAULT '[]'",
		"vulnerabilities TEXT DEFAULT '[]'",
	} {
		db.Exec("ALTER TABLE devices ADD COLUMN " + col) // ignore "duplicate column" errors
	}

	return &EmbeddedStore{db: db}, nil
}

func (s *EmbeddedStore) EnsureModule(ctx context.Context) (uint64, error) {
	return 1, nil
}

func (s *EmbeddedStore) FindDevice(ctx context.Context, _ uint64, d Device) (uint64, error) {
	if mac := normalizeMAC(d.MAC); mac != "" {
		var id uint64
		err := s.db.QueryRowContext(ctx, `
			SELECT id FROM devices
			WHERE lower(replace(replace(mac, '-', ':'), '.', ':')) = ?
			LIMIT 1`, mac).Scan(&id)
		if err == nil && id > 0 {
			return id, nil
		}
		if err != nil && err != sql.ErrNoRows {
			return 0, err
		}
	}
	if ip := normalizeIP(d.IP); ip != "" {
		id, err := s.FindDeviceByIP(ctx, 0, ip)
		if err == nil {
			return id, nil
		}
		if !isDeviceNotFound(err) {
			return 0, err
		}
	}
	if host := stableHostname(d.Hostname); host != "" {
		var id uint64
		err := s.db.QueryRowContext(ctx, `
			SELECT id FROM devices WHERE lower(hostname) = ? LIMIT 1`, host).Scan(&id)
		if err == nil && id > 0 {
			return id, nil
		}
		if err != nil && err != sql.ErrNoRows {
			return 0, err
		}
	}
	return 0, errDeviceNotFound
}

func (s *EmbeddedStore) FindDeviceByIP(ctx context.Context, _ uint64, ip string) (uint64, error) {
	var id uint64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM devices WHERE ip = ?", ip).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errDeviceNotFound
		}
		return 0, err
	}
	return id, nil
}

func (s *EmbeddedStore) CreateDevice(ctx context.Context, _ uint64, d Device) (uint64, error) {
	portsJSON, _ := json.Marshal(d.OpenPorts)
	svcJSON, _ := json.Marshal(d.Services)
	sharesJSON, _ := json.Marshal(d.Shares)
	vulnsJSON, _ := json.Marshal(d.Vulnerabilities)
	now := time.Now().Format(time.RFC3339)

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO devices (ip, mac, hostname, vendor, device_type, os, domain, open_ports, services, shares, vulnerabilities, last_seen, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.IP, d.MAC, d.Hostname, d.Vendor, d.DeviceType, d.OS, d.Domain,
		string(portsJSON), string(svcJSON), string(sharesJSON), string(vulnsJSON), d.LastSeen, d.Status, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("cannot create device: %w", err)
	}
	id, _ := res.LastInsertId()
	return uint64(id), nil
}

func (s *EmbeddedStore) UpdateDevice(ctx context.Context, _, recordID uint64, d Device) error {
	portsJSON, _ := json.Marshal(d.OpenPorts)
	svcJSON, _ := json.Marshal(d.Services)
	sharesJSON, _ := json.Marshal(d.Shares)
	vulnsJSON, _ := json.Marshal(d.Vulnerabilities)
	now := time.Now().Format(time.RFC3339)

	_, err := s.db.ExecContext(ctx, `
		UPDATE devices SET ip=?, mac=?, hostname=?, vendor=?, device_type=?, os=?, domain=?,
			open_ports=?, services=?, shares=?, vulnerabilities=?, last_seen=?, status=?, updated_at=?
		WHERE id=?`,
		d.IP, d.MAC, d.Hostname, d.Vendor, d.DeviceType, d.OS, d.Domain,
		string(portsJSON), string(svcJSON), string(sharesJSON), string(vulnsJSON), d.LastSeen, d.Status, now, recordID,
	)
	return err
}

func (s *EmbeddedStore) ListDevices(ctx context.Context, _ uint64) ([]Device, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, ip, mac, hostname, vendor, device_type, os, domain, open_ports, services, shares, vulnerabilities, last_seen, status
		FROM devices ORDER BY ip`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		var mac, hostname, vendor, deviceType, os, domain, portsStr, svcStr, sharesStr, vulnsStr, lastSeen, status sql.NullString
		err := rows.Scan(&d.RecordID, &d.IP, &mac, &hostname, &vendor, &deviceType, &os, &domain, &portsStr, &svcStr, &sharesStr, &vulnsStr, &lastSeen, &status)
		if err != nil {
			return nil, err
		}
		d.MAC = mac.String
		d.Hostname = hostname.String
		d.Vendor = vendor.String
		d.DeviceType = deviceType.String
		d.OS = os.String
		d.Domain = domain.String
		d.LastSeen = lastSeen.String
		d.Status = status.String
		if d.Status == "" {
			d.Status = "unknown"
		}
		if portsStr.Valid && portsStr.String != "" {
			json.Unmarshal([]byte(portsStr.String), &d.OpenPorts)
		}
		if svcStr.Valid && svcStr.String != "" {
			json.Unmarshal([]byte(svcStr.String), &d.Services)
		}
		if sharesStr.Valid && sharesStr.String != "" {
			json.Unmarshal([]byte(sharesStr.String), &d.Shares)
		}
		if vulnsStr.Valid && vulnsStr.String != "" {
			json.Unmarshal([]byte(vulnsStr.String), &d.Vulnerabilities)
		}
		devices = append(devices, d)
	}
	if devices == nil {
		devices = []Device{}
	}
	return devices, rows.Err()
}

func (s *EmbeddedStore) GetDevice(ctx context.Context, _, recordID uint64) (*Device, error) {
	var d Device
	var mac, hostname, vendor, deviceType, os, domain, portsStr, svcStr, sharesStr, vulnsStr, lastSeen, status sql.NullString

	err := s.db.QueryRowContext(ctx, `
		SELECT id, ip, mac, hostname, vendor, device_type, os, domain, open_ports, services, shares, vulnerabilities, last_seen, status
		FROM devices WHERE id=?`, recordID).Scan(
		&d.RecordID, &d.IP, &mac, &hostname, &vendor, &deviceType, &os, &domain, &portsStr, &svcStr, &sharesStr, &vulnsStr, &lastSeen, &status,
	)
	if err != nil {
		return nil, fmt.Errorf("device not found")
	}
	d.MAC = mac.String
	d.Hostname = hostname.String
	d.Vendor = vendor.String
	d.DeviceType = deviceType.String
	d.OS = os.String
	d.Domain = domain.String
	d.LastSeen = lastSeen.String
	d.Status = status.String
	if d.Status == "" {
		d.Status = "unknown"
	}
	if portsStr.Valid && portsStr.String != "" {
		json.Unmarshal([]byte(portsStr.String), &d.OpenPorts)
	}
	if svcStr.Valid && svcStr.String != "" {
		json.Unmarshal([]byte(svcStr.String), &d.Services)
	}
	if sharesStr.Valid && sharesStr.String != "" {
		json.Unmarshal([]byte(sharesStr.String), &d.Shares)
	}
	if vulnsStr.Valid && vulnsStr.String != "" {
		json.Unmarshal([]byte(vulnsStr.String), &d.Vulnerabilities)
	}
	return &d, nil
}

func (s *EmbeddedStore) DeleteDevice(ctx context.Context, _, recordID uint64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM devices WHERE id=?", recordID)
	return err
}

func (s *EmbeddedStore) Close() error {
	return s.db.Close()
}
