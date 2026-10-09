package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
)

// IncidentRouteResult is returned after assigning a routing row + staff member.
type IncidentRouteResult struct {
	IncidentID   uint64 `json:"incidentID" db:"incident_id"`
	Status       string `json:"status" db:"status"`
	PersonID     string `json:"personID" db:"person_id"`
	DepartmentID int64  `json:"departmentID" db:"department_id"`
	Position     string `json:"position" db:"position"`
	SLAHours     int    `json:"slaHours" db:"sla_hours"`
	RoutingFound bool   `json:"routingFound"`
	Assigned     bool   `json:"assigned"`
	Message      string `json:"message"`
}

// IncidentEscalateResult is a batch SLA sweep summary.
type IncidentEscalateResult struct {
	Escalated int64  `json:"escalated" db:"escalated"`
	Message   string `json:"message"`
}

func incidentDB(ctx context.Context) (sqlx.ExtContext, error) {
	rs := rdbmsStore(DefaultStore)
	if rs == nil || rs.DB == nil {
		return nil, errors.Internal("store type not supported")
	}
	return rs.DB, nil
}

// RouteIncident looks up incident_routing by type+criticality and assigns
// a retail_staff person (same store preferred) to the incident.
func RouteIncident(ctx context.Context, recordID uint64) (*IncidentRouteResult, error) {
	db, err := incidentDB(ctx)
	if err != nil {
		return nil, err
	}
	if recordID == 0 {
		return nil, fmt.Errorf("recordID is required")
	}

	var out IncidentRouteResult
	err = sqlx.GetContext(ctx, db, &out, `
WITH inc AS (
  SELECT i.id, i.store_id, i.incident_type, i.criticality, i.incident_status, i.person_id
    FROM incidents i
   WHERE i.id = $1
),
rt AS (
  SELECT r.department_id, r.position, r.time
    FROM incident_routing r
    JOIN inc ON r.incident_type = inc.incident_type
            AND r.criticality = inc.criticality
   LIMIT 1
),
staff AS (
  SELECT s.person_id
    FROM retail_staff s
    JOIN rt ON s.department_id = rt.department_id
           AND (
             s.position ILIKE '%' || rt.position || '%'
             OR split_part(s.position, ' - ', 1) ILIKE rt.position
           )
    JOIN inc ON TRUE
   ORDER BY CASE WHEN s.store_id = inc.store_id THEN 0 ELSE 1 END, s.person_id
   LIMIT 1
),
upd AS (
  UPDATE incidents i
     SET person_id = COALESCE('person_' || (staff.person_id::bigint)::text, i.person_id)
    FROM inc, staff
   WHERE i.id = inc.id
     AND inc.incident_status NOT IN ('Closed', 'Resolved')
  RETURNING i.id, i.incident_status, i.person_id
)
SELECT
  COALESCE(upd.id, inc.id) AS incident_id,
  COALESCE(upd.incident_status, inc.incident_status) AS status,
  COALESCE(upd.person_id, inc.person_id, '') AS person_id,
  COALESCE(rt.department_id, 0)::bigint AS department_id,
  COALESCE(rt.position, '') AS position,
  COALESCE(rt.time, 0)::int AS sla_hours
FROM inc
LEFT JOIN rt ON TRUE
LEFT JOIN upd ON TRUE
`, recordID)
	if err != nil {
		return nil, fmt.Errorf("route incident: %w", scanIncident(err, recordID))
	}

	out.RoutingFound = out.DepartmentID > 0 || out.Position != ""
	out.Assigned = strings.HasPrefix(out.PersonID, "person_")
	switch {
	case !out.RoutingFound:
		out.Message = "Нет строки маршрутизации для типа/критичности"
	case !out.Assigned:
		out.Message = "Маршрут найден, но нет сотрудника в отделе"
	default:
		out.Message = fmt.Sprintf("Назначен %s (%s, SLA %d ч)", out.PersonID, out.Position, out.SLAHours)
	}
	return &out, nil
}

// SetIncidentStatus sets incident_status (In Progress / Resolved / Escalated / …).
func SetIncidentStatus(ctx context.Context, recordID uint64, status string) (*IncidentRouteResult, error) {
	db, err := incidentDB(ctx)
	if err != nil {
		return nil, err
	}
	if recordID == 0 {
		return nil, fmt.Errorf("recordID is required")
	}
	status = strings.TrimSpace(status)
	switch status {
	case "Open", "In Progress", "Escalated", "Resolved", "Closed":
	default:
		return nil, fmt.Errorf("unsupported status %q", status)
	}

	var out IncidentRouteResult
	err = sqlx.GetContext(ctx, db, &out, `
UPDATE incidents
   SET incident_status = $2
 WHERE id = $1
RETURNING id AS incident_id, incident_status AS status, COALESCE(person_id, '') AS person_id,
          0::bigint AS department_id, '' AS position, 0 AS sla_hours
`, recordID, status)
	if err != nil {
		return nil, fmt.Errorf("set incident status: %w", scanIncident(err, recordID))
	}
	out.Message = "Статус: " + out.Status
	return &out, nil
}

