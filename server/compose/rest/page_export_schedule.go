package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	jsonxt "github.com/jmoiron/sqlx/types"
	gomail "gopkg.in/mail.v2"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/mail"
	"github.com/madnikulin50/lowcode/server/store"
	systemTypes "github.com/madnikulin50/lowcode/server/system/types"
)

const (
	pageExportScheduleKeyPrefix = "compose.page-export-schedule."
	pageExportScheduleTick      = 30 * time.Second
)

type (
	// pageExportSchedule describes periodic sending of the page export by email
	pageExportSchedule struct {
		PageID      uint64 `json:"pageID,string"`
		NamespaceID uint64 `json:"namespaceID,string"`

		Enabled    bool     `json:"enabled"`
		Recipients []string `json:"recipients"`
		Subject    string   `json:"subject"`
		Body       string   `json:"body"`
		Format     string   `json:"format"` // xlsx | pdf | pptx

		Frequency string `json:"frequency"` // daily | weekly | monthly
		Time      string `json:"time"`      // HH:MM, server local time
		Weekday   int    `json:"weekday"`   // 1 (monday) .. 7 (sunday), for weekly
		MonthDay  int    `json:"monthDay"`  // 1 .. 31, for monthly (clamped to the last day of month)

		// address of the page in the webapp
		URL string `json:"url"`

		// user on whose behalf the page is rendered (the one who saved the schedule)
		OwnerID uint64 `json:"ownerID,string"`

		LastRunAt *time.Time `json:"lastRunAt,omitempty"`
		LastError string     `json:"lastError,omitempty"`
	}
)

func pageExportScheduleKey(pageID uint64) string {
	return pageExportScheduleKeyPrefix + strconv.FormatUint(pageID, 10)
}

func (s *pageExportSchedule) validate() error {
	s.Format = strings.ToLower(s.Format)
	if s.Format != "pdf" && s.Format != "xlsx" && s.Format != "pptx" {
		return fmt.Errorf("unsupported format")
	}

	rr := make([]string, 0, len(s.Recipients))
	for _, r := range s.Recipients {
		if r = strings.TrimSpace(r); r == "" {
			continue
		}
		if !mail.IsValidAddress(r) {
			return fmt.Errorf("invalid recipient: %s", r)
		}
		rr = append(rr, r)
	}
	s.Recipients = rr

	if s.Enabled && len(s.Recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	if _, err := time.Parse("15:04", s.Time); err != nil {
		return fmt.Errorf("time must be in HH:MM format")
	}

	switch s.Frequency {
	case "daily":
	case "weekly":
		if s.Weekday < 1 || s.Weekday > 7 {
			return fmt.Errorf("weekday must be between 1 and 7")
		}
	case "monthly":
		if s.MonthDay < 1 || s.MonthDay > 31 {
			return fmt.Errorf("month day must be between 1 and 31")
		}
	default:
		return fmt.Errorf("unsupported frequency")
	}

	return nil
}

// due tells if the schedule has to fire at the given minute
func (s pageExportSchedule) due(now time.Time) bool {
	if !s.Enabled || now.Format("15:04") != s.Time {
		return false
	}

	switch s.Frequency {
	case "daily":
		return true
	case "weekly":
		wd := int(now.Weekday())
		if wd == 0 {
			wd = 7
		}
		return wd == s.Weekday
	case "monthly":
		last := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
		day := s.MonthDay
		if day > last {
			day = last
		}
		return now.Day() == day
	}

	return false
}

func loadPageExportSchedule(ctx context.Context, pageID uint64) (*pageExportSchedule, error) {
	v, err := store.LookupSettingValueByNameOwnedBy(ctx, service.DefaultStore, pageExportScheduleKey(pageID), 0)
	if err != nil || v == nil {
		return nil, nil
	}

	s := &pageExportSchedule{}
	if err = json.Unmarshal(v.Value, s); err != nil {
		return nil, err
	}
	return s, nil
}

func savePageExportSchedule(ctx context.Context, s *pageExportSchedule) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return store.UpsertSettingValue(ctx, service.DefaultStore, &systemTypes.SettingValue{
		Name:      pageExportScheduleKey(s.PageID),
		Value:     jsonxt.JSONText(raw),
		UpdatedAt: time.Now(),
		UpdatedBy: s.OwnerID,
	})
}

func defaultPageExportSchedule(namespaceID, pageID uint64) *pageExportSchedule {
	return &pageExportSchedule{
		PageID:      pageID,
		NamespaceID: namespaceID,
		Recipients:  []string{},
		Format:      "xlsx",
		Frequency:   "daily",
		Time:        "09:00",
		Weekday:     1,
		MonthDay:    1,
	}
}

