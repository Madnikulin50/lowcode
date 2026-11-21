// Package riskstore is the persistence for the risk engine's domain
// objects: RiskFactorDef, RiskModel, RiskSubjectBinding, RiskAssessment,
// RiskTreatment.
//
// It is a separate package (rather than living inside server/compose/rest,
// where it started in phase 0) specifically so server/compose/service can
// depend on it too: the auto-recalc trigger (risk_trigger.go) needs to look
// up RiskSubjectBinding by module, and service already imports things rest
// imports — service importing rest would cycle.
//
// Storage: an in-memory cache (always present, so every read is a map
// lookup) backed by one Postgres table (risk_object, a JSONB blob per
// entity, keyed by kind+id) when DB_DSN points at Postgres — the same
// isolated-connection pattern server/pkg/objstore/dbblob already uses for a
// side subsystem, rather than the full store/rdbms codegen used for the
// platform's core entities. EnsurePersistence loads that table into the
// cache once at startup; every Save/Delete after that writes through
// synchronously. If DB_DSN isn't Postgres (or the connection fails),
// EnsurePersistence logs a warning and the package quietly falls back to
// cache-only behavior — the same in-memory-only mode phase 0 shipped with,
// so nothing else has to change to keep working.
package riskstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	_ "github.com/lib/pq"

	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/pkg/options"
	"github.com/madnikulin50/lowcode/server/pkg/riskengine"
)

const (
	tableName = "risk_object"

	kindFactor     = "factor"
	kindModel      = "model"
	kindBinding    = "binding"
	kindAssessment = "assessment"
	kindTreatment  = "treatment"
)

var (
	mu          sync.RWMutex
	factors     = map[uint64]*riskengine.RiskFactorDef{}
	models      = map[uint64]*riskengine.RiskModel{}
	bindings    = map[uint64]*riskengine.RiskSubjectBinding{}
	treatments  = map[uint64]*riskengine.RiskTreatment{}
	assessments = map[uint64]*riskengine.RiskAssessment{}

	dbOnce sync.Once
	db     *sql.DB
)

// NextID hands out a process-wide unique id for any of the risk entities,
// via the same generator (pkg/id, sonyflake) the rest of the platform uses
// for record/module/etc. ids — unlike a simple in-memory counter, ids it
// hands out never collide with ones a previous process run already
// persisted.
func NextID() uint64 {
	return id.Next()
}

// EnsurePersistence opens the Postgres connection and loads risk_object into
// the in-memory cache. Safe to call multiple times (only the first call
// does anything) and safe to call when DB_DSN isn't Postgres — logs and
// leaves the package in cache-only mode rather than returning an error, so
// callers (server/compose/mcp/bridge.go) can fire-and-forget it at startup.
func EnsurePersistence() {
	dbOnce.Do(func() {
		dsn := options.DB().DSN
		if !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
			log.Printf("[riskstore] DB_DSN is not Postgres — risk factors/models/assessments will not survive a restart")
			return
		}

		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("[riskstore] could not open connection: %v (falling back to in-memory only)", err)
			return
		}
		// Same rationale as dbblob: this is a side subsystem, keep its pool
		// small so it can't starve the main store's connections.
		conn.SetMaxOpenConns(4)

		ctx := context.Background()
		if err := ensureTable(ctx, conn); err != nil {
			log.Printf("[riskstore] could not ensure table: %v (falling back to in-memory only)", err)
			return
		}

		n, err := loadAll(ctx, conn)
		if err != nil {
			log.Printf("[riskstore] could not load from Postgres: %v (falling back to in-memory only)", err)
			return
		}

		db = conn
		log.Printf("[riskstore] loaded %d risk object(s) from Postgres", n)
	})
}

