package agent

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/madnikulin50/lowcode/agents/sdk"
)

type Agent struct {
	cfg          Config
	scanner      *Scanner
	classifier   *Classifier
	enricher     *Enricher
	vulnScanner  *VulnScanner
	store        Storage
	mu           sync.RWMutex
	scans        map[string]*ScanStatus
	moduleID     uint64
	cidrs []string
	cb    *sdk.Callback
}

func New(cfg Config, store Storage) *Agent {
	cidrs := make([]string, len(cfg.AutoCIDRs))
	copy(cidrs, cfg.AutoCIDRs)
	return &Agent{
		cfg:          cfg,
		scanner:      NewScanner(cfg),
		classifier:   NewClassifier(cfg.LLMBaseURL, cfg.LLMModel),
		enricher:     NewEnricher(),
		vulnScanner:  NewVulnScanner(),
		store:        store,
		scans:        make(map[string]*ScanStatus),
		cidrs:        cidrs,
		cb:           sdk.NewCallback(),
	}
}

func (a *Agent) StartScan(ctx context.Context, target ScanTarget) (*ScanStatus, error) {
	id := uuid.New().String()
	status := &ScanStatus{
		ID: id, Target: target.CIDR, Status: "running",
		StartedAt: time.Now(),
	}
	nsID := uint64(target.NamespaceID)
	if nsID == 0 {
		nsID = a.cfg.NamespaceID
	}
	target.NamespaceID = FlexUint64(nsID)

	a.mu.Lock()
	a.scans[id] = status
	a.mu.Unlock()

	go a.runScan(context.Background(), id, target)

	return status, nil
}

func (a *Agent) runScan(ctx context.Context, id string, target ScanTarget) {
	cidr := target.CIDR
	nsID := uint64(target.NamespaceID)
	modID := uint64(target.ModuleID)
	stores := a.storesForScan(target)
	scanRecID := parseUint64(target.ScanRecordID)

	finish := func(status, errMsg string, found int, items []Device) {
		s := a.getStatus(id)
		if s == nil {
			return
		}
		s.Status = status
		s.Error = errMsg
		if errMsg != "" && s.Message == "" {
			s.Message = errMsg
		}
		s.Found = found
		s.Progress = 100
		now := time.Now()
		s.FinishedAt = &now
		a.setScanItems(id, items)
		a.syncScanRecord(ctx, stores, scanRecID, s)
		kind := "complete"
		if status == "error" {
			kind = "failed"
		}
		a.notifyCallback(target, id, kind, items)
	}

	resolved := ResolveScanCIDRs(cidr)
	if st := a.getStatus(id); st != nil {
		st.CIDRs = resolved
		if len(resolved) > 0 {
			st.Target = strings.Join(resolved, ", ")
		}
		a.syncScanRecord(ctx, stores, scanRecID, st)
	}
	log.Printf("scan %s: requested %q resolved to %v", id[:8], cidr, resolved)

	devices, err := a.scanner.Scan(ctx, cidr, func(current, total int, ip string) {
		s := a.getStatus(id)
		if s == nil {
			return
		}
		s.ScanningIP = ip
		s.ScannedIPs = current
		s.TotalIPs = total
		s.Progress = float64(current) / float64(total) * 40
		a.notifyCallback(target, id, "progress", nil)
	})
	if err != nil {
		finish("error", err.Error(), 0, nil)
		return
	}

	if len(devices) == 0 {
		msg := fmt.Sprintf("no live hosts on %s (TCP/ping/ARP); this machine's physical RFC1918 CIDRs: %s",
			strings.Join(resolved, ", "), strings.Join(LocalIPv4CIDRs(), ", "))
		if len(LocalIPv4CIDRs()) == 0 {
			msg = fmt.Sprintf("no live hosts on %s; no physical RFC1918 interface found to scan", strings.Join(resolved, ", "))
		}
		finish("done", msg, 0, nil)
		return
	}

	s := a.getStatus(id)
	s.Found = len(devices)
	s.Progress = 40
	s.ScanningIP = ""
	s.ScannedIPs = len(devices)
	a.syncScanRecord(ctx, stores, scanRecID, s)

	devices = a.enricher.Enrich(ctx, devices)
	for i := range devices {
		if devices[i].DeviceType == "" {
			devices[i].DeviceType = guessDeviceType(devices[i])
		}
	}

	if modID == 0 {
		mid, err := stores[0].EnsureModule(ctx)
		if err != nil {
			finish("error", fmt.Sprintf("module error: %v", err), len(devices), devices)
			return
		}
		modID = mid
		a.mu.Lock()
		a.moduleID = mid
		a.mu.Unlock()
	}

	status := a.getStatus(id)
	status.ModuleID = modID
	status.Progress = 50
	a.persistAll(ctx, stores, nsID, modID, devices)
	a.syncScanRecord(ctx, stores, scanRecID, status)
	log.Printf("scan %s: persisted %d devices", id[:8], len(devices))

	for i := range devices {
		devices[i] = a.classifier.Classify(ctx, devices[i])
		vulns := a.vulnScanner.Scan(ctx, &devices[i])
		if len(vulns) > 0 {
			devices[i].Vulnerabilities = vulns
		}
		s = a.getStatus(id)
		s.Progress = 50 + float64(i+1)/float64(len(devices))*40
	}
	a.persistAll(ctx, stores, nsID, modID, devices)

	s = a.getStatus(id)
	s.Progress = 100
	s.Status = "done"
	s.Found = len(devices)
	now := time.Now()
	s.FinishedAt = &now
	a.setScanItems(id, devices)
	a.syncScanRecord(ctx, stores, scanRecID, s)
	a.AddCIDR(cidr)
	a.notifyCallback(target, id, "complete", devices)
}