// schedulePage checks that the current user is allowed to manage the page schedule
func schedulePage(r *http.Request) (namespaceID, pageID uint64, err error) {
	ctx := r.Context()

	if namespaceID, err = strconv.ParseUint(chi.URLParam(r, "namespaceID"), 10, 64); err != nil {
		return 0, 0, fmt.Errorf("invalid namespace ID")
	}
	if pageID, err = strconv.ParseUint(chi.URLParam(r, "pageID"), 10, 64); err != nil {
		return 0, 0, fmt.Errorf("invalid page ID")
	}

	p, err := service.DefaultPage.FindByID(ctx, namespaceID, pageID)
	if err != nil {
		return 0, 0, err
	}

	if !service.DefaultAccessControl.CanUpdatePage(ctx, p) {
		return 0, 0, fmt.Errorf("not allowed to update the page")
	}

	return namespaceID, pageID, nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"response": v})
}

// ScheduleGet returns the schedule of the page (defaults when it was never saved)
func (PageExport) ScheduleGet(w http.ResponseWriter, r *http.Request) {
	namespaceID, pageID, err := schedulePage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	s, err := loadPageExportSchedule(r.Context(), pageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s == nil {
		s = defaultPageExportSchedule(namespaceID, pageID)
	}

	writeJSON(w, s)
}

// ScheduleSet validates and stores the schedule of the page
func (PageExport) ScheduleSet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	namespaceID, pageID, err := schedulePage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	in := &pageExportSchedule{}
	if err = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(in); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err = in.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = pageExportCheckURL(in.URL, requestHost(r)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	prev, _ := loadPageExportSchedule(ctx, pageID)

	in.PageID, in.NamespaceID = pageID, namespaceID
	in.OwnerID = auth.GetIdentityFromContext(ctx).Identity()
	if prev != nil {
		in.LastRunAt, in.LastError = prev.LastRunAt, prev.LastError
	}

	if err = savePageExportSchedule(ctx, in); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, in)
}

// ScheduleSendNow sends the stored schedule immediately (to check the settings)
func (PageExport) ScheduleSendNow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, pageID, err := schedulePage(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	s, err := loadPageExportSchedule(ctx, pageID)
	if err != nil || s == nil {
		http.Error(w, "schedule is not saved yet", http.StatusBadRequest)
		return
	}

	if err = runPageExportSchedule(context.WithoutCancel(ctx), s); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, s)
}

// runPageExportSchedule renders the page, sends it by email and records the result
func runPageExportSchedule(ctx context.Context, s *pageExportSchedule) (err error) {
	defer func() {
		now := time.Now()
		s.LastRunAt = &now
		s.LastError = ""
		if err != nil {
			s.LastError = err.Error()
		}
		_ = savePageExportSchedule(ctx, s)
	}()

	if len(s.Recipients) == 0 {
		return fmt.Errorf("no recipients")
	}

	identity, err := pageExportOwnerIdentity(ctx, s.OwnerID)
	if err != nil {
		return err
	}

	target, err := pageExportRewriteURL(s.URL)
	if err != nil {
		return err
	}

	body, _, _, err := pageExportBuild(ctx, identity, target, s.Format)
	if err != nil {
		return err
	}

	msg := mail.New()
	msg.SetHeader("To", s.Recipients...)
	msg.SetHeader("Subject", s.Subject)
	msg.SetBody("text/plain", s.Body)

	name := fmt.Sprintf("page-%d-%s.%s", s.PageID, time.Now().Format("2006-01-02"), s.Format)
	msg.Attach(name, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(body)
		return err
	}))

	return mail.Send(msg)
}

func pageExportOwnerIdentity(ctx context.Context, userID uint64) (auth.Identifiable, error) {
	u, err := store.LookupUserByID(ctx, service.DefaultStore, userID)
	if err != nil || u == nil || !u.Valid() {
		return nil, fmt.Errorf("schedule owner is not available")
	}

	mm, _, err := store.SearchRoleMembers(ctx, service.DefaultStore, systemTypes.RoleMemberFilter{
		Resource: fmt.Sprintf("corteza::system:user/%d", userID),
	})
	if err != nil {
		return nil, err
	}

	roles := make([]uint64, 0, len(mm))
	for _, m := range mm {
		roles = append(roles, m.RoleID)
	}

	return auth.Authenticated(userID, roles...), nil
}

// StartPageExportScheduler fires saved page export schedules; the schedules are
// re-read from the settings on every tick, so edits apply without restart.
func StartPageExportScheduler(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(pageExportScheduleTick)
		defer ticker.Stop()

		// pageID -> minute the schedule has fired at, protects from double firing within a minute
		fired := map[uint64]string{}

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				pageExportScheduleTickRun(ctx, t, fired)
			}
		}
	}()
}

func pageExportScheduleTickRun(ctx context.Context, now time.Time, fired map[uint64]string) {
	set, _, err := store.SearchSettingValues(ctx, service.DefaultStore, systemTypes.SettingsFilter{Prefix: pageExportScheduleKeyPrefix})
	if err != nil {
		return
	}

	minute := now.Format("2006-01-02 15:04")

	for _, v := range set {
		s := &pageExportSchedule{}
		if json.Unmarshal(v.Value, s) != nil || !s.due(now) || fired[s.PageID] == minute {
			continue
		}

		fired[s.PageID] = minute
		_ = runPageExportSchedule(ctx, s)
	}
}