// EscalateIncident forces Escalated and reassigns to escalation dept/position.
func EscalateIncident(ctx context.Context, recordID uint64) (*IncidentRouteResult, error) {
	db, err := incidentDB(ctx)
	if err != nil {
		return nil, err
	}
	if recordID == 0 {
		return nil, fmt.Errorf("recordID is required")
	}

	var out IncidentRouteResult
	err = sqlx.GetContext(ctx, db, &out, `
WITH inc AS (
  SELECT i.id, i.store_id, i.incident_type, i.criticality
    FROM incidents i WHERE i.id = $1
),
rt AS (
  SELECT r.escalation_department_id AS department_id, r.escalation_position AS position, r.time
    FROM incident_routing r
    JOIN inc ON r.incident_type = inc.incident_type AND r.criticality = inc.criticality
   LIMIT 1
),
staff AS (
  SELECT s.person_id
    FROM retail_staff s
    JOIN rt ON s.department_id = rt.department_id
           AND (
             s.position ILIKE '%' || rt.position || '%'
             OR split_part(s.position, ' - ', 1) ILIKE rt.position
           )
    JOIN inc ON TRUE
   ORDER BY CASE WHEN s.store_id = inc.store_id THEN 0 ELSE 1 END, s.person_id
   LIMIT 1
),
upd AS (
  UPDATE incidents i
     SET incident_status = 'Escalated',
         person_id = COALESCE('person_' || (staff.person_id::bigint)::text, i.person_id)
    FROM inc
    LEFT JOIN staff ON TRUE
   WHERE i.id = inc.id
  RETURNING i.id, i.incident_status, i.person_id
)
SELECT
  upd.id AS incident_id,
  upd.incident_status AS status,
  COALESCE(upd.person_id, '') AS person_id,
  COALESCE(rt.department_id, 0)::bigint AS department_id,
  COALESCE(rt.position, '') AS position,
  COALESCE(rt.time, 0)::int AS sla_hours
FROM upd
LEFT JOIN rt ON TRUE
`, recordID)
	if err != nil {
		return nil, fmt.Errorf("escalate incident: %w", scanIncident(err, recordID))
	}
	out.Assigned = strings.HasPrefix(out.PersonID, "person_")
	out.RoutingFound = out.DepartmentID > 0 || out.Position != ""
	out.Message = "Эскалация: " + out.PersonID
	return &out, nil
}

// EscalateOverdueIncidents promotes Open/In Progress past SLA (hours in incident_routing.time).
// Only rows with dt in the last 21 days are touched, so frozen demo history is left alone.
func EscalateOverdueIncidents(ctx context.Context) (*IncidentEscalateResult, error) {
	db, err := incidentDB(ctx)
	if err != nil {
		return nil, err
	}

	var n int64
	err = sqlx.GetContext(ctx, db, &n, `
WITH due AS (
  SELECT i.id, i.store_id, r.escalation_department_id, r.escalation_position
    FROM incidents i
    JOIN incident_routing r
      ON r.incident_type = i.incident_type AND r.criticality = i.criticality
   WHERE i.incident_status IN ('Open', 'In Progress')
     AND i.dt >= CURRENT_TIMESTAMP - INTERVAL '21 days'
     AND i.dt + make_interval(hours => COALESCE(r.time, 0)::int) < CURRENT_TIMESTAMP
   LIMIT 100
),
staff AS (
  SELECT DISTINCT ON (d.id) d.id AS incident_id, s.person_id
    FROM due d
    JOIN retail_staff s
      ON s.department_id = d.escalation_department_id
     AND (
       s.position ILIKE '%' || d.escalation_position || '%'
       OR split_part(s.position, ' - ', 1) ILIKE d.escalation_position
     )
   ORDER BY d.id, CASE WHEN s.store_id = d.store_id THEN 0 ELSE 1 END, s.person_id
),
upd AS (
  UPDATE incidents i
     SET incident_status = 'Escalated',
         person_id = COALESCE('person_' || (staff.person_id::bigint)::text, i.person_id)
    FROM due
    LEFT JOIN staff ON staff.incident_id = due.id
   WHERE i.id = due.id
  RETURNING i.id
)
SELECT COUNT(*)::bigint FROM upd
`)
	if err != nil {
		return nil, fmt.Errorf("escalate overdue: %w", err)
	}
	return &IncidentEscalateResult{
		Escalated: n,
		Message:   fmt.Sprintf("Эскалировано по SLA: %d", n),
	}, nil
}

func scanIncident(err error, recordID uint64) error {
	if err == sql.ErrNoRows {
		return fmt.Errorf("incident %d not found", recordID)
	}
	return err
}
func ParseRecordID(v interface{}) uint64 {
	switch t := v.(type) {
	case uint64:
		return t
	case int64:
		if t > 0 {
			return uint64(t)
		}
	case int:
		if t > 0 {
			return uint64(t)
		}
	case float64:
		if t > 0 {
			return uint64(t)
		}
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0
		}
		n, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			return n
		}
		f, err := strconv.ParseFloat(s, 64)
		if err == nil && f > 0 {
			return uint64(f)
		}
	}
	return 0
}