func (a *Agent) storesForScan(target ScanTarget) []Storage {
	stores := []Storage{a.store}
	if strings.TrimSpace(target.CallbackURL) != "" {
		// Ingest chain is the source of truth; SQLite stays a local cache.
		return stores
	}
	token := strings.TrimSpace(target.Token)
	if token == "" {
		token = strings.TrimSpace(a.cfg.Token)
	}
	if token == "" {
		return stores
	}
	api := strings.TrimRight(target.API, "/")
	if api == "" {
		api = strings.TrimRight(a.cfg.CortezaAPI, "/")
	}
	if api == "" {
		api = "http://localhost:3333/api"
	}
	cs := NewCortezaStore(api, token, uint64(target.NamespaceID))
	stores = append(stores, cs)
	return stores
}

func (a *Agent) persistAll(ctx context.Context, stores []Storage, nsID, modID uint64, devices []Device) {
	devices = DedupeDevices(devices)
	for _, st := range stores {
		mid := modID
		_, isCorteza := st.(*CortezaStore)
		if isCorteza {
			if resolved, err := st.EnsureModule(ctx); err == nil {
				mid = resolved
			} else if isUnauthorizedAPI(err) {
				log.Printf("corteza persist skipped: %v", err)
				continue
			}
		}
		authFailed := false
		for i, d := range devices {
			if authFailed {
				break
			}
			existingID, err := st.FindDevice(ctx, mid, d)
			if err == nil && existingID > 0 {
				if uErr := st.UpdateDevice(ctx, mid, existingID, d); uErr != nil {
					log.Printf("update device %s: %v", d.IP, uErr)
					if isUnauthorizedAPI(uErr) {
						authFailed = true
					}
				} else {
					devices[i].RecordID = existingID
				}
				continue
			}
			if err != nil && !isDeviceNotFound(err) {
				log.Printf("lookup device %s: %v (skip create to avoid duplicates)", d.IP, err)
				if isUnauthorizedAPI(err) {
					log.Printf("corteza persist stopped after unauthorized lookup")
					authFailed = true
				}
				continue
			}
			if id, cErr := st.CreateDevice(ctx, mid, d); cErr != nil {
				log.Printf("create device %s: %v", d.IP, cErr)
				if isUnauthorizedAPI(cErr) {
					authFailed = true
				}
			} else {
				devices[i].RecordID = id
			}
		}
		if cs, ok := st.(*CortezaStore); ok && !authFailed {
			cs.syncRelated(ctx, nsID, devices)
		}
	}
}

func (a *Agent) syncScanRecord(ctx context.Context, stores []Storage, recordID uint64, s *ScanStatus) {
	if s == nil || recordID == 0 {
		return
	}
	for _, st := range stores {
		cs, ok := st.(*CortezaStore)
		if !ok {
			continue
		}
		if err := cs.UpdateScan(ctx, recordID, s); err != nil {
			log.Printf("update scan record %d: %v", recordID, err)
		}
	}
}

func (a *Agent) GetStatus(id string) *ScanStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	s, ok := a.scans[id]
	if !ok {
		return nil
	}
	return s
}