func ensureTable(ctx context.Context, conn *sql.DB) error {
	_, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS `+tableName+` (
		kind         text NOT NULL,
		id           bigint NOT NULL,
		namespace_id bigint NOT NULL DEFAULT 0,
		parent_id    bigint NOT NULL DEFAULT 0,
		secondary_id bigint NOT NULL DEFAULT 0,
		data         jsonb NOT NULL,
		created_at   timestamptz NOT NULL DEFAULT now(),
		updated_at   timestamptz,
		PRIMARY KEY (kind, id)
	)`)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS risk_object_namespace ON `+tableName+` (kind, namespace_id)`)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS risk_object_parent ON `+tableName+` (kind, parent_id)`)
	return err
}

func loadAll(ctx context.Context, conn *sql.DB) (int, error) {
	rows, err := conn.QueryContext(ctx, `SELECT kind, id, data FROM `+tableName)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	mu.Lock()
	defer mu.Unlock()

	n := 0
	for rows.Next() {
		var kind string
		var rowID uint64
		var data []byte
		if err := rows.Scan(&kind, &rowID, &data); err != nil {
			return n, err
		}
		switch kind {
		case kindFactor:
			var f riskengine.RiskFactorDef
			if err := json.Unmarshal(data, &f); err != nil {
				log.Printf("[riskstore] skip factor %d: %v", rowID, err)
				continue
			}
			factors[f.ID] = &f
		case kindModel:
			var m riskengine.RiskModel
			if err := json.Unmarshal(data, &m); err != nil {
				log.Printf("[riskstore] skip model %d: %v", rowID, err)
				continue
			}
			models[m.ID] = &m
		case kindBinding:
			var b riskengine.RiskSubjectBinding
			if err := json.Unmarshal(data, &b); err != nil {
				log.Printf("[riskstore] skip binding %d: %v", rowID, err)
				continue
			}
			bindings[b.ID] = &b
		case kindAssessment:
			var a riskengine.RiskAssessment
			if err := json.Unmarshal(data, &a); err != nil {
				log.Printf("[riskstore] skip assessment %d: %v", rowID, err)
				continue
			}
			assessments[a.ID] = &a
		case kindTreatment:
			var t riskengine.RiskTreatment
			if err := json.Unmarshal(data, &t); err != nil {
				log.Printf("[riskstore] skip treatment %d: %v", rowID, err)
				continue
			}
			treatments[t.ID] = &t
		default:
			continue
		}
		n++
	}
	return n, rows.Err()
}

// persist writes one entity through to Postgres. Best-effort: errors are
// logged, not returned — every Save*/Delete* function here already
// committed the in-memory change, which is the source of truth callers see
// immediately, and none of their signatures return an error today (matching
// the phase-0 in-memory-only API) — see the package doc comment.
func persist(kind string, objID, namespaceID, parentID, secondaryID uint64, v interface{}) {
	if db == nil {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("[riskstore] marshal %s %d: %v", kind, objID, err)
		return
	}
	_, err = db.Exec(fmt.Sprintf(`
		INSERT INTO %s (kind, id, namespace_id, parent_id, secondary_id, data, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (kind, id) DO UPDATE SET
			namespace_id = EXCLUDED.namespace_id,
			parent_id    = EXCLUDED.parent_id,
			secondary_id = EXCLUDED.secondary_id,
			data         = EXCLUDED.data,
			updated_at   = now()
	`, tableName), kind, objID, namespaceID, parentID, secondaryID, data)
	if err != nil {
		log.Printf("[riskstore] persist %s %d: %v", kind, objID, err)
	}
}

func remove(kind string, objID uint64) {
	if db == nil {
		return
	}
	if _, err := db.Exec(fmt.Sprintf(`DELETE FROM %s WHERE kind = $1 AND id = $2`, tableName), kind, objID); err != nil {
		log.Printf("[riskstore] delete %s %d: %v", kind, objID, err)
	}
}

// ---------------------------------------------------------------------------
// factors
// ---------------------------------------------------------------------------

func ListFactors(namespaceID uint64) []*riskengine.RiskFactorDef {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*riskengine.RiskFactorDef, 0, len(factors))
	for _, f := range factors {
		if namespaceID > 0 && f.NamespaceID != namespaceID {
			continue
		}
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func GetFactor(id uint64) (*riskengine.RiskFactorDef, bool) {
	mu.RLock()
	defer mu.RUnlock()
	f, ok := factors[id]
	return f, ok
}

// FactorsByID resolves a set of factor ids in one call, skipping any that
// don't exist — the shape Evaluate/Explain need.
func FactorsByID(ids []uint64) map[uint64]*riskengine.RiskFactorDef {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[uint64]*riskengine.RiskFactorDef, len(ids))
	for _, id := range ids {
		if f, ok := factors[id]; ok {
			out[id] = f
		}
	}
	return out
}

// FactorsForModel resolves every factor a model's nodes reference.
func FactorsForModel(model *riskengine.RiskModel) map[uint64]*riskengine.RiskFactorDef {
	ids := make([]uint64, 0, len(model.Nodes))
	for _, n := range model.Nodes {
		ids = append(ids, n.FactorID)
	}
	return FactorsByID(ids)
}

func SaveFactor(f *riskengine.RiskFactorDef) {
	mu.Lock()
	factors[f.ID] = f
	mu.Unlock()
	persist(kindFactor, f.ID, f.NamespaceID, 0, 0, f)
}

func DeleteFactor(id uint64) bool {
	mu.Lock()
	_, ok := factors[id]
	delete(factors, id)
	mu.Unlock()
	if ok {
		remove(kindFactor, id)
	}
	return ok
}

// ---------------------------------------------------------------------------
// models
// ---------------------------------------------------------------------------

func ListModels(namespaceID uint64) []*riskengine.RiskModel {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*riskengine.RiskModel, 0, len(models))
	for _, m := range models {
		if namespaceID > 0 && m.NamespaceID != namespaceID {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func GetModel(id uint64) (*riskengine.RiskModel, bool) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := models[id]
	return m, ok
}

func SaveModel(m *riskengine.RiskModel) {
	mu.Lock()
	models[m.ID] = m
	mu.Unlock()
	persist(kindModel, m.ID, m.NamespaceID, 0, 0, m)
}

func DeleteModel(id uint64) bool {
	mu.Lock()
	_, ok := models[id]
	delete(models, id)
	mu.Unlock()
	if ok {
		remove(kindModel, id)
	}
	return ok
}

// ---------------------------------------------------------------------------
// subject bindings
// ---------------------------------------------------------------------------

func ListBindings(namespaceID uint64) []*riskengine.RiskSubjectBinding {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*riskengine.RiskSubjectBinding, 0, len(bindings))
	for _, b := range bindings {
		if namespaceID > 0 && b.NamespaceID != namespaceID {
			continue
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func GetBinding(id uint64) (*riskengine.RiskSubjectBinding, bool) {
	mu.RLock()
	defer mu.RUnlock()
	b, ok := bindings[id]
	return b, ok
}

// BindingsForModule returns every binding (across namespaces) whose
// ModuleID matches — what the record-event trigger looks up per event.
func BindingsForModule(moduleID uint64) []*riskengine.RiskSubjectBinding {
	mu.RLock()
	defer mu.RUnlock()
	var out []*riskengine.RiskSubjectBinding
	for _, b := range bindings {
		if b.ModuleID == moduleID {
			out = append(out, b)
		}
	}
	return out
}

func SaveBinding(b *riskengine.RiskSubjectBinding) {
	mu.Lock()
	bindings[b.ID] = b
	mu.Unlock()
	persist(kindBinding, b.ID, b.NamespaceID, b.ModuleID, 0, b)
}

func DeleteBinding(id uint64) bool {
	mu.Lock()
	_, ok := bindings[id]
	delete(bindings, id)
	mu.Unlock()
	if ok {
		remove(kindBinding, id)
	}
	return ok
}

// ---------------------------------------------------------------------------
// assessments (stand-in for RiskSubjectBinding.AssessmentModuleID records
// until real Compose-record persistence is wired in — see the architecture
// note's "generic Risk Register poverh Compose" decision)
// ---------------------------------------------------------------------------

func SaveAssessment(a *riskengine.RiskAssessment) {
	if a.ID == 0 {
		a.ID = NextID()
	}
	mu.Lock()
	assessments[a.ID] = a
	mu.Unlock()
	persist(kindAssessment, a.ID, 0, a.BindingID, a.SubjectRecordID, a)
}

func GetAssessment(id uint64) (*riskengine.RiskAssessment, bool) {
	mu.RLock()
	defer mu.RUnlock()
	a, ok := assessments[id]
	return a, ok
}

// AssessmentsForBinding returns a binding's history, newest first. When
// subjectRecordID is non-zero it's filtered to that one subject.
func AssessmentsForBinding(bindingID, subjectRecordID uint64) []*riskengine.RiskAssessment {
	mu.RLock()
	defer mu.RUnlock()
	var out []*riskengine.RiskAssessment
	for _, a := range assessments {
		if a.BindingID != bindingID {
			continue
		}
		if subjectRecordID > 0 && a.SubjectRecordID != subjectRecordID {
			continue
		}
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AssessedAt.After(out[j].AssessedAt) })
	return out
}

// LatestPerSubject reduces a binding's history to each subject's most
// recent assessment — the shape a portfolio register/dashboard wants.
func LatestPerSubject(bindingID uint64) []*riskengine.RiskAssessment {
	all := AssessmentsForBinding(bindingID, 0)
	seen := make(map[uint64]bool, len(all))
	out := make([]*riskengine.RiskAssessment, 0, len(all))
	for _, a := range all { // already newest-first
		if seen[a.SubjectRecordID] {
			continue
		}
		seen[a.SubjectRecordID] = true
		out = append(out, a)
	}
	return out
}

// ---------------------------------------------------------------------------
// treatments
// ---------------------------------------------------------------------------

func ListTreatments(bindingID uint64) []*riskengine.RiskTreatment {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*riskengine.RiskTreatment, 0, len(treatments))
	for _, t := range treatments {
		if bindingID > 0 && t.BindingID != bindingID {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func GetTreatment(id uint64) (*riskengine.RiskTreatment, bool) {
	mu.RLock()
	defer mu.RUnlock()
	t, ok := treatments[id]
	return t, ok
}

func SaveTreatment(t *riskengine.RiskTreatment) {
	if t.ID == 0 {
		t.ID = NextID()
	}
	mu.Lock()
	treatments[t.ID] = t
	mu.Unlock()
	persist(kindTreatment, t.ID, 0, t.BindingID, t.SubjectRecordID, t)
}

func DeleteTreatment(id uint64) bool {
	mu.Lock()
	_, ok := treatments[id]
	delete(treatments, id)
	mu.Unlock()
	if ok {
		remove(kindTreatment, id)
	}
	return ok
}