func (a *Agent) ListScans() []ScanStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]ScanStatus, 0, len(a.scans))
	for _, s := range a.scans {
		result = append(result, *s)
	}
	return result
}

func (a *Agent) AddCIDR(cidr string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range a.cidrs {
		if c == cidr {
			return
		}
	}
	a.cidrs = append(a.cidrs, cidr)
}

func (a *Agent) StartPeriodicScan(ctx context.Context) {
	interval := a.cfg.ScanInterval
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.mu.RLock()
			cidrs := make([]string, len(a.cidrs))
			copy(cidrs, a.cidrs)
			a.mu.RUnlock()
			for _, cidr := range cidrs {
				id := uuid.New().String()
				s := &ScanStatus{
					ID: id, Target: cidr, Status: "running",
					StartedAt: time.Now(),
				}
				a.mu.Lock()
				a.scans[id] = s
				a.mu.Unlock()
				go a.runScan(context.Background(), id, ScanTarget{CIDR: cidr, NamespaceID: FlexUint64(a.cfg.NamespaceID)})
			}
		}
	}
}

func (a *Agent) checkDeviceLiveness(ctx context.Context, ip string, ports []Port) bool {
	// Try the first few previously open ports
	checkPorts := ports
	if len(checkPorts) > 5 {
		checkPorts = checkPorts[:5]
	}
	if len(checkPorts) == 0 {
		checkPorts = append(checkPorts, Port{Port: 80, Proto: "tcp"}, Port{Port: 22, Proto: "tcp"}, Port{Port: 443, Proto: "tcp"})
	}
	for _, p := range checkPorts {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", p.Port)), 2*time.Second)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

func (a *Agent) StartStatusChecker(ctx context.Context) {
	interval := a.cfg.StatusInterval
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			devices, err := a.store.ListDevices(ctx, 0)
			if err != nil {
				log.Printf("status check: list devices: %v", err)
				continue
			}
			for i := range devices {
				select {
				case <-ctx.Done():
					return
				default:
				}
				alive := a.checkDeviceLiveness(ctx, devices[i].IP, devices[i].OpenPorts)
				if alive {
					if devices[i].Status != "online" {
						devices[i].Status = "online"
					}
					devices[i].LastSeen = time.Now().Format(time.RFC3339)
				} else {
					devices[i].Status = "offline"
				}
				if err := a.store.UpdateDevice(ctx, 0, devices[i].RecordID, devices[i]); err != nil {
					log.Printf("status check: update %s: %v", devices[i].IP, err)
				}
			}
			log.Printf("status check: updated %d devices", len(devices))
		}
	}
}

func parseUint64(s string) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	var n uint64
	fmt.Sscanf(s, "%d", &n)
	return n
}

func (a *Agent) getStatus(id string) *ScanStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.scans[id]
}

func (a *Agent) setScanItems(id string, items []Device) {
	s := a.getStatus(id)
	if s == nil {
		return
	}
	cp := make([]Device, len(items))
	copy(cp, items)
	s.Items = cp
}

func (a *Agent) ListDevices(ctx context.Context, modID uint64) ([]Device, error) {
	if modID == 0 {
		a.mu.RLock()
		modID = a.moduleID
		a.mu.RUnlock()
	}
	if modID == 0 {
		mid, err := a.store.EnsureModule(ctx)
		if err != nil {
			return nil, fmt.Errorf("no devices module found: %w", err)
		}
		modID = mid
		a.mu.Lock()
		a.moduleID = mid
		a.mu.Unlock()
	}
	return a.store.ListDevices(ctx, modID)
}

func (a *Agent) GetDevice(ctx context.Context, modID, recordID uint64) (*Device, error) {
	if modID == 0 {
		a.mu.RLock()
		modID = a.moduleID
		a.mu.RUnlock()
	}
	return a.store.GetDevice(ctx, modID, recordID)
}

func (a *Agent) DeleteDevice(ctx context.Context, modID, recordID uint64) error {
	if modID == 0 {
		a.mu.RLock()
		modID = a.moduleID
		a.mu.RUnlock()
	}
	return a.store.DeleteDevice(ctx, modID, recordID)
}

func (a *Agent) EnsureModule(ctx context.Context) (uint64, error) {
	mid, err := a.store.EnsureModule(ctx)
	if err != nil {
		return 0, err
	}
	a.mu.Lock()
	a.moduleID = mid
	a.mu.Unlock()
	return mid, nil
}
